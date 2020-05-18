package imgresizer

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	stdLog "log"
	"net/http"
	"time"

	"github.com/h2non/filetype"
	svg "github.com/h2non/go-is-svg"
	"github.com/nfnt/resize"
	dwh_common "github.com/p2p-org/dwh/x/common"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

type ImageProcessingWorker struct {
	receiver            *dwh_common.RMQReceiver
	interpolationMethod resize.InterpolationFunction
	resolutions         []dwh_common.Resolution
	destination         string
	encoder             png.Encoder
	client              http.Client
	cfg                 *dwh_common.DwhCommonServiceConfig
}

func NewImageProcessingWorker(configFileName, configPath string) (*ImageProcessingWorker, error) {
	cfg := dwh_common.ReadCommonConfig(configFileName, configPath)
	receiver, err := dwh_common.NewRMQReceiver(cfg, cfg.ImgQueueName, cfg.ImgQueueMaxPriority, cfg.ImgQueuePrefetchCount)
	if err != nil {
		return nil, fmt.Errorf("could not create rabbitMQ receiver, error: %+v", err)
	}

	return &ImageProcessingWorker{
		resolutions:         cfg.Resolutions,
		destination:         fmt.Sprintf("%s:%d", cfg.StorageAddr, cfg.StoragePort),
		interpolationMethod: resize.InterpolationFunction(cfg.InterpolationMethod),
		encoder:             png.Encoder{CompressionLevel: png.BestCompression},
		client:              http.Client{Timeout: time.Second * 15},
		receiver:            receiver,
		cfg:                 cfg,
	}, nil
}

func (irw *ImageProcessingWorker) Closer() error {
	err := irw.receiver.Closer()
	if err != nil {
		return fmt.Errorf("could not invoke receiver closer, error: %+v", err)
	}
	return nil
}

func (irw *ImageProcessingWorker) Run() error {
	msgs, err := irw.receiver.GetMessageChan()
	if err != nil {
		return fmt.Errorf("could not get rabbitMQ msg chan, error: %+v", err)
	}

	for d := range msgs {
		err = irw.ProcessMessage(d.Body)
		if err != nil {
			stdLog.Println("failed to process rabbitMQ message: ", err)
			// TODO continue?
			// continue
		}

		err = d.Ack(false)
		if err != nil {
			stdLog.Println("failed to ack to rabbitMQ: ", err)
			// TODO continue or panic?
			// continue
		}
	}
	return nil
}

func (irw *ImageProcessingWorker) ProcessMessage(msg []byte) error {
	fmt.Println("Got message", string(msg))
	var rcvd dwh_common.TaskInfo
	err := json.Unmarshal(msg, &rcvd)
	if err != nil {
		return fmt.Errorf("img resizer unmarshal error: %+v", err)
	}

	originalImgBytes, err := irw.getImage(rcvd.URL)
	if err != nil {
		return fmt.Errorf("img resizer get image error: %+v", err)
	}
	originalImage, isVector, err := irw.decodeImage(originalImgBytes)
	if err != nil {
		return fmt.Errorf("could not decode image, error: %+v", err)
	}

	if isVector {
		// need no conversion or resize
		if err := irw.checkAndSendImage(originalImgBytes, dwh_common.Resolution{}, &rcvd); err != nil {
			return fmt.Errorf("vector image checkAndSend error: %+v", err)
		}
	} else {
		for _, r := range irw.resolutions {
			r := r
			err := irw.resizeAndSendRasterImage(originalImage, r, &rcvd)
			if err != nil {
				return fmt.Errorf("raster image resizeAndSend error: %+v", err)
			}
		}
	}
	return nil
}

func (irw *ImageProcessingWorker) getImage(imgUrl string) ([]byte, error) {
	resp, err := irw.client.Get(imgUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ba, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body, error: %+v", err)
	}

	return ba, nil
}

func (irw *ImageProcessingWorker) checkImgExistence(imgBytes []byte, resolution dwh_common.Resolution, info *dwh_common.TaskInfo) (bool, error) {
	sum := md5.Sum(imgBytes)
	req := dwh_common.ImageCheckSumRequest{
		Owner:      info.Owner,
		TokenId:    info.TokenID,
		Resolution: resolution,
		MD5Sum:     sum[:],
	}

	ba, err := json.Marshal(&req)
	if err != nil {
		return false, fmt.Errorf("could not marshal checkSum request, error: %+v", err)
	}
	dataBuf := bytes.NewReader(ba)

	resp, err := irw.client.Post(irw.destination+dwh_common.GetCheckSumPath, "application/json", dataBuf)
	if err != nil {
		return false, fmt.Errorf("could not post checkSum message, error: %+v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("check image existence error, code: %+v", resp.StatusCode)
	}

	ba, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("could not read checkSum response, error: %+v", err)
	}

	var repl dwh_common.ImageCheckSumResponse
	err = json.Unmarshal(ba, &repl)
	if err != nil {
		return false, fmt.Errorf("could not unmarshal checkSum response, error: %+v", err)
	}

	return repl.ImageExists, nil
}

func (irw *ImageProcessingWorker) resizeAndSendRasterImage(
	originalImg image.Image,
	resolution dwh_common.Resolution,
	info *dwh_common.TaskInfo,
) error {
	img := resize.Resize(resolution.Width, resolution.Height, originalImg, irw.interpolationMethod)
	buf := new(bytes.Buffer)

	if err := irw.encoder.Encode(buf, img); err != nil {
		return fmt.Errorf("encode image error: %+v", err)
	}
	return irw.checkAndSendImage(buf.Bytes(), resolution, info)
}

func (irw *ImageProcessingWorker) checkAndSendImage(
	imgBytes []byte,
	resolution dwh_common.Resolution,
	info *dwh_common.TaskInfo,
) error {
	ok, err := irw.checkImgExistence(imgBytes, resolution, info)
	if err != nil {
		// no return, work further
		stdLog.Println("checkImgExistence error:", err)
	}

	// image exists, do nothing
	if ok {
		return nil
	}

	err = irw.sendImage(imgBytes, resolution, info)
	if err != nil {
		return fmt.Errorf("could not send image, error: %+v", err)
	}

	return nil
}

func (irw *ImageProcessingWorker) decodeImage(
	originalImgBytes []byte,
) (image.Image, bool, error) {
	if svg.IsSVG(originalImgBytes) {
		return nil, true, nil
	}

	head := make([]byte, 261)
	seekReader := bytes.NewReader(originalImgBytes)
	if _, err := seekReader.Read(head); err != nil {
		return nil, false, fmt.Errorf("could not read image head, error: %+v", err)
	}

	if !filetype.IsImage(head) {
		return nil, false, fmt.Errorf("unknown image format")
	}

	t, err := filetype.Match(head)
	if err != nil {
		return nil, false, fmt.Errorf("could not guess image, error: %+v", err)
	}
	imgFormat := t.MIME.Value

	// rewind reader
	if _, err := seekReader.Seek(0, io.SeekStart); err != nil {
		return nil, false, fmt.Errorf("could not rewind reader, error: %+v", err)
	}

	var originalImg image.Image
	switch imgFormat {
	case "image/bmp":
		originalImg, err = bmp.Decode(seekReader)
	case "image/webp":
		originalImg, err = webp.Decode(seekReader)
	case "image/tiff":
		originalImg, err = tiff.Decode(seekReader)
	case "image/jpeg":
		originalImg, err = jpeg.Decode(seekReader)
	case "image/gif":
		originalImg, err = gif.Decode(seekReader)
	case "image/png":
		originalImg, err = png.Decode(seekReader)
	default:
		return nil, false, fmt.Errorf("could not decode image, error: unknown image format")
	}
	if err != nil {
		return nil, false, fmt.Errorf("could not decode image, error: %+v", err)
	}

	return originalImg, false, nil
}

func (irw *ImageProcessingWorker) sendImage(
	imgBytes []byte,
	resolution dwh_common.Resolution,
	info *dwh_common.TaskInfo,
) error {
	var gzipBuf bytes.Buffer
	zw := gzip.NewWriter(&gzipBuf)

	_, err := zw.Write(imgBytes)
	if err != nil {
		return fmt.Errorf("gzip write error: %+v", err)
	}

	if err := zw.Flush(); err != nil {
		return fmt.Errorf("gzip flush error: %+v", err)
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf("gzip close error: %+v", err)
	}

	req := dwh_common.ImageStoreRequest{
		Owner:      info.Owner,
		TokenId:    info.TokenID,
		Resolution: resolution,
		ImageBytes: gzipBuf.Bytes(),
	}

	ba, err := json.Marshal(&req)
	if err != nil {
		return fmt.Errorf("image store marshal error: %+v", err)
	}

	dataBuf := bytes.NewReader(ba)

	resp, err := irw.client.Post(irw.destination+dwh_common.StoreImagePath, "application/json", dataBuf)
	if err != nil {
		return fmt.Errorf("image store post error: %+v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error storing, status code:  %+v", resp.StatusCode)
	}

	return nil
}

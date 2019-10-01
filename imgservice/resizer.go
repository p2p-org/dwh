package imgservice

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/nfnt/resize"
)

type ImageProcessingWorker struct {
	receiver            *RMQReceiver
	interpolationMethod resize.InterpolationFunction
	resolutions         []Resolution
	destination         string
	encoder             png.Encoder
	client              http.Client
	cfg                 *DwhImgServiceConfig
}

func NewImageProcessingWorker(configFileName, configPath string) (*ImageProcessingWorker, error) {
	cfg := ReadDwhImageServiceConfig(configFileName, configPath)

	receiver, err := NewRMQReceiver(cfg)
	if err != nil {
		return nil, err
	}

	return &ImageProcessingWorker{
		resolutions:         cfg.Resolutions,
		destination:         fmt.Sprintf("%s:%d", cfg.StoreAddr, cfg.StorePort),
		interpolationMethod: resize.InterpolationFunction(cfg.InterpolationMethod),
		encoder:             png.Encoder{CompressionLevel: png.BestCompression},
		client:              http.Client{Timeout: time.Second * 15},
		receiver:            receiver,
	}, nil
}

func (irw *ImageProcessingWorker) Closer() error {
	err := irw.receiver.Closer()
	if err != nil {
		return err
	}
	return nil
}

func (irw *ImageProcessingWorker) Run() error {
	msgs, err := irw.receiver.GetMessageChan()
	if err != nil {
		return err
	}

	for d := range msgs {
		err = irw.processMessage(d.Body)
		if err != nil {
			fmt.Println("failed to process rabbitMQ message: ", err)
			continue
		}

		err = d.Ack(false)
		if err != nil {
			fmt.Println("failed to ack to rabbitMQ: ", err)
			continue
		}

	}
	return nil
}

func (irw *ImageProcessingWorker) processMessage(msg []byte) error {
	var rcvd ImageInfo
	err := json.Unmarshal(msg, &rcvd)
	if err != nil {
		return fmt.Errorf("unmarshal error: %v", err)
	}

	originalImg, err := irw.getImage(rcvd.ImgUrl)
	if err != nil {
		return fmt.Errorf("get image error: %v", err)
	}
	for _, r := range irw.resolutions {
		r := r
		err := irw.resizeAndSendImage(originalImg, r, &rcvd)
		if err != nil {
			return fmt.Errorf("resize and send image error: %v", err)
		}
	}
	return nil
}

func (irw *ImageProcessingWorker) getImage(imgUrl string) (image.Image, error) {
	resp, err := irw.client.Get(imgUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (irw *ImageProcessingWorker) checkImgExistence(imgBytes []byte, resolution Resolution, info *ImageInfo) (bool, error) {
	sum := md5.Sum(imgBytes)
	req := ImageCheckSumRequest{
		Owner:      info.Owner,
		ImgType:    info.ImgType,
		Resolution: resolution,
		MD5Sum:     sum[:],
	}

	ba, err := json.Marshal(&req)
	if err != nil {
		return false, err
	}
	dataBuf := bytes.NewReader(ba)

	resp, err := irw.client.Post(irw.destination+GetCheckSumPath, "application/json", dataBuf)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("check image existence error, code: %v", resp.StatusCode)
	}

	ba, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var repl ImageCheckSumResponse
	err = json.Unmarshal(ba, &repl)
	if err != nil {
		return false, err
	}

	return repl.ImageExists, nil
}

func (irw *ImageProcessingWorker) resizeAndSendImage(originalImg image.Image, resolution Resolution, info *ImageInfo) error {
	img := resize.Resize(resolution.Width, resolution.Height, originalImg, irw.interpolationMethod)
	buf := new(bytes.Buffer)

	if err := irw.encoder.Encode(buf, img); err != nil {
		return fmt.Errorf("encode image error: %v", err)
	}

	ok, err := irw.checkImgExistence(buf.Bytes(), resolution, info)
	if err != nil {
		fmt.Println("checkImgExistence error:", err)
	}

	// image exists, do nothing
	if ok {
		return nil
	}

	var gzipBuf bytes.Buffer
	zw := gzip.NewWriter(&gzipBuf)

	_, err = zw.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("gzip write error: %v", err)
	}

	if err := zw.Flush(); err != nil {
		return fmt.Errorf("gzip flush error: %v", err)
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf("gzip close error: %v", err)
	}

	req := ImageStoreRequest{
		Owner:      info.Owner,
		ImgType:    info.ImgType,
		Resolution: resolution,
		ImageBytes: gzipBuf.Bytes(),
	}

	ba, err := json.Marshal(&req)
	if err != nil {
		return fmt.Errorf("image store marshal error: %v", err)
	}

	dataBuf := bytes.NewReader(ba)

	resp, err := irw.client.Post(irw.destination+StoreImagePath, "application/json", dataBuf)
	if err != nil {
		return fmt.Errorf("image store post error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error storing, status code:  %v", resp.StatusCode)
	}

	return nil
}

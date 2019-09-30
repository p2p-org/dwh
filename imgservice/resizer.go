package imgservice

import (
	"bytes"
	"compress/gzip"
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
		destination:         cfg.StoreAddr,
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
		return err
	}

	originalImg, err := irw.getImage(rcvd.ImgUrl)
	if err != nil {
		return err
	}
	for _, r := range irw.resolutions {
		r := r
		err := irw.resizeAndSendImage(originalImg, r, rcvd.Owner)
		if err != nil {
			return err
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

func (irw *ImageProcessingWorker) resizeAndSendImage(originalImg image.Image, resolution Resolution, owner string) error {
	img := resize.Resize(uint(resolution.Width), uint(resolution.Height), originalImg, irw.interpolationMethod)
	buf := new(bytes.Buffer)

	if err := irw.encoder.Encode(buf, img); err != nil {
		return err
	}

	var gzipBuf bytes.Buffer
	zw := gzip.NewWriter(&gzipBuf)

	_, err := zw.Write(buf.Bytes())
	if err != nil {
		return err
	}

	if err := zw.Flush(); err != nil {
		return err
	}

	if err := zw.Close(); err != nil {
		return err
	}

	req := ImagePostRequest{
		Owner:      owner,
		Resolution: resolution,
		ImageBytes: gzipBuf.Bytes(),
	}

	ba, err := json.Marshal(&req)
	if err != nil {
		return err
	}
	dataBuf := bytes.NewReader(ba)

	resp, err := irw.client.Post(irw.destination, "application/json", dataBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ba, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var repl ImagePostResponse
	err = json.Unmarshal(ba, &repl)
	if err != nil {
		return err
	}

	if repl.Code != 200 {
		return fmt.Errorf("error storing  %v", repl.Error)
	}

	return nil
}

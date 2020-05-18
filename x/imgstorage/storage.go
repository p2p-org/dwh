package imgstorage

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	dwh_common "github.com/p2p-org/dwh/x/common"
)

func (ims *ImgStorage) storeImg(req *dwh_common.ImageStoreRequest) error {
	dirPath := path.Join(ims.storagePath, req.Owner)
	inf, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0777)
		if err != nil {
			return fmt.Errorf("could not mkDir, error: %+v", err)
		}
	}

	if err == nil && !inf.IsDir() {
		err := os.Remove(dirPath)
		if err != nil {
			return fmt.Errorf("could not rm file, error: %+v", err)
		}
		err = os.MkdirAll(dirPath, 0777)
		if err != nil {
			return fmt.Errorf("could not mkDir instead of file, error: %+v", err)
		}
	}

	name := fmt.Sprintf(FileNameFormat, req.Owner, req.TokenId, req.Resolution.Width, req.Resolution.Height)
	filePrefix := fmt.Sprintf("%x", md5.Sum([]byte(name)))
	names, err := filepath.Glob(path.Join(dirPath, filePrefix) + "*")
	if err != nil {
		return fmt.Errorf("could not search in directory, error: %+v", err)
	}

	buf := bytes.NewBuffer(req.ImageBytes)

	zr, err := gzip.NewReader(buf)
	if err != nil {
		return fmt.Errorf("could not create gzip reader, error: %+v", err)
	}

	ba, err := ioutil.ReadAll(zr)
	if err != nil {
		return fmt.Errorf("could not read from gzip, error: %+v", err)
	}

	err = zr.Close()
	if err != nil {
		return fmt.Errorf("could not close gzip reader, error: %+v", err)
	}

	fileName := fmt.Sprintf("%s+%x", filePrefix, md5.Sum(ba))
	fileFullName := path.Join(dirPath, fileName)

	var bytesToStore []byte
	if ims.optionStoreCompressed {
		bytesToStore = req.ImageBytes
	} else {
		bytesToStore = ba
	}

	err = ioutil.WriteFile(fileFullName, bytesToStore, 0777)
	if err != nil {
		return fmt.Errorf("could not write file, error: %+v", err)
	}

	for _, v := range names {
		err := os.Remove(v)
		if err != nil {
			return fmt.Errorf("could not remove old file, error: %+v", err)
		}
	}

	return nil
}

func (ims *ImgStorage) loadImg(owner, imgType string, width, height int) (string, error) {
	dirPath := path.Join(ims.storagePath, owner)
	inf, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("dir not found error: %v", err)
	}
	if !inf.IsDir() {
		return "", fmt.Errorf("error: dir not found")
	}

	name := fmt.Sprintf(FileNameFormat, owner, imgType, width, height)
	filePrefix := fmt.Sprintf("%x+", md5.Sum([]byte(name)))

	names, err := filepath.Glob(path.Join(dirPath, filePrefix) + "*")
	if err != nil {
		return "", fmt.Errorf("glob error: %v", err)
	}

	if len(names) == 0 {
		name = fmt.Sprintf(FileNameFormat, owner, imgType, 0, 0)
		filePrefix = fmt.Sprintf("%x+", md5.Sum([]byte(name)))
		names, err = filepath.Glob(path.Join(dirPath, filePrefix) + "*")
		if err != nil {
			return "", fmt.Errorf("glob error: %v", err)
		}

		if len(names) == 0 {
			return "", fmt.Errorf("no files found")
		}
	}

	fullFileName := names[0]

	_, err = os.Stat(fullFileName)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("load img os stat error: %v", err)
	}

	return fullFileName, nil
}

func (ims *ImgStorage) getCheckFile(req *dwh_common.ImageCheckSumRequest) bool {
	dirPath := path.Join(ims.storagePath, req.Owner)
	inf, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false
	}

	if err == nil && !inf.IsDir() {
		return false
	}

	name := fmt.Sprintf(FileNameFormat, req.Owner, req.TokenId, req.Resolution.Width, req.Resolution.Height)
	filename := fmt.Sprintf("%x+%x", md5.Sum([]byte(name)), req.MD5Sum)
	fileFullName := path.Join(ims.storagePath, req.Owner, filename)

	_, err = os.Stat(fileFullName)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

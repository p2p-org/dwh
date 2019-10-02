package imgservice

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

const FileNameFormat = "%s_%s_%d_%d"

type ImgStore struct {
	optionStoreCompressed bool
	storagePath           string
}

func NewImgStore(storagePath string, storeCompressed bool) *ImgStore {
	return &ImgStore{
		optionStoreCompressed: storeCompressed,
		storagePath:           storagePath,
	}
}

func (ims *ImgStore) StoreHandler(w http.ResponseWriter, r *http.Request) {
	reqB, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req ImageStoreRequest
	err = json.Unmarshal(reqB, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = ims.storeImg(&req)
	if err != nil {
		fmt.Println("store image error")
		return
	}
}

func (ims *ImgStore) storeImg(req *ImageStoreRequest) error {
	dirPath := path.Join(ims.storagePath, req.Owner)
	inf, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0777)
		if err != nil {
			return err
		}
	}

	if err == nil && !inf.IsDir() {
		err := os.Remove(dirPath)
		if err != nil {
			return err
		}
		err = os.MkdirAll(dirPath, 0777)
		if err != nil {
			return err
		}
	}

	name := fmt.Sprintf(FileNameFormat, req.Owner, req.TokenId, req.Resolution.Width, req.Resolution.Height)
	filePrefix := fmt.Sprintf("%x", md5.Sum([]byte(name)))
	names, err := filepath.Glob(path.Join(dirPath, filePrefix) + "*")
	if err != nil {
		return nil
	}

	buf := bytes.NewBuffer(req.ImageBytes)

	zr, err := gzip.NewReader(buf)
	if err != nil {
		return err
	}

	ba, err := ioutil.ReadAll(zr)
	if err != nil {
		return err
	}

	err = zr.Close()
	if err != nil {
		return err
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
		return err
	}

	for _, v := range names {
		err := os.Remove(v)
		if err != nil {
			fmt.Println("error removing file:", err)
		}
	}

	return nil
}

func (ims *ImgStore) LoadHandler(w http.ResponseWriter, r *http.Request) {
	owner := r.FormValue("owner")
	imgType := r.FormValue("img_type")
	widthString := r.FormValue("width")
	heightString := r.FormValue("height")
	width, err := strconv.Atoi(widthString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong width"))
		return
	}
	height, err := strconv.Atoi(heightString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong height"))
		return
	}

	ba, err := ims.loadImg(owner, imgType, width, height)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("img not found"))
		return
	}

	var bytesToLoad []byte
	if ims.optionStoreCompressed {
		buf := bytes.NewBuffer(ba)
		zr, err := gzip.NewReader(buf)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bytesToLoad, err = ioutil.ReadAll(zr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := zr.Close(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		bytesToLoad = ba
	}

	_, err = w.Write(bytesToLoad)
	if err != nil {
		fmt.Println("load img write response error:", err)
		return
	}
}

func (ims *ImgStore) loadImg(owner, imgType string, width, height int) ([]byte, error) {
	dirPath := path.Join(ims.storagePath, owner)
	inf, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0777)
		if err != nil {
			return nil, fmt.Errorf("dir not found error: %v", err)
		}
	}
	if !inf.IsDir() {
		return nil, fmt.Errorf("dir not found")
	}

	name := fmt.Sprintf(FileNameFormat, owner, imgType, width, height)
	filePrefix := fmt.Sprintf("%x+", md5.Sum([]byte(name)))

	names, err := filepath.Glob(path.Join(dirPath, filePrefix) + "*")
	if err != nil {
		return nil, fmt.Errorf("glob error: %v", err)
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no files found")
	}

	fullFileName := names[0]

	_, err = os.Stat(fullFileName)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("load img os stat error: %v", err)
	}

	ba, err := ioutil.ReadFile(fullFileName)
	if err != nil {
		return nil, fmt.Errorf("load img read file error: %v", err)
	}

	return ba, nil
}

func (ims *ImgStore) GetCheckSumHandler(w http.ResponseWriter, r *http.Request) {
	reqB, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req ImageCheckSumRequest
	err = json.Unmarshal(reqB, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ok := ims.getCheckFile(&req)

	ba, err := json.Marshal(&ImageCheckSumResponse{
		ImageExists: ok,
	})
	if err != nil {
		fmt.Println("check file marshal response error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(ba)
	if err != nil {
		fmt.Println("check sum write response error:", err)
		return
	}
}

func (ims *ImgStore) getCheckFile(req *ImageCheckSumRequest) bool {
	dirPath := path.Join(ims.storagePath, req.Owner)
	inf, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0777)
		if err != nil {
			return false
		}
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

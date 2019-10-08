package imgstore

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	dwh_common "github.com/dgamingfoundation/dwh/x/common"
)

func (ims *ImgStore) StoreHandler(w http.ResponseWriter, r *http.Request) {
	reqB, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req dwh_common.ImageStoreRequest
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

func (ims *ImgStore) GetCheckSumHandler(w http.ResponseWriter, r *http.Request) {
	reqB, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req dwh_common.ImageCheckSumRequest
	err = json.Unmarshal(reqB, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ok := ims.getCheckFile(&req)

	ba, err := json.Marshal(&dwh_common.ImageCheckSumResponse{
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

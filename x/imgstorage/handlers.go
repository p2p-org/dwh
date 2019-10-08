package imgstorage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	dwh_common "github.com/dgamingfoundation/dwh/x/common"
)

func (ims *ImgStorage) StoreHandler(w http.ResponseWriter, r *http.Request) {
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

func (ims *ImgStorage) LoadHandler(w http.ResponseWriter, r *http.Request) {
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

	fileName, err := ims.loadImg(owner, imgType, width, height)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("img not found"))
		return
	}

	http.ServeFile(w, r, fileName)
}

func (ims *ImgStorage) GetCheckSumHandler(w http.ResponseWriter, r *http.Request) {
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

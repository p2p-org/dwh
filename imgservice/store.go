package imgservice

import "net/http"

type ImgStore struct {
}

func NewImgStore() *ImgStore {

}

func (ims *ImgStore) StoreHandler(w http.ResponseWriter, r *http.Request) {

}

func (ims *ImgStore) LoadHandler(w http.ResponseWriter, r *http.Request) {

}

func (ims *ImgStore) GetCheckSumHandler(w http.ResponseWriter, r *http.Request) {

}

func (ims *ImgStore) storeImg(owner, name string, resolution Resolution) error {
}

func (ims *ImgStore) loadImg(owner, name string, resolution Resolution) ([]byte, error) {
}

func (ims *ImgStore) getCheckSum(owner, name string, resolution Resolution) (string, error) {

}

func (ims *ImgStore) writeImgInfo(info *StoredImgInfo) error {

}

type StoredImgInfo struct {
	Resolution `json:"resolution"`
	Owner      string `json:"owner"`
	Name       string `json:"name"`
	MD5        string `json:"md5"`
}

type ImgInfoFile map[Resolution]StoredImgInfo

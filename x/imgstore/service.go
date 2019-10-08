package imgstore

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

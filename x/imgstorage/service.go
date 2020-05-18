package imgstorage

import (
	dwh_common "github.com/p2p-org/dwh/x/common"
)

const FileNameFormat = "%s_%s_%d_%d"

type ImgStorage struct {
	cfg                   *dwh_common.DwhCommonServiceConfig
	optionStoreCompressed bool
	storagePath           string
}

func NewImgStorage(cfg *dwh_common.DwhCommonServiceConfig) *ImgStorage {
	return &ImgStorage{
		cfg:                   cfg,
		optionStoreCompressed: cfg.StorageCompressedOption,
		storagePath:           cfg.StorageDiskPath,
	}
}

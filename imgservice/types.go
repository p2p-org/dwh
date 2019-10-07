package imgservice

type ImgQueuePriority uint8

const (
	RegularUpdatePriority ImgQueuePriority = 1 + iota
	TransferTriggeredPriority
	FreshlyMadePriority
	ForcedUpdatesPriority
)

const (
	ImgTypeAvatar = "avatar"
)

const (
	StoreImagePath  = "/imgstore/store_img"
	LoadImagePath   = "/imgstore/load_img"
	GetCheckSumPath = "/imgstore/get_check_sum"
)

type ImageInfo struct {
	Owner   string `json:"owner"`
	TokenId string `json:"token_id"`
	ImgUrl  string `json:"img_url"`
}

type ImageStoreRequest struct {
	Owner      string `json:"owner"`
	TokenId    string `json:"token_id"`
	Resolution `json:"resolution"`
	ImageBytes []byte `json:"image_bytes"` // compressed
}

type ImageCheckSumRequest struct {
	Owner      string `json:"owner"`
	TokenId    string `json:"token_id"`
	Resolution `json:"resolution"`
	MD5Sum     []byte `json:"md5_sum"`
}

type ImageCheckSumResponse struct {
	ImageExists bool `json:"image_exists,omitempty"`
}

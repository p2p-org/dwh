package imgservice

type ImgQueuePriority uint8

const (
	RegularUpdatePriority ImgQueuePriority = 1 + iota
	TransferTriggeredPriority
	FreshlyMadePriority
	ForcedUpdatesPriority
)

type ImageInfo struct {
	Owner  string `json:"owner"`
	ImgUrl string `json:"img_url"`
}

type ImagePostRequest struct {
	Owner      string `json:"owner"`
	Resolution `json:"resolution"`
	ImageBytes []byte `json:"image_bytes"`
}

type ImagePostResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error,omitempty"`
}

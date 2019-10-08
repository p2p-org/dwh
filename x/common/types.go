package dwh_common

type ImgQueuePriority uint8

const (
	RegularUpdatePriority ImgQueuePriority = 1 + iota
	TransferTriggeredPriority
	FreshlyMadePriority
	ForcedUpdatesPriority
)

type Resolution struct {
	Width  uint `mapstructure:"width",json:"width"`
	Height uint `mapstructure:"height",json:"height"`
}

type TaskInfo struct {
	Owner   string `json:"owner"`
	TokenID string `json:"token_id"`
	URL     string `json:"url"`
}

package mongoDaemon

type ImgQueuePriority uint8

const (
	RegularUpdatePriority ImgQueuePriority = 1 + iota
	TransferTriggeredPriority
	FreshlyMadePriority
	ForcedUpdatesPriority
)

type TokenInfo struct {
	TokenID string `json:"token_id"`
	URL     string `json:"url"`
	Owner   string `json:"owner"`
}

var MongoTaskInf = make([]byte, 0)

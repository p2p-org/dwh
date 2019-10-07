package tokenMetadataService

type TokenInfo struct {
	TokenID string `json:"token_id"`
	URL     string `json:"url"`
	Owner   string `json:"owner"`
}

type URIQueuePriority uint8

const (
	RegularUpdatePriority URIQueuePriority = 1 + iota
	TransferTriggeredPriority
	FreshlyMadePriority
	ForcedUpdatesPriority
)

const erc721Schema = `
{
    "title": "Asset Metadata",
    "type": "object",
    "properties": {
        "name": {
            "type": "string",
            "description": "Identifies the asset to which this NFT represents"
        },
        "description": {
            "type": "string",
            "description": "Describes the asset to which this NFT represents"
        },
        "image": {
            "type": "string",
            "description": "A URI pointing to a resource with mime type image/* representing the asset to which this NFT represents. Consider making any images at a width between 320 and 1080 pixels and aspect ratio between 1.91:1 and 4:5 inclusive."
        }
    }
}`

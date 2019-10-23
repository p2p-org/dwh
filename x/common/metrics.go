package dwh_common

import (
	"github.com/prometheus/client_golang/prometheus"
)

type MsgMetrics struct {
	NumMsgs *prometheus.CounterVec
}

const (
	PrometheusLabelStatus                    = "status"
	PrometheusLabelMsgType                   = "msg_type"
	PrometheusValueReceived                  = "Received"
	PrometheusValueAccepted                  = "Accepted"
	PrometheusValueCommon                    = "Common"
	PrometheusValueMsgMintNFT                = "MsgMintNFT"
	PrometheusValueMsgEditNFTMetadata        = "MsgEditNFTMetadata"
	PrometheusValueMsgPutNFTOnMarket         = "MsgPutNFTOnMarket"
	PrometheusValueMsgRemoveNFTFromMarket    = "MsgRemoveNFTFromMarket"
	PrometheusValueMsgBuyNFT                 = "MsgBuyNFT"
	PrometheusValueMsgTransferNFT            = "MsgTransferNFT"
	PrometheusValueMsgCreateFungibleToken    = "MsgCreateFungibleToken"
	PrometheusValueMsgTransferFungibleTokens = "MsgTransferFungibleTokens"
	PrometheusValueMsgMakeOffer              = "MsgMakeOffer"
	PrometheusValueMsgAcceptOffer            = "MsgAcceptOffer"
	PrometheusValueMsgRemoveOffer            = "MsgRemoveOffer"
	PrometheusValueMsgPutNFTOnAuction        = "MsgPutNFTOnAuction"
	PrometheusValueMsgRemoveFromAuction      = "MsgRemoveFromAuction"
	PrometheusValueMsgMakeBidOnAuction       = "MsgMakeBidOnAuction"
	PrometheusValueMsgBuyoutOnAuction        = "MsgBuyoutOnAuction"
	PrometheusValueMsgFinishAuction          = "MsgFinishAuction"
)

func NewPrometheusMsgMetrics(module string) *MsgMetrics {
	numMsgs := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "DWH",
		Subsystem: module + "_MetricsSubsystem",
		Name:      "NumMsgs",
		Help:      "number of messages since start",
	},
		[]string{PrometheusLabelStatus, PrometheusLabelMsgType},
	)
	prometheus.MustRegister(numMsgs)
	return &MsgMetrics{
		NumMsgs: numMsgs,
	}
}

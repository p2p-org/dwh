package common

import (
	"github.com/prometheus/client_golang/prometheus"
)

type MsgMetrics struct {
	NumMsgs         prometheus.Counter
	NumMsgsAccepted prometheus.Counter
}

func NewPrometheusMsgMetrics(module string) *MsgMetrics {
	numMsgs := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "DWH",
		Subsystem: module + "_MetricsSubsystem",
		Name:      "NumMsgs",
		Help:      "number of messages since start",
	})
	msgsAccepted := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "DWH",
		Subsystem: module + "_MetricsSubsystem",
		Name:      "NumMsgsAccepted",
		Help:      "number of messages handled without error",
	})
	prometheus.MustRegister(numMsgs)
	prometheus.MustRegister(msgsAccepted)
	return &MsgMetrics{
		NumMsgs:         numMsgs,
		NumMsgsAccepted: msgsAccepted,
	}
}

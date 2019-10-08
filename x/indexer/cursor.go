package indexer

import "encoding/json"

type cursor struct {
	Height  int64
	TxIndex uint32
	MsgID   int
}

func (m *cursor) Marshal() []byte {
	bz, _ := json.Marshal(m)
	return bz
}

func (m *cursor) Unmarshal(bz []byte) error {
	return json.Unmarshal(bz, m)
}

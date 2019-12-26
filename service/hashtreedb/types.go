package hashtreedb

type StoreItem interface {
	Parse([]byte, uint32) error
	Serialize() []byte

	Size() uint32 // 数据长度
}

///////////////////////////////////////////////

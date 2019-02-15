package block

type Block interface {

	// 序列化 与 反序列化
	Serialize() ([]byte, error)
	Parse([]byte, uint32) (uint32, error)

	SerializeHead() ([]byte, error)
	ParseHead([]byte, uint32) (uint32, error)
}

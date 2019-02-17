package block

type Block interface {

	// 序列化 与 反序列化
	Serialize() ([]byte, error)
	Parse([]byte, uint32) (uint32, error)
	Size() uint32

	SerializeHead() ([]byte, error)
	ParseHead([]byte, uint32) (uint32, error)

	SerializeBody() ([]byte, error)
	ParseBody([]byte, uint32) (uint32, error)
}

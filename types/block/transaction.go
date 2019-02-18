package block

type Transaction interface {

	// 交易类型
	Type() uint8

	// 序列化 与 反序列化
	Serialize() ([]byte, error)
	Parse([]byte, uint32) (uint32, error)
	Size() uint32

	// 交易唯一哈希值
	Hash() []byte
}

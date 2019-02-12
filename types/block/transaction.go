package block

type Transaction interface {

	// 交易类型
	Type() uint8

	// 序列化 与 反序列化
	Serialize() ([]byte, error)
	Parse(*[]byte, int) (int, error)
}

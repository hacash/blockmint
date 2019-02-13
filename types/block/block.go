package block

type Block interface {

	// 序列化 与 反序列化
	Serialize() ([]byte, error)
	Parse(*[]byte, int) (int, error)
}

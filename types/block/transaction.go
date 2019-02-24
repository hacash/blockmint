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
	HashNoFee() []byte // 无手续费的哈希

	// 从 actions 拿出需要签名的地址
	RequestSignAddrs() ([][]byte, error)
	// 填充签名
	FillNeedSigns(map[string][]byte) error
	// 验证需要的签名
	VerifyNeedSigns() (bool, error)

	// 其他
	FeePurity() uint64 // 手续费含量

}

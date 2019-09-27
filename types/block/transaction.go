package block

import . "github.com/hacash/blockmint/types/state"

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
	RequestSignAddrs([][]byte) ([][]byte, error)
	// 填充签名
	FillNeedSigns(map[string][]byte, [][]byte) error
	// 验证需要的签名，包含参数时则检查指定地址的签名
	VerifyNeedSigns([][]byte) (bool, error)

	// 修改 / 恢复 状态数据库
	ChangeChainState(ChainStateOperation) error
	RecoverChainState(ChainStateOperation) error

	// 其他
	FeePurity() uint64 // 手续费含量

	// 查询
	GetAddress() []byte
	GetFee() []byte
	GetActions() []Action
	GetTimestamp() uint64 // 时间戳
}

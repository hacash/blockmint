package block

import . "github.com/hacash/blockmint/types/state"

type Action interface {

	// 种类id
	Kind() uint16

	// 序列化 与 反序列化
	Size() uint32
	Serialize() ([]byte, error)
	Parse([]byte, uint32) (uint32, error)

	// 请求签名地址
	RequestSignAddrs() [][]byte

	// 修改 / 恢复 状态数据库
	ChangeChainState(ChainStateOperation) error
	RecoverChainState(ChainStateOperation) error

	// 设置所属 trs
	SetBelongTrs(Transaction)
}

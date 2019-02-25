package block

import . "github.com/hacash/blockmint/types/state"

type Action interface {

	// 种类id
	Kind() uint16

	// 序列化 与 反序列化
	Serialize() ([]byte, error)
	Parse([]byte, uint32) (uint32, error)
	Size() uint32

	// 请求签名地址
	RequestSignAddrs() [][]byte

	// 修改 / 恢复 状态数据库
	ChangeChainState(ChainStateOperation) error
	RecoverChainState(ChainStateOperation) error

	//

}

package block

import (
	"github.com/hacash/blockmint/types/store"
)

type Action interface {

	// 种类id
	Kind() uint16

	// 序列化 与 反序列化
	Serialize() ([]byte, error)
	Parse([]byte, uint32) (uint32, error)

	// 请求签名地址
	SignatureRequestAddress() []string

	// 修改余额数据库状态
	ChangeChainState(*store.ChainStateDB)

	//

}

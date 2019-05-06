package actions

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/state"
	"github.com/hacash/x16rs"
	"strings"
)

/**
 * 钻石交易类型
 */

// 挖出钻石
type Action_3_DiamondCreate struct {
	Diamond  fields.Bytes6  // 钻石字面量 WTYUIAHXVMEKBSZN
	PrevHash fields.Bytes32 // 上一个包含钻石的区块hash
	Nonce    fields.Bytes8  // 随机数
	Address  fields.Address // 所属账户

	// 数据指针
	// 所属交易
	// trs block.Transaction
}

func (elm *Action_3_DiamondCreate) Kind() uint16 {
	return 3
}

func (elm *Action_3_DiamondCreate) SetBelongTrs(t block.Transaction) {

}

func (elm *Action_3_DiamondCreate) Size() uint32 {
	return elm.Diamond.Size() + elm.Address.Size() + elm.Nonce.Size()
}

func (elm *Action_3_DiamondCreate) Serialize() ([]byte, error) {
	var kindByte = make([]byte, 2)
	binary.BigEndian.PutUint16(kindByte, elm.Kind())
	var diamondBytes, _ = elm.Diamond.Serialize()
	var prevBytes, _ = elm.PrevHash.Serialize()
	var nonceBytes, _ = elm.Nonce.Serialize()
	var addrBytes, _ = elm.Address.Serialize()
	var buffer bytes.Buffer
	buffer.Write(kindByte)
	buffer.Write(diamondBytes)
	buffer.Write(prevBytes)
	buffer.Write(nonceBytes)
	buffer.Write(addrBytes)
	return buffer.Bytes(), nil
}

func (elm *Action_3_DiamondCreate) Parse(buf []byte, seek uint32) (uint32, error) {
	var moveseek1, _ = elm.Diamond.Parse(buf, seek)
	var moveseek2, _ = elm.PrevHash.Parse(buf, moveseek1)
	var moveseek3, _ = elm.Nonce.Parse(buf, moveseek2)
	var moveseek4, _ = elm.Address.Parse(buf, moveseek3)
	return moveseek4, nil
}

func (elm *Action_3_DiamondCreate) RequestSignAddrs() [][]byte {
	return make([][]byte, 0) // 无需签名
}

func (act *Action_3_DiamondCreate) ChangeChainState(state state.ChainStateOperation) error {
	blk := state.Block().(block.Block) // 强制类型转换
	if blk == nil {
		panic("Action get state.Block() cannot be nil !")
	}
	miner := state.Miner()
	if miner == nil {
		panic("Action get state.Miner() cannot be nil !")
	}
	blkhei := blk.GetHeight()
	// 检查钻石挖矿计算
	diamond_str := x16r.Diamond(act.PrevHash, act.Nonce, act.Address)
	diamondstrval, isdia := x16r.IsDiamondHashResultString(diamond_str)
	if !isdia {
		return fmt.Errorf("String <%s> is not diamond.", diamond_str)
	}
	if strings.Compare(diamondstrval, string(act.Diamond)) != 0 {
		return fmt.Errorf("Diamond need <%s> but got <%s>", act.Diamond, diamondstrval)
	}
	// 检查区块高度值是否为5的倍数
	// {BACKTOPOOL} 表示扔回交易池等待再次处理
	if blkhei%5 != 0 {
		return fmt.Errorf("{BACKTOPOOL} Diamond must be in block height multiple of 5.")
	}
	// 检查一个区块只能包含一枚钻石
	blkhash := blk.HashFresh()
	minerprevhash := miner.GetPrevDiamondHash()
	if bytes.Compare(blkhash, minerprevhash) == 0 {
		return fmt.Errorf("This block height:%d has already exist diamond.", blkhei)
	}
	// 检查钻石是否是从上一个区块得来
	if bytes.Compare(act.PrevHash, minerprevhash) != 0 {
		return fmt.Errorf("Diamond prev hash must be <%s> but got <%s>.", hex.EncodeToString(minerprevhash), hex.EncodeToString(act.PrevHash))
	}
	// 查询钻石是否已经存在
	hasaddr := state.Diamond(act.Diamond)
	if hasaddr != nil {
		return fmt.Errorf("Diamond <%s> already exist.", string(act.Diamond))
	}
	// 存入钻石
	state.DiamondSet(act.Diamond, act.Address)
	// 设置矿工状态
	state.SetPrevDiamondHash(blkhash)
	return nil
}

func (act *Action_3_DiamondCreate) RecoverChainState(state state.ChainStateOperation) error {
	miner := state.Miner()
	if miner == nil {
		panic("Action get state.Miner() cannot be nil !")
	}
	// 删除钻石
	state.DiamondDel(act.Diamond)
	// 回退矿工状态
	state.SetPrevDiamondHash(act.PrevHash)
	return nil

}

///////////////////////////////////////////////////////////////

// 转移钻石
type Action_4_DiamondTransfer struct {
	Diamond fields.Bytes6  // 钻石字面量 WTYUIAHXVMEKBSZN
	Address fields.Address // 收钻方账户

	// 数据指针
	// 所属交易
	trs block.Transaction
}

func (elm *Action_4_DiamondTransfer) Kind() uint16 {
	return 3
}

func (elm *Action_4_DiamondTransfer) SetBelongTrs(t block.Transaction) {
	elm.trs = t
}

func (elm *Action_4_DiamondTransfer) Size() uint32 {
	return elm.Diamond.Size() + elm.Address.Size()
}

func (elm *Action_4_DiamondTransfer) Serialize() ([]byte, error) {
	var kindByte = make([]byte, 2)
	binary.BigEndian.PutUint16(kindByte, elm.Kind())
	var diamondBytes, _ = elm.Diamond.Serialize()
	var addrBytes, _ = elm.Address.Serialize()
	var buffer bytes.Buffer
	buffer.Write(kindByte)
	buffer.Write(diamondBytes)
	buffer.Write(addrBytes)
	return buffer.Bytes(), nil
}

func (elm *Action_4_DiamondTransfer) Parse(buf []byte, seek uint32) (uint32, error) {
	var moveseek1, _ = elm.Diamond.Parse(buf, seek)
	var moveseek2, _ = elm.Address.Parse(buf, moveseek1)
	return moveseek2, nil
}

func (elm *Action_4_DiamondTransfer) RequestSignAddrs() [][]byte {
	return make([][]byte, 0) // 无需签名
}

func (act *Action_4_DiamondTransfer) ChangeChainState(state state.ChainStateOperation) error {
	if act.trs == nil {
		panic("Action belong to transaction not be nil !")
	}
	// 查询钻石是否已经存在
	hasaddr := state.Diamond(act.Diamond)
	if hasaddr == nil {
		return fmt.Errorf("Diamond <%s> not exist.", string(act.Diamond))
	}
	// 检查所属
	if bytes.Compare(hasaddr, act.trs.GetAddress()) != 0 {
		return fmt.Errorf("Diamond <%s> not belong to trs address.", string(act.Diamond))
	}
	// 转移钻石
	state.DiamondSet(act.Diamond, act.Address)
	return nil
}

func (act *Action_4_DiamondTransfer) RecoverChainState(state state.ChainStateOperation) error {
	if act.trs == nil {
		panic("Action belong to transaction not be nil !")
	}
	// 返回钻石
	state.DiamondSet(act.Diamond, act.trs.GetAddress())
	return nil
}
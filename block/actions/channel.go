package actions

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/chain/state/db"
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/state"
)

/**
 * 支付通道交易类型
 */

// 开启支付通道
type Action_2_OpenPaymentChannel struct {
	ChannelId    fields.Bytes16 // 通道id
	LeftAddress  fields.Address // 账户1
	LeftAmount   fields.Amount  // 锁定金额
	RightAddress fields.Address // 账户2
	RightAmount  fields.Amount  // 锁定金额

	// 数据指针
	// 所属交易
	//trs block.Transaction
}

func (elm *Action_2_OpenPaymentChannel) Kind() uint16 {
	return 2
}

func (elm *Action_2_OpenPaymentChannel) SetBelongTrs(t block.Transaction) {

}

func (elm *Action_2_OpenPaymentChannel) Size() uint32 {
	return 2 + elm.ChannelId.Size() + ((elm.LeftAddress.Size() + elm.LeftAddress.Size()) * 2)
}

func (elm *Action_2_OpenPaymentChannel) Serialize() ([]byte, error) {
	var kindByte = make([]byte, 2)
	binary.BigEndian.PutUint16(kindByte, elm.Kind())
	var idBytes, _ = elm.ChannelId.Serialize()
	var addr1Bytes, _ = elm.LeftAddress.Serialize()
	var amt1Bytes, _ = elm.LeftAmount.Serialize()
	var addr2Bytes, _ = elm.RightAddress.Serialize()
	var amt2Bytes, _ = elm.RightAmount.Serialize()
	var buffer bytes.Buffer
	buffer.Write(kindByte)
	buffer.Write(idBytes)
	buffer.Write(addr1Bytes)
	buffer.Write(amt1Bytes)
	buffer.Write(addr2Bytes)
	buffer.Write(amt2Bytes)
	return buffer.Bytes(), nil
}

func (elm *Action_2_OpenPaymentChannel) Parse(buf []byte, seek uint32) (uint32, error) {
	seek, _ = elm.ChannelId.Parse(buf, seek)
	seek, _ = elm.LeftAddress.Parse(buf, seek)
	seek, _ = elm.LeftAmount.Parse(buf, seek)
	seek, _ = elm.RightAddress.Parse(buf, seek)
	seek, _ = elm.RightAmount.Parse(buf, seek)
	return seek, nil
}

func (elm *Action_2_OpenPaymentChannel) RequestSignAddrs() [][]byte {
	reqs := make([][]byte, 2)
	reqs[0] = elm.LeftAddress
	reqs[1] = elm.RightAddress
	return reqs
}

func (act *Action_2_OpenPaymentChannel) ChangeChainState(state state.ChainStateOperation) error {
	// 查询通道是否存在
	sto := state.Channel(act.ChannelId)
	if sto != nil {
		return fmt.Errorf("Payment Channel Id <%s> already exist.", hex.EncodeToString(act.ChannelId))
	}
	// 检查金额储存的位数
	labt, _ := act.LeftAmount.Serialize()
	rabt, _ := act.RightAmount.Serialize()
	if len(labt) > 6 || len(rabt) > 6 {
		return fmt.Errorf("Payment Channel create error: left or right Amount bytes too long.")
	}
	// 不能为零或负数
	if !act.LeftAmount.IsPositive() || !act.RightAmount.IsPositive() {
		return fmt.Errorf("Payment Channel create error: left or right Amount is not positive.")
	}
	// 检查余额是否充足
	amt1 := state.Balance(act.LeftAddress)
	if amt1.LessThan(&act.LeftAmount) {
		return fmt.Errorf("Address %s Balance is not enough.", act.LeftAddress.ToReadable())
	}
	amt2 := state.Balance(act.RightAddress)
	if amt2.LessThan(&act.RightAmount) {
		return fmt.Errorf("Address %s Balance is not enough.", act.RightAddress.ToReadable())
	}
	curheight := uint64(1)
	curblkptr := state.Block()
	if curblkptr != nil {
		curblk := curblkptr.(block.Block)
		curheight = curblk.GetHeight()
	}
	// 创建 channel
	var storeItem db.ChannelStoreItemData
	storeItem.BelongHeight = fields.VarInt5(curheight)
	storeItem.LockBlock = fields.VarInt2(uint16(5000)) // 单方面提出的锁定期约为 17 天
	storeItem.LeftAddress = act.LeftAddress
	storeItem.LeftAmount = act.LeftAmount
	storeItem.RightAddress = act.RightAddress
	storeItem.RightAmount = act.RightAmount
	storeItem.IsClosed = 0 // 打开状态
	// 扣除余额
	DoSubBalanceFromChainState(state, act.LeftAddress, act.LeftAmount)
	DoSubBalanceFromChainState(state, act.RightAddress, act.RightAmount)
	// 储存通道
	state.ChannelCreate(act.ChannelId, &storeItem)
	//
	return nil
}

func (act *Action_2_OpenPaymentChannel) RecoverChainState(state state.ChainStateOperation) error {
	// 删除通道
	state.ChannelDelete(act.ChannelId)
	// 恢复余额
	DoAddBalanceFromChainState(state, act.LeftAddress, act.LeftAmount)
	DoAddBalanceFromChainState(state, act.RightAddress, act.RightAmount)
	return nil
}

/////////////////////////////////////////////////////////////////

// 关闭、结算 支付通道（资金分配不变的情况）
type Action_3_ClosePaymentChannel struct {
	ChannelId fields.Bytes16 // 通道id
	// 数据指针
	// 所属交易
	trs block.Transaction
}

func (elm *Action_3_ClosePaymentChannel) Kind() uint16 {
	return 3
}

func (elm *Action_3_ClosePaymentChannel) SetBelongTrs(t block.Transaction) {
	elm.trs = t
}

func (elm *Action_3_ClosePaymentChannel) Size() uint32 {
	return 2 + elm.ChannelId.Size()
}

func (elm *Action_3_ClosePaymentChannel) Serialize() ([]byte, error) {
	var kindByte = make([]byte, 2)
	binary.BigEndian.PutUint16(kindByte, elm.Kind())
	var idBytes, _ = elm.ChannelId.Serialize()
	var buffer bytes.Buffer
	buffer.Write(kindByte)
	buffer.Write(idBytes)
	return buffer.Bytes(), nil
}

func (elm *Action_3_ClosePaymentChannel) Parse(buf []byte, seek uint32) (uint32, error) {
	seek, _ = elm.ChannelId.Parse(buf, seek)
	return seek, nil
}

func (elm *Action_3_ClosePaymentChannel) RequestSignAddrs() [][]byte {
	// 在执行的时候，查询出数据之后再检查检查签名
	return [][]byte{}
}

func (act *Action_3_ClosePaymentChannel) ChangeChainState(state state.ChainStateOperation) error {
	if act.trs == nil {
		panic("Action belong to transaction not be nil !")
	}
	// 查询通道
	paychanptr := state.Channel(act.ChannelId)
	if paychanptr == nil {
		return fmt.Errorf("Payment Channel Id <%s> not find.", hex.EncodeToString(act.ChannelId))
	}
	paychan := paychanptr.(*db.ChannelStoreItemData)
	// 判断通道已经关闭
	if paychan.IsClosed > 0 {
		return fmt.Errorf("Payment Channel <%s> is be closed.", hex.EncodeToString(act.ChannelId))
	}
	// 检查两个账户的签名
	signok, e1 := act.trs.VerifyNeedSigns([][]byte{paychan.LeftAddress, paychan.RightAddress})
	if e1 != nil {
		return e1
	}
	if !signok { // 签名检查失败
		return fmt.Errorf("Payment Channel <%s> address signature verify fail.", hex.EncodeToString(act.ChannelId))
	}
	// 通过时间计算利息
	leftAmount := paychan.LeftAmount
	rightAmount := paychan.RightAmount
	// 计算获得当前的区块高度
	//var curheight uint64 = 1
	blkptr := state.Block()
	if blkptr != nil {
		curheight := blkptr.(block.Block).GetHeight()
		// 增加利息计算，复利次数：约 8.68 天增加一次万分之一的复利，少于8天忽略不计
		insnum := (curheight - uint64(paychan.BelongHeight)) / 2500
		if insnum > 0 {
			a1, a2 := DoAppendCompoundInterest1Of10000By2500Height(&leftAmount, &rightAmount, insnum)
			leftAmount, rightAmount = *a1, *a2
		}
	}
	// 增加余额（将锁定的金额和利息从通道中提取出来）
	DoAddBalanceFromChainState(state, paychan.LeftAddress, leftAmount)
	DoAddBalanceFromChainState(state, paychan.RightAddress, rightAmount)
	// 暂时保留通道用于数据回退
	paychan.IsClosed = fields.VarInt1(1) // 标记通道已经关闭了
	state.ChannelCreate(act.ChannelId, paychan)
	//
	return nil
}

func (act *Action_3_ClosePaymentChannel) RecoverChainState(state state.ChainStateOperation) error {
	// 查询通道
	paychanptr := state.Channel(act.ChannelId)
	if paychanptr == nil {
		// 通道必须被保存，才能被回退
		panic(fmt.Errorf("Payment Channel Id <%s> not find.", hex.EncodeToString(act.ChannelId)))
	}
	paychan := paychanptr.(*db.ChannelStoreItemData)
	// 判断通道必须是已经关闭的状态
	if paychan.IsClosed != 0 {
		panic(fmt.Errorf("Payment Channel <%s> is be closed.", hex.EncodeToString(act.ChannelId)))
	}
	leftAmount := paychan.LeftAmount
	rightAmount := paychan.RightAmount
	// 计算差额
	blkptr := state.Block()
	if blkptr != nil {
		curheight := blkptr.(block.Block).GetHeight()
		// 增加利息计算，复利次数：约 8.68 天增加一次万分之一的复利，少于8天忽略不计
		insnum := (curheight - uint64(paychan.BelongHeight)) / 2500
		if insnum > 0 {
			a1, a2 := DoAppendCompoundInterest1Of10000By2500Height(&leftAmount, &rightAmount, insnum)
			leftAmount, rightAmount = *a1, *a2
		}
	}
	// 减除余额（重新将金额放入通道）
	DoSubBalanceFromChainState(state, paychan.LeftAddress, leftAmount)
	DoSubBalanceFromChainState(state, paychan.RightAddress, rightAmount)
	// 恢复通道状态
	paychan.IsClosed = fields.VarInt1(0) // 重新标记通道为开启状态
	state.ChannelCreate(act.ChannelId, paychan)
	return nil
}

package transactions

import (
	"bytes"
	"github.com/hacash/blockmint/block/actions"
	"github.com/hacash/blockmint/types/state"
	"math/big"

	"github.com/hacash/blockmint/block/fields"

	"golang.org/x/crypto/sha3"
)

type Transaction_0_Coinbase struct {
	Address fields.Address
	Reward  fields.Amount
	Message fields.TrimString16
	// nonce fields.VarInt8
	WitnessCount fields.VarInt1 // 投票见证人数量
	WitnessSigs  []uint8        // 见证人指定哈希尾数
	Witnesses    []fields.Sign  // 对prev区块hash的签名，投票分叉

	// cache data
	TotalFee fields.Amount // 区块总交易手续费
}

func NewTransaction_0_Coinbase() *Transaction_0_Coinbase {
	return &Transaction_0_Coinbase{
		WitnessCount: 0,
	}
}

func (trs *Transaction_0_Coinbase) GetReward() *fields.Amount {
	return &trs.Reward
}

func (trs *Transaction_0_Coinbase) Type() uint8 {
	return 0
}

func (trs *Transaction_0_Coinbase) Serialize() ([]byte, error) {

	var buffer bytes.Buffer
	b1, _ := trs.Address.Serialize()
	b2, _ := trs.Reward.Serialize()
	b3, _ := trs.Message.Serialize()
	// fmt.Println("trs.Message=", trs.Message)
	buffer.Write([]byte{trs.Type()}) // type
	buffer.Write(b1)
	buffer.Write(b2)
	buffer.Write(b3)
	// 见证人
	witnessCount := uint8(trs.WitnessCount)
	buffer.Write([]byte{witnessCount})
	for i := uint8(0); i < witnessCount; i++ {
		b := trs.WitnessSigs[i]
		buffer.Write([]byte{b})
	}
	for i := uint8(0); i < witnessCount; i++ {
		b := trs.Witnesses[i]
		s1, e := b.Serialize()
		if e != nil {
			return nil, e
		}
		buffer.Write(s1)
	}
	return buffer.Bytes(), nil
}

func (trs *Transaction_0_Coinbase) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = trs.ParseHead(buf, seek)
	if e != nil {
		return 0, e
	}
	// 见证人
	seek, e = trs.WitnessCount.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	if trs.WitnessCount > 0 {
		len := int(trs.WitnessCount)
		trs.WitnessSigs = make([]uint8, len)
		trs.Witnesses = make([]fields.Sign, len)
		for i := 0; i < len; i++ {
			trs.WitnessSigs[i] = buf[seek]
			seek++
		}
		for i := 0; i < len; i++ {
			var sign fields.Sign
			seek, e = sign.Parse(buf, seek)
			if e != nil {
				return 0, e
			}
			trs.Witnesses[i] = sign
		}
	}
	return seek, nil
}

func (trs *Transaction_0_Coinbase) ParseHead(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = trs.Address.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = trs.Reward.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = trs.Message.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (trs *Transaction_0_Coinbase) Size() uint32 {
	base := trs.Address.Size() + trs.Reward.Size() + trs.Message.Size()
	base += 1
	length := int(trs.WitnessCount)
	base += uint32(length)
	for i := 0; i < length; i++ {
		base += trs.Witnesses[i].Size()
	}
	return base
}

// 交易唯一哈希值
func (trs *Transaction_0_Coinbase) Hash() []byte {
	stuff, _ := trs.Serialize()
	digest := sha3.Sum256(stuff)
	return digest[:]
}

func (trs *Transaction_0_Coinbase) HashNoFee() []byte {
	return trs.Hash()
}

// 从 actions 拿出需要签名的地址

func (trs *Transaction_0_Coinbase) RequestSignAddrs() ([][]byte, error) {
	return make([][]byte, 0), nil
}

// 填充签名
func (trs *Transaction_0_Coinbase) FillNeedSigns(map[string][]byte) error {
	return nil
}

// 验证需要的签名
func (trs *Transaction_0_Coinbase) VerifyNeedSigns([][]byte) (bool, error) {
	return true, nil
}

// 需要的余额检查
func (trs *Transaction_0_Coinbase) RequestAddressBalance() ([][]byte, []big.Int, error) {
	return nil, nil, nil
}

// 修改 / 恢复 状态数据库
func (trs *Transaction_0_Coinbase) ChangeChainState(state state.ChainStateOperation) error {

	// fmt.Printf("trs.TotalFee = %s\n", trs.TotalFee.ToFinString())
	rwd, _ := trs.Reward.Add(&trs.TotalFee)
	// addr, _ := base58check.Encode(trs.Address)
	// fmt.Printf("coinbase.ChangeChainState,  %s  +=  %s\n", addr, rwd.ToFinString())
	return actions.DoAddBalanceFromChainState(state, trs.Address, *rwd)
}
func (trs *Transaction_0_Coinbase) RecoverChainState(state state.ChainStateOperation) error {
	rwd, _ := trs.Reward.Add(&trs.TotalFee)
	return actions.DoSubBalanceFromChainState(state, trs.Address, *rwd)
}

// 手续费含量
func (trs *Transaction_0_Coinbase) FeePurity() uint64 {
	return 0
}

// 查询
func (trs *Transaction_0_Coinbase) GetAddress() []byte {
	return []byte{}
}

func (trs *Transaction_0_Coinbase) GetFee() []byte {
	bts, _ := trs.TotalFee.Serialize()
	return bts
}

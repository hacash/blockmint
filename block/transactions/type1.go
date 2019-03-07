package transactions

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/bitcoin/address/address"
	"github.com/hacash/bitcoin/address/base58check"
	"github.com/hacash/bitcoin/address/btcec"
	"github.com/hacash/blockmint/block/actions"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/core/account"
	"github.com/hacash/blockmint/sys/err"
	typesblock "github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/state"
	"golang.org/x/crypto/sha3"
	"math/big"
)

type Transaction_1_Simple struct {
	Timestamp fields.VarInt5
	Address   fields.Address
	Fee       fields.Amount

	ActionCount fields.VarInt2
	Actions     []typesblock.Action

	SignCount fields.VarInt2
	Signs     []fields.Sign

	MultisignCount fields.VarInt2
	Multisigns     []fields.Multisign

	// cache data
	hash      []byte
	hashnofee []byte
}

func NewEmptyTransaction_1_Simple(master fields.Address) (*Transaction_1_Simple, error) {
	if !master.IsValid() {
		return nil, err.New("Master Address is InValid ")
	}
	timeUnix := 1550839748 //time.Now().Unix()
	return &Transaction_1_Simple{
		Timestamp:      fields.VarInt5(uint64(timeUnix)),
		Address:        master,
		Fee:            *fields.NewEmptyAmount(),
		ActionCount:    fields.VarInt2(0),
		SignCount:      fields.VarInt2(0),
		MultisignCount: fields.VarInt2(0),
	}, nil
}

func (trs *Transaction_1_Simple) Type() uint8 {
	return 1
}

func (trs *Transaction_1_Simple) Serialize() ([]byte, error) {
	body, e0 := trs.SerializeNoSign()
	if e0 != nil {
		return nil, e0
	}
	var buffer = new(bytes.Buffer)
	buffer.Write(body)
	// sign
	b1, e1 := trs.SignCount.Serialize()
	if e1 != nil {
		return nil, e1
	}
	buffer.Write(b1)
	for i := 0; i < int(trs.SignCount); i++ {
		var bi, e = trs.Signs[i].Serialize()
		if e != nil {
			return nil, e
		}
		buffer.Write(bi)
	}
	// muilt sign
	b2, e2 := trs.MultisignCount.Serialize()
	if e2 != nil {
		return nil, e2
	}
	buffer.Write(b2)
	for i := 0; i < int(trs.MultisignCount); i++ {
		var bi, e = trs.Multisigns[i].Serialize()
		if e != nil {
			return nil, e
		}
		buffer.Write(bi)
	}
	// ok
	return buffer.Bytes(), nil
}

func (trs *Transaction_1_Simple) SerializeNoSign() ([]byte, error) {
	return trs.SerializeNoSignEx(false)
}

func (trs *Transaction_1_Simple) SerializeNoSignEx(nofee bool) ([]byte, error) {
	var buffer = new(bytes.Buffer)
	b1, _ := trs.Timestamp.Serialize()
	b2, _ := trs.Address.Serialize()
	b3, _ := trs.Fee.Serialize()
	b4, _ := trs.ActionCount.Serialize()
	buffer.Write([]byte{trs.Type()}) // type
	buffer.Write(b1)
	buffer.Write(b2)
	if !nofee {
		buffer.Write(b3) // 费用付出着 签名 不需要 fee
	}
	buffer.Write(b4)
	for i := 0; i < len(trs.Actions); i++ {
		var bi, e = trs.Actions[i].Serialize()
		if e != nil {
			return nil, e
		}
		buffer.Write(bi)
	}
	//if nofee {
	//	fmt.Println( "SerializeNoSignEx: " + hex.EncodeToString(buffer.Bytes()))
	//}
	return buffer.Bytes(), nil
}

func (trs *Transaction_1_Simple) Parse(buf []byte, seek uint32) (uint32, error) {
	m1, _ := trs.Timestamp.Parse(buf, seek)
	m2, _ := trs.Address.Parse(buf, m1)
	m3, _ := trs.Fee.Parse(buf, m2)
	m4, _ := trs.ActionCount.Parse(buf, m3)
	iseek := m4
	for i := 0; i < int(trs.ActionCount); i++ {
		var act, sk, err = ParseAction(buf, iseek)
		trs.Actions = append(trs.Actions, act)
		iseek = sk
		if err != nil {
			return 0, err
		}
	}
	var e error
	iseek, e = trs.SignCount.Parse(buf, iseek)
	if e != nil {
		return 0, e
	}
	for i := 0; i < int(trs.SignCount); i++ {
		var sign fields.Sign
		iseek, e = sign.Parse(buf, iseek)
		if e != nil {
			return 0, e
		}
		trs.Signs = append(trs.Signs, sign)
	}
	iseek, e = trs.MultisignCount.Parse(buf, iseek)
	if e != nil {
		return 0, e
	}
	for i := 0; i < int(trs.MultisignCount); i++ {
		var multisign fields.Multisign
		iseek, e = multisign.Parse(buf, iseek)
		if e != nil {
			return 0, e
		}
		trs.Multisigns = append(trs.Multisigns, multisign)
	}
	return iseek, nil
}

func (trs *Transaction_1_Simple) Size() uint32 {
	totalsize := 1 + trs.Timestamp.Size() + trs.Address.Size() + trs.Fee.Size() + trs.ActionCount.Size()
	for i := 0; i < int(trs.ActionCount); i++ {
		totalsize += trs.Actions[i].Size()
	}
	return totalsize
}

// 交易唯一哈希值
func (trs *Transaction_1_Simple) Hash() []byte {
	if trs.hash == nil {
		return trs.HashFresh()
	}
	return trs.hash
}

func (trs *Transaction_1_Simple) HashFresh() []byte {
	stuff, _ := trs.SerializeNoSign()
	digest := sha3.Sum256(stuff)
	trs.hash = digest[:]
	return trs.hash
}

func (trs *Transaction_1_Simple) HashNoFee() []byte {
	if trs.hashnofee == nil {
		return trs.HashNoFeeFresh()
	}
	return trs.hashnofee
}
func (trs *Transaction_1_Simple) HashNoFeeFresh() []byte {
	notFee := true
	stuff, _ := trs.SerializeNoSignEx(notFee)
	digest := sha3.Sum256(stuff)
	trs.hashnofee = digest[:]
	return trs.hashnofee
}

func (trs *Transaction_1_Simple) AppendAction(action typesblock.Action) error {
	if trs.ActionCount >= 65535 {
		return err.New("Action too much")
	}
	trs.ActionCount += 1
	trs.Actions = append(trs.Actions, action)
	return nil
}

// 从 actions 拿出需要签名的地址
func (trs *Transaction_1_Simple) RequestSignAddrs() ([][]byte, error) {
	if !trs.Address.IsValid() {
		return nil, err.New("Master Address is InValid ")
	}
	requests := make([][]byte, 0)
	for i := 0; i < int(trs.ActionCount); i++ {
		actreqs := trs.Actions[i].RequestSignAddrs()
		requests = append(requests, actreqs...)
	}
	// 去重
	results := make([][]byte, len(requests))
	has := make(map[string]bool)
	has[string(trs.Address)] = true // 费用方去除
	for i := 0; i < len(requests); i++ {
		strkey := string(requests[i])
		if _, ok := has[strkey]; !ok {
			results = append(results, requests[i])
		}
	}
	// 返回
	return requests, nil
}

// 填充签名
func (trs *Transaction_1_Simple) FillNeedSigns(addrPrivates map[string][]byte) error {
	hash := trs.HashFresh()
	hashNoFee := trs.HashNoFee()
	requests, e0 := trs.RequestSignAddrs()
	if e0 != nil {
		return e0
	}
	// 主签名
	e1 := trs.addOneSign(hashNoFee, addrPrivates, trs.Address)
	if e1 != nil {
		return e1
	}
	// 其他签名
	for i := 0; i < len(requests); i++ {
		e1 := trs.addOneSign(hash, addrPrivates, requests[i])
		if e1 != nil {
			return e1
		}
	}
	// 填充成功
	return nil
}

func (trs *Transaction_1_Simple) addOneSign(hash []byte, addrPrivates map[string][]byte, address []byte) error {
	privitebytes, has := addrPrivates[string(address)]
	if !has {
		return err.New("Private Key '" + base58check.Encode(address) + "' necessary")
	}
	privite, e1 := account.GetAccountByPriviteKey(privitebytes)
	if e1 != nil {
		return err.New("Private Key '" + base58check.Encode(address) + "' error")
	}
	signature, e2 := privite.Private.Sign(hash)
	if e2 != nil {
		return err.New("Private Key '" + base58check.Encode(address) + "' do sign error")
	}
	// append
	trs.SignCount += 1
	trs.Signs = append(trs.Signs, fields.Sign{
		PublicKey: privite.PublicKey,
		Signature: signature.Serialize64(),
	})
	return nil
}

// 验证需要的签名
func (trs *Transaction_1_Simple) VerifyNeedSigns() (bool, error) {
	hash := trs.HashFresh()
	hashNoFee := trs.HashNoFee()
	requests, e0 := trs.RequestSignAddrs()
	if e0 != nil {
		return false, e0
	}
	allSigns := make(map[string]fields.Sign)
	for i := 0; i < len(trs.Signs); i++ {
		sig := trs.Signs[i]
		addr := address.NewAddressFromPublicKey([]byte{0}, sig.PublicKey)
		allSigns[string(addr)] = sig
	}
	// 验证主签名
	ok, e := verifyOneSignature(allSigns, trs.Address, hashNoFee)
	if e != nil || !ok {
		return ok, e
	}
	// 验证其他所有签名
	for i := 0; i < len(requests); i++ {
		ok, e := verifyOneSignature(allSigns, trs.Address, hash)
		if e != nil || !ok {
			return ok, e
		}
	}
	// 验证成功
	return true, nil
}

func verifyOneSignature(allSigns map[string]fields.Sign, address fields.Address, hash []byte) (bool, error) {

	main, ok := allSigns[string(address)]
	if !ok {
		return false, nil
	}
	sigobj, e3 := btcec.ParseSignatureByte64(main.Signature)
	if e3 != nil {
		return false, e3
	}
	pubKey, e4 := btcec.ParsePubKey(main.PublicKey, btcec.S256())
	if e4 != nil {
		return false, e4
	}
	verok := sigobj.Verify(hash, pubKey)
	if !verok {
		return false, nil
	}
	// ok
	return true, nil
}

// 需要的余额检查
func (trs *Transaction_1_Simple) RequestAddressBalance() ([][]byte, []big.Int, error) {
	return nil, nil, nil
}

// 修改 / 恢复 状态数据库
func (trs *Transaction_1_Simple) ChangeChainState(state state.ChainStateOperation) error {
	// actions
	for i := 0; i < len(trs.Actions); i++ {
		trs.Actions[i].SetBelongTrs(trs)
		e := trs.Actions[i].ChangeChainState(state)
		if e != nil {
			return e
		}
	}
	// 扣除手续费
	return actions.DoSubBalanceFromChainState(state, trs.Address, trs.Fee)
}

func (trs *Transaction_1_Simple) RecoverChainState(state state.ChainStateOperation) error {
	// actions
	for i := len(trs.Actions) - 1; i >= 0; i-- {
		trs.Actions[i].SetBelongTrs(trs)
		e := trs.Actions[i].RecoverChainState(state)
		if e != nil {
			return e
		}
	}
	// 回退手续费
	return actions.DoAddBalanceFromChainState(state, trs.Address, trs.Fee)
}

// 手续费含量 每byte的含有多少烁代币
func (trs *Transaction_1_Simple) FeePurity() uint64 {

	bigfee := trs.Fee.GetValue()
	bigfee = bigfee.Div(bigfee, fields.NewAmountNumOneByUnit(232).GetValue())
	bigfee = bigfee.Div(bigfee, new(big.Int).SetUint64(uint64(trs.Size())))
	maxUint64 := uint64(18446744073709551615)
	if bigfee.Cmp(new(big.Int).SetUint64(maxUint64)) == 1 {
		return maxUint64
	}
	return bigfee.Uint64()
}

// 查询
func (trs *Transaction_1_Simple) GetAddress() []byte {
	return trs.Address
}

func (trs *Transaction_1_Simple) GetFee() []byte {
	feebts, _ := trs.Fee.Serialize()
	return feebts
}

/* *********************************************************** */

func NewActionByKind(kind uint16) (typesblock.Action, error) {
	switch kind {
	////////////////////   ACTIONS   ////////////////////
	case 1:
		return new(actions.Action_1_SimpleTransfer), nil
		////////////////////    END      ////////////////////
	}
	return nil, err.New("Cannot find Action kind of " + string(kind))
}

func ParseAction(buf []byte, seek uint32) (typesblock.Action, uint32, error) {
	var kind = binary.BigEndian.Uint16(buf[seek : seek+2])
	var act, _ = NewActionByKind(kind)
	var mv, err = act.Parse(buf, seek+2)
	return act, mv, err
}

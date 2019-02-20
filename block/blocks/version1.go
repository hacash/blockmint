package blocks

import (
	"bytes"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/sys/err"
	typesblock "github.com/hacash/blockmint/types/block"
)

type Block_v1 struct {
	// Version   fields.VarInt1
	Height           fields.VarInt5
	Timestamp        fields.VarInt5
	PrevMark         fields.Bytes32
	MrklRoot         fields.Bytes32
	TransactionCount fields.VarInt4
	// meta
	Nonce fields.VarInt4 // 挖矿随机值
	// body
	Transactions []typesblock.Transaction

	// cache data
	hash []byte
}

func (block *Block_v1) Version() uint8 {
	return 1
}

func (block *Block_v1) SerializeHead() ([]byte, error) {
	var buffer = new(bytes.Buffer)
	//b1, _ := block.Version.Serialize()
	buffer.Write([]byte{block.Version()})
	b2, _ := block.Height.Serialize()
	b3, _ := block.Timestamp.Serialize()
	b4, _ := block.PrevMark.Serialize()
	b5, _ := block.MrklRoot.Serialize()
	b6, _ := block.TransactionCount.Serialize()
	buffer.Write(b2)
	buffer.Write(b3)
	buffer.Write(b4)
	buffer.Write(b5)
	buffer.Write(b6)
	return buffer.Bytes(), nil
}

func (block *Block_v1) SerializeBody() ([]byte, error) {

	var buffer = new(bytes.Buffer)
	b1, e1 := block.SerializeMeta()
	if e1 != nil {
		return nil, e1
	}
	b2, e2 := block.SerializeTransactions(nil)
	if e2 != nil {
		return nil, e2
	}
	buffer.Write(b1)
	buffer.Write(b2)
	return buffer.Bytes(), nil

}

func (block *Block_v1) SerializeMeta() ([]byte, error) {
	var buffer = new(bytes.Buffer)
	b1, _ := block.Nonce.Serialize() // miner nonce
	buffer.Write(b1)
	return buffer.Bytes(), nil

}

func (block *Block_v1) SerializeTransactions(itr typesblock.SerializeTransactionsIterator) ([]byte, error) {
	var buffer = new(bytes.Buffer)
	var trslen = uint32(len(block.Transactions))
	if itr != nil { // 迭代器
		itr.Init(trslen)
	}
	for i := uint32(0); i < trslen; i++ {
		var trs = block.Transactions[i]
		var bi, e = trs.Serialize()
		if e != nil {
			return nil, e
		}
		buffer.Write(bi)
		if itr != nil { // 迭代器
			itr.FinishOneTrs(i, trs, bi)
		}
	}
	return buffer.Bytes(), nil

}

// 序列化 与 反序列化
func (block *Block_v1) Serialize() ([]byte, error) {

	var buffer = new(bytes.Buffer)

	head, _ := block.SerializeHead()
	buffer.Write(head)
	body, _ := block.SerializeBody()
	buffer.Write(body)

	return buffer.Bytes(), nil
}

func (block *Block_v1) ParseHead(buf []byte, seek uint32) (uint32, error) {
	//fmt.Println(*buf)
	//fmt.Println(seek)
	//fmt.Println((*buf)[seek:])
	//m1, _ := block.Version.Parse(buf, seek)
	m2, _ := block.Height.Parse(buf, seek)
	m3, _ := block.Timestamp.Parse(buf, m2)
	m4, _ := block.PrevMark.Parse(buf, m3)
	m5, _ := block.MrklRoot.Parse(buf, m4)
	m6, _ := block.TransactionCount.Parse(buf, m5)
	iseek := m6
	return iseek, nil
}

func (block *Block_v1) ParseBody(buf []byte, seek uint32) (uint32, error) {
	seek, _ = block.ParseMeta(buf, seek)
	seek, _ = block.ParseTransactions(buf, seek)
	return seek, nil
}

func (block *Block_v1) ParseMeta(buf []byte, seek uint32) (uint32, error) {
	seek, _ = block.Nonce.Parse(buf, seek) // miner nonce
	return seek, nil
}

func (block *Block_v1) ParseTransactions(buf []byte, seek uint32) (uint32, error) {
	length := int(block.TransactionCount)
	block.Transactions = make([]typesblock.Transaction, length)
	for i := 0; i < length; i++ {
		var trx, sk, err = ParseTransaction(buf, seek)
		block.Transactions[i] = trx
		seek = sk
		if err != nil {
			return seek, err
		}
	}
	return seek, nil

}

func (block *Block_v1) Parse(buf []byte, seek uint32) (uint32, error) {
	// head
	iseek, _ := block.ParseHead(buf, seek)
	iseek2, _ := block.ParseBody(buf, iseek)
	return iseek2, nil
}

func (block *Block_v1) Size() uint32 {
	totalsize := 1 +
		block.Height.Size() +
		block.Timestamp.Size() +
		block.PrevMark.Size() +
		block.MrklRoot.Size() +
		block.TransactionCount.Size()
	for i := uint32(0); i < uint32(block.TransactionCount); i++ {
		totalsize += block.Transactions[i].Size()
	}
	return totalsize
}

// HASH
func (block *Block_v1) Hash() []byte {
	if block.hash == nil {
		block.hash = CalculateBlockHash(block)
	}
	return block.hash
}

func (block *Block_v1) GetTransactions() []typesblock.Transaction {
	return block.Transactions
}
func (block *Block_v1) AddTransaction(trs typesblock.Transaction) {
	block.Transactions = append(block.Transactions, trs)
}

////////////////////////////////////////////////////////////////////////

func NewTransactionByType(ty uint8) (typesblock.Transaction, error) {
	switch ty {
	////////////////////  TRANSATION  ////////////////////
	case 0:
		return new(transactions.Transaction_0_Coinbase), nil
	case 1:
		return new(transactions.Transaction_1_Simple), nil
		////////////////////     END      ////////////////////
	}
	return nil, err.New("Cannot find Transaction type of " + string(ty))
}

func ParseTransaction(buf []byte, seek uint32) (typesblock.Transaction, uint32, error) {
	ty := uint8(buf[seek])
	var trx, _ = NewTransactionByType(ty)
	var mv, err = trx.Parse(buf, seek+1)
	return trx, mv, err
}

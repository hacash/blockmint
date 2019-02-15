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
	Height    fields.VarInt5
	Timestamp fields.VarInt5

	PrevMark fields.Bytes32
	MrklRoot fields.Bytes32

	Nonce fields.VarInt4 // 挖矿随机值

	TransactionCount fields.VarInt4

	// body
	Transactions []typesblock.Transaction
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

// 序列化 与 反序列化
func (block *Block_v1) Serialize() ([]byte, error) {

	var buffer = new(bytes.Buffer)

	head, _ := block.SerializeHead()
	buffer.Write(head)
	b6, _ := block.Nonce.Serialize() // miner nonce
	buffer.Write(b6)

	for i := 0; i < len(block.Transactions); i++ {
		var bi, _ = block.Transactions[i].Serialize()
		buffer.Write(bi)
	}
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

func (block *Block_v1) Parse(buf []byte, seek uint32) (uint32, error) {
	// head
	iseek, _ := block.ParseHead(buf, seek)
	iseek, _ = block.Nonce.Parse(buf, iseek) // miner nonce
	// body
	length := int(block.TransactionCount)
	for i := 0; i < length; i++ {
		var trx, sk, err = ParseTransaction(buf, iseek)

		block.Transactions = append(block.Transactions, trx)
		iseek = sk
		if err != nil {
			break
		}
	}
	return iseek, nil
}

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

package pool

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/actions"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/chain/state/db"
	"github.com/hacash/blockmint/service/txpool"
	"math/big"
	"time"
)

// 检查交易确认
func (mp *MiningPool) checkTransactionConfirm() error {

	blkdb := store.GetGlobalInstanceBlocksDataStore()

	trc := mp.StoreDB.ReadTransferRecord(false) // 读取统计记录
	if trc.TxConfirm >= trc.TxLatestId {
		return nil
	}
	trc_end := trc.TxConfirm + 20
	if trc_end > trc.TxLatestId {
		trc_end = trc.TxLatestId
	}
	var tx_confirm = trc.TxConfirm
	// 依次查看或提交交易确认
	for i := trc.TxConfirm + 1; i <= trc_end; i++ {
		txbody := mp.StoreDB.ReadTransactionBody(i)
		if len(txbody) > 0 {
			tx, _, err := blocks.ParseTransaction(txbody, 0)
			if err == nil {
				if ok, e := blkdb.CheckTransactionExist(tx.HashNoFee()); ok && e == nil {
					tx_confirm = i // 交易得到确认
				}
			}
		}
	}
	// 保存交易确认
	trc_save := mp.StoreDB.ReadTransferRecord(true) // 读取统计记录
	trc_save.TxConfirm = tx_confirm
	mp.StoreDB.SaveTransferRecord(trc_save, true)

	return nil
}

// 打币转账
func (mp *MiningPool) transfer(wk *PowWorker) {
	// 转账数量
	payrwds := wk.StatisticsData.DeservedRewards
	// 扣除代发奖励统计
	wk.StatisticsData.CompleteRewards += payrwds
	wk.StatisticsData.DeservedRewards = 0
	wk.StatisticsData.PrevTransferBlockHeight = uint32(mp.StateData.CurrentMiningBlock.Block.GetHeight())
	mp.StoreDB.SaveWorker(wk) // 保存
	// 生成转账交易
	trc := mp.StoreDB.ReadTransferRecord(true) // 读取统计记录
	trc.Latest += 1
	defer mp.StoreDB.SaveTransferRecord(trc, true) // 保存统计
	var td TransferData
	td.Id = trc.Latest
	td.TxId = 0 // 还未有
	td.CreateTimestamp = uint64(time.Now().Unix())
	td.Amount = payrwds
	td.Address = *wk.RewordAddress
	mp.StoreDB.SaveTransfer(&td) // 保存转账交易
	fmt.Printf("mining pool create save transfer %d, amount: ㄜ%d:240, address: %s \n", trc.Latest, payrwds, wk.RewordAddress.ToReadable())
}

// 创建并发送交易
func (mp *MiningPool) createSendTransaction() error {
	trc := mp.StoreDB.ReadTransferRecord(false) // 读取记录
	if trc.Latest == trc.Submit {
		return nil
	}
	if !(mp.StateData.CurrentMiningBlock.Block.GetHeight()-trc.PrevSendHeight >= 10 || trc.Latest-trc.Submit >= 20) {
		return nil // 十个区块 或者 20笔交易 发起一次转账
	}
	// 读取交易，最多二十条
	curtxid := trc.TxLatestId + 1
	trc_length := trc.Latest - trc.Submit
	if trc_length > 20 {
		trc_length = 20
	}
	alltrs := make(map[string]uint64)
	alltds := make([]*TransferData, 0, 10)
	for i := trc.Submit + 1; i <= trc.Submit+trc_length; i++ {
		td := mp.StoreDB.ReadTransfer(i)
		if td != nil {
			if _, h := alltrs[string(td.Address)]; !h {
				alltrs[string(td.Address)] = 0 // init
			}
			alltrs[string(td.Address)] += td.Amount
			td.TxId = curtxid
			alltds = append(alltds, td)
		}
	}
	// 创建转账
	if len(alltds) == 0 {
		return nil
	}
	var totalfee uint64 = 5000
	var totalamt uint64 = 0
	var action_count = 0
	tx, _ := transactions.NewEmptyTransaction_2_Simple(mp.StateData.FeeAccount.Address)
	for k, v := range alltrs {
		payamt, e2 := fields.NewAmountByBigIntWithUnit(new(big.Int).SetUint64(v), 240)
		if e2 == nil {
			totalfee += 5000 // 手续费
			totalamt += v
			act := actions.Action_1_SimpleTransfer{
				Address: fields.Address(k),
				Amount:  *payamt,
			}
			tx.AppendAction(&act) // 添加转账
			action_count++
		}
	}
	total_amount, e3 := fields.NewAmountByBigIntWithUnit(new(big.Int).SetUint64(totalamt), 240)
	if e3 != nil {
		return e3
	}
	// 检查余额
	feeaddr := mp.StateData.FeeAccount.Address
	sto, err := db.GetGlobalInstanceBalanceDB().Read(feeaddr)
	if err != nil {
		return fmt.Errorf("MiningPool transfer error: %s", err)
	}
	if sto.Amount.LessThan(total_amount) {
		// 余额不足
		return fmt.Errorf("MiningPool transfer error: fee address amount not enough, need %s but %s \n", total_amount.ToFinString(), sto.Amount.ToFinString())
	}
	// 设置手续费
	total_fee, e4 := fields.NewAmountByBigIntWithUnit(new(big.Int).SetUint64(totalfee), 240)
	if e4 != nil {
		return e4
	}
	tx.Fee = *total_fee
	// 签名交易
	privates := make(map[string][]byte)
	privates[string(feeaddr)] = mp.StateData.FeeAccount.PrivateKey
	e6 := tx.FillNeedSigns(privates, nil)
	if e6 != nil {
		return fmt.Errorf("sign transaction error, " + e6.Error())
	}
	// 检查签名
	sigok, sigerr := tx.VerifyNeedSigns(nil)
	if sigerr != nil {
		return fmt.Errorf("transaction VerifyNeedSigns error")
	}
	if !sigok {
		return fmt.Errorf("transaction VerifyNeedSigns fail")
	}
	// 加入交易池
	e5 := txpool.GetGlobalInstanceMemTxPool().AddTx(tx)
	if e5 != nil {
		return e5
	}
	// 保存统计
	trc_save := mp.StoreDB.ReadTransferRecord(true)
	trc_save.PrevSendHeight = mp.StateData.CurrentMiningBlock.Block.GetHeight() // 转账记录
	trc_save.Submit += trc_length
	trc_save.TxLatestId += 1
	defer mp.StoreDB.SaveTransferRecord(trc_save, true)
	// 保存交易
	mp.StoreDB.SaveTransactionBody(trc_save.TxLatestId, tx)
	// 保存全部交易
	curtime := uint64(time.Now().Unix())
	for _, td := range alltds {
		td.TxId = trc_save.TxLatestId
		td.SubmitTimestamp = curtime
		mp.StoreDB.SaveTransfer(td)
	}
	fmt.Printf("-+-+-+-+- miner pool put transaction %d, amt ㄜ%d:240, actions %d, hash <%s> to mem pool.\n", trc_save.TxLatestId, totalamt, action_count, hex.EncodeToString(tx.HashNoFee()))

	// 成功返回
	return nil
}

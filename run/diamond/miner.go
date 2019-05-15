package diamond

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/actions"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/core/account"
	"github.com/hacash/blockmint/miner"
	"github.com/hacash/blockmint/service/txpool"
	"github.com/hacash/x16rs"
	"math/rand"
	"strings"
	"time"
)

type ReStartMinerStat struct {
	Number   uint32
	PrevHash []byte
}

type DiamondMiner struct {
	blkminer *miner.HacashMiner

	// 有几个地址就开几个线程
	rewards     []fields.Address
	feePassword string // 手续费地址私钥

	// 停止并启动下一轮挖矿
	statCh     chan *ReStartMinerStat
	stopmarkCh chan bool
}

func NewDiamondMiner() *DiamondMiner {
	return &DiamondMiner{}
}

func (dm *DiamondMiner) Start(blkminer *miner.HacashMiner) error {

	// 数据初始化
	dm.blkminer = blkminer
	dm.statCh = make(chan *ReStartMinerStat, 1)
	dm.stopmarkCh = make(chan bool, 10)

	// 检查配置
	dm.feePassword = config.Config.DiamondMiner.Feepassword
	if len(dm.feePassword) < 6 { // 手续费地址私钥
		panic(fmt.Sprintf("Fee Secret Must."))
	}
	dm.rewards = make([]fields.Address, 0, 8)
	rwds := config.Config.DiamondMiner.Rewards
	if len(rwds) < 1 {
		panic(fmt.Sprintf("Rewards Address Must."))
	}
	for _, v := range rwds {
		address, e := fields.CheckReadableAddress(v)
		if e != nil {
			panic(fmt.Sprintf("Reward Address '%s' format error.", v))
		}
		dm.rewards = append(dm.rewards, *address)
	}

	// 首次进入状态初始
	dm.ChangeMinerState()

	// 监听状态改变
	go dm.ListenMinerStateChange()

	// 开始挖矿
	for {
		stat := <-dm.statCh
		dm.DoMining(stat)
	}
	return nil
}

func (dm *DiamondMiner) DoMining(stat *ReStartMinerStat) error {

	//supervene := len(dm.rewards)
	var successAddr *fields.Address = nil
	successDiamond := ""
	successCh := make(chan []byte)
	stopMark := byte(0)

	supervene := len(dm.rewards)

	fmt.Printf("--◈--◈--◈-- do diamond mining, supervene:%d, number:%d prevhash:<%s>   \n", supervene, stat.Number+1, hex.EncodeToString(stat.PrevHash))

	for i, addr := range dm.rewards {
		// 开启挖矿线程
		go func(i int, addr fields.Address) {
			loopnum := (stat.Number+1)/2048 + 1
			if loopnum > 16 {
				loopnum = 16
			}
			nonce, diastr := x16rs.MinerHacashDiamond(int(stat.Number+1), int(loopnum), &stopMark, stat.PrevHash, addr)
			//fmt.Println(hex.EncodeToString(nonce), diastr, addr.ToReadable())
			if dm.CheckDiamond(stat, nonce, addr, diastr) {
				if successAddr == nil {
					// 挖掘成功
					fmt.Printf("◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈◈ 【%s】 number:%d, prevhash:<%s>, nonce:<%s>, addr:%s,  mining successfully!\n", diastr, stat.Number+1, hex.EncodeToString(stat.PrevHash), hex.EncodeToString(nonce), addr.ToReadable())
					successAddr = &addr
					successDiamond = string([]byte(diastr)[10:])
					stopMark = 1 // 停止挖掘
					successCh <- nonce
				} else {
					fmt.Printf("[%d] miner '%s' finish out.\n", i, addr.ToReadable())
				}
			} else {
				if stopMark > 0 {
					fmt.Printf("[%d] miner '%s' break out.\n", i, addr.ToReadable())
				} else {
					fmt.Printf("[%d] miner '%s' over max nonce.\n", i, addr.ToReadable())
				}

			}
		}(i, addr)
	}
	// 停止
	go func() {
		<-dm.stopmarkCh
		stopMark = 1 // 停止挖掘
		successCh <- nil
	}()

	// 等待挖掘结果
	successNonce := <-successCh
	if successNonce == nil {
		// 停止本轮挖掘
		return nil
	}

	// 创建并发送钻石交易
	fmt.Println("create and send transaction...")
	dm.CreateAndSendTransaction(stat, successDiamond, successNonce, *successAddr)

	// 等待交易成功，下一轮
	return nil
}

// 创建并发送钻石交易
func (dm *DiamondMiner) CreateAndSendTransaction(stat *ReStartMinerStat, diamond string, nonce []byte, address fields.Address) {

	// 创建 action
	var dimcreate actions.Action_4_DiamondCreate
	dimcreate.Number = fields.VarInt3(stat.Number + 1)
	dimcreate.Diamond = fields.Bytes6(diamond)
	dimcreate.PrevHash = stat.PrevHash
	dimcreate.Nonce = fields.Bytes8(nonce)
	dimcreate.Address = address

	// 拿出手续费账户
	feeacc := account.CreateAccountByPassword(dm.feePassword)

	// 创建交易
	newTrs, e5 := transactions.NewEmptyTransaction_2_Simple(feeacc.Address)
	newTrs.Timestamp = fields.VarInt5(time.Now().Unix()) // 使用 hold 的时间戳
	if e5 != nil {
		panic("create transaction error, " + e5.Error())
	}
	newTrs.Fee = *fields.NewAmountSmall(4, 244) // set fee
	// 放入action
	newTrs.AppendAction(&dimcreate)

	// 数据化
	_, e7 := newTrs.Serialize()
	if e7 != nil {
		panic("transaction serialize error, " + e7.Error())
	}
	// sign
	privates := make(map[string][]byte)
	privates[string(feeacc.Address)] = feeacc.PrivateKey
	e6 := newTrs.FillNeedSigns(privates, nil)
	if e6 != nil {
		panic("sign transaction error, " + e6.Error())
	}
	// 检查签名
	sigok, sigerr := newTrs.VerifyNeedSigns(nil)
	if sigerr != nil {
		panic("transaction VerifyNeedSigns error")
	}
	if !sigok {
		panic("transaction VerifyNeedSigns fail")
	}

	// 加入交易池
	err := txpool.GetGlobalInstanceMemTxPool().AddTx(newTrs)

	if err != nil {
		fmt.Println(err)
	}else{
		fmt.Printf("put trs <%s> to mem pool.\n", hex.EncodeToString(newTrs.HashNoFee()))
	}


}

// 判断是否为合格的钻石
func (dm *DiamondMiner) CheckDiamond(stat *ReStartMinerStat, nonce []byte, address fields.Address, diamond string) bool {
	// 检查钻石挖矿计算
	diamond_resbytes, diamond_str := x16rs.Diamond(uint32(stat.Number+1), stat.PrevHash, nonce, address)
	_, isdia := x16rs.IsDiamondHashResultString(diamond_str)
	if !isdia {
		return false // fmt.Errorf("String <%s> is not diamond.", diamond_str)
	}
	if strings.Compare(diamond_str, string(diamond)) != 0 {
		return false // fmt.Errorf("Diamond need <%s> but got <%s>", act.Diamond, diamondstrval)
	}
	// 检查钻石难度值
	difok := x16rs.CheckDiamondDifficulty(uint32(stat.Number+1), diamond_resbytes)
	if !difok {
		return false // fmt.Errorf("Diamond difficulty not meet the requirements.")
	}
	// 检查成功
	return true
}

func (dm *DiamondMiner) ChangeMinerState() error {
	num, hash := dm.blkminer.GetPrevDiamondHash()
	dm.statCh <- &ReStartMinerStat{
		num,
		hash,
	}
	return nil
}

func (dm *DiamondMiner) ListenMinerStateChange() {
	num1, _ := dm.blkminer.GetPrevDiamondHash()
	for {
		time.Sleep(time.Second) // 休眠一秒
		num2, _ := dm.blkminer.GetPrevDiamondHash()
		if num1 != num2 { // 对比是否有变化
			dm.ChangeMinerState()
			num1 = num2
			dm.stopmarkCh <- true // 停止之前的
		}
	}
}

// 随机取一个地址
func (dm *DiamondMiner) GetRandomMinerRewardAddress() fields.Address {
	length := len(dm.rewards)
	if length == 0 {
		panic("Miner Reward Address must be give at lest one !")
	}
	idx := rand.Intn(length)
	//fmt.Println(idx)
	return dm.rewards[idx]
}

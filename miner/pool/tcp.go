package pool

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/miner/difficulty"
	"github.com/hacash/x16rs"
	"log"
	"math/big"
	"math/rand"
	"net"
	"time"
)

// 监听端口

func (ps *PoolState) startListenAndServe(address string) {
	// 绑定监听地址
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(fmt.Sprintf("listen err: %v", err))
	}
	defer listener.Close()
	log.Println(fmt.Sprintf("bind: %s, start listening...", address))

	for {
		// Accept 会一直阻塞直到有新的连接建立或者listen中断才会返回
		conn, err := listener.Accept()
		if err != nil {
			// 通常是由于listener被关闭无法继续监听导致的错误
			log.Fatal(fmt.Sprintf("accept err: %v", err))
		}
		// 开启新的 goroutine 处理该连接
		go ps.handleConn(conn)
	}
}

func (ps *PoolState) handleConn(conn net.Conn) {
	if ps.CurrentMiningBlock == nil {
		// return // 数据没准备好，先断开，下次重连
	}

	cctime := time.Now()
	var client = &Client{
		Id:              rand.Uint64(), // 标识符
		Conn:            conn,
		ActiveTimestamp: &cctime, // 活跃时间
		//CurrentMiningMarkNumber: 0,
	}

	// 使用 bufio 标准库提供的缓冲区功能
	reader := bufio.NewReader(conn)
	for {
		// ReadString 会一直阻塞直到遇到分隔符 '\n'
		// 遇到分隔符后会返回上次遇到分隔符或连接建立后收到的所有数据, 包括分隔符本身
		// 若在遇到分隔符之前遇到异常, ReadString 会返回已收到的数据和错误信息
		msgbytes, err := x16rs.MiningPoolReadTcpMsgBytes(reader)
		if err != nil {
			ps.pool.removeCloseClient(client)
			return
		}

		//println(msgbytes)

		// 处理数据
		err1 := ps.handleMessage(client, msgbytes[0], msgbytes[1:])
		if err1 != nil {
			return // 发生错误并返回
		}
	}
}

// 处理消息
func (ps *PoolState) handleMessage(client *Client, typenum uint8, msgcon []byte) error {
	switch typenum {

	case 11: // 心跳活跃

		cctime := time.Now()
		client.ActiveTimestamp = &cctime

	case 0: // 注册奖励地址
		// 检查数量上限
		if ps.ClientCount+1 > ps.MaxClientCount {
			// 已经满员，返回错误信息
			x16rs.MiningPoolWriteTcpMsgBytes(client.Conn, 255, []byte(`Pool miner member has exceeded the maximum, Please contact the mine service provider\n矿池数量已经超出最大值，暂时无法连接挖矿，请联系你的矿池服务商`))
			return fmt.Errorf("Max Client Count overflow")
		}
		// 检查并注册
		addr, err := fields.CheckReadableAddress(string(msgcon))
		if err != nil {
			return fmt.Errorf("rewards address error")
		}
		ps.register(client, addr)
		// 注册完成
		log.Println(fmt.Sprintf("pool client id:%d, address:%s register", client.Id, string(msgcon)))

	case 3: // 接收算力统计

		if len(msgcon) == 6 {
			ps.addPowCount(client, msgcon)
			if msgcon[5] == 1 { // 发送新区块
				ps.sendMiningStuffData(client)
			}
		}

	case 2: // 【【【挖出区块】】】

		var success x16rs.MiningSuccess
		success.Parse(msgcon, 0)
		ps.gotSuccessBlock(client, &success)

	}
	return nil
}

// 成功挖出区块
func (ps *PoolState) gotSuccessBlock(client *Client, success *x16rs.MiningSuccess) {
	if ps.CurrentMiningBlock == nil {
		return
	}
	curblock := ps.CurrentMiningBlock.Block
	ps.updateSetCurrentMiningBlockCoinbaseMessage(success.MiningIndex, success.Nonce)
	hash := curblock.HashFresh()
	curdiff := difficulty.HashToBig(curblock.GetHeight(), hash)
	targetDifficultyHash := difficulty.Uint32ToBig(curblock.GetHeight(), curblock.GetDifficulty())
	//redouble := int64(1) // 算力倍数/
	if curdiff.Cmp(targetDifficultyHash) <= 0 {
		log.Println("mining pool find a valid nonce for block", "height", curblock.GetHeight(), "address", client.RewordAddress.ToReadable())
		// 写入区块
		ps.SuccessFindBlockHash = hash                          // 成功的哈希
		ps.SuccessFindBlockRewardAddress = client.RewordAddress // 成功挖出的地址
		ps.pool.CalcSuccessBlockCh <- *success
		//// 成功挖出区块，算力记3倍
		//redouble = 3
	} else {
		//redouble = 1
		log.Println("hash not valid", curblock.GetHeight(), hex.EncodeToString(hash))
	}

	// 增加算力统计，挖出区块加上三倍算力
	// ps.addMinerPowerValue(, client, redouble)

}

// 统计算力
func (ps *PoolState) addPowCount(client *Client, msgcon []byte) {
	// 先统计上一个区块的算力
	if client.MiningBlockStuffPrev != nil {
		stuff := client.MiningBlockStuffPrev
		copy(stuff[79:83], msgcon[1:5])
		//lp := ps.CurrentMiningBlock.Block.GetHeight() / 50000 + 1
		hash := blocks.CalculateBlockHashByStuff(int(msgcon[0]), stuff)
		ps.addMinerPowerValue(client, hash, 1) // 统计算力
		// 每个区块头运算只统计一次
		client.MiningBlockStuffPrev = nil
		return
	}
	// 检查当前正在挖的区块
	if client.MiningBlockStuffCurrent != nil {
		stuff := client.MiningBlockStuffCurrent
		copy(stuff[79:83], msgcon[1:5])
		hash := blocks.CalculateBlockHashByStuff(int(msgcon[0]), stuff)
		ps.addMinerPowerValue(client, hash, 1) // 统计算力
		// 每个区块头运算只统计一次
		client.MiningBlockStuffCurrent = nil
		return
	}
}

// 增加算力统计
func (ps *PoolState) addMinerPowerValue(client *Client, hash []byte, redouble int64) *big.Int {
	if redouble < 1 || redouble > 5 {
		redouble = 1 // 合理的取值范围： 1~5 倍
	}
	// fmt.Printf("addMinerPowerValue hash: %s, address: %s\n", hex.EncodeToString(hash), client.RewordAddress.ToReadable())
	if client.Worker != nil {
		client.Worker.RealtimeWorkSubmitCount += 1
		powvalue := x16rs.CalculateHashPowerValue(hash)
		// fmt.Printf("addMinerPowerValue: %s, powvalue: %d\n", hex.EncodeToString(hash), powvalue)
		client.Worker.RealtimePower = client.Worker.RealtimePower.Add(client.Worker.RealtimePower, powvalue)
		return powvalue
	} else {
		return big.NewInt(0)
	}
}

// 注册
func (ps *PoolState) register(client *Client, addr *fields.Address) {
	// 分配 client id
	client.RewordAddress = addr
	ps.ClientCount += 1
	// 获取 worker
	wkr := ps.pool.getThePowWorker(addr)
	wkr.ClientCount += 1 // 统计 +1
	client.Worker = wkr  // worker set
	// 连接放入内存
	ps.AllClients.Store(client.Id, client)
	// 注册之后，发送挖矿信息
	ps.sendMiningStuffData(client)
}

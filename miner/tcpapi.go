package miner

import (
	"bufio"
	hex "encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/miner/difficulty"
	"github.com/hacash/x16rs"
	"io"
	"log"
	"math/big"
	"net"
	"time"
)

// 启动 tcp
func (mp *MiningPool) startApiListen() error {

	listenAndServe(mp, ":3339")

	return nil
}

// 发送挖矿信息
func (mp *MiningPool) sendMiningStuffToAllNode(ncb *NewCreateBlock) error {
	mp.AllActiveConns.Range(func(key interface{}, val interface{}) bool {
		client := key.(*Client)
		mp.sendMiningStuff(client, ncb)
		//client.Close()
		return true
	})

	return nil
}

// 发送挖矿信息
func (mp *MiningPool) sendMiningStuff(client *Client, ncb *NewCreateBlock) error {
	// 标号自增
	mp.autoincrementMiningCoinbaseStuffNum++
	miningNum := mp.autoincrementMiningCoinbaseStuffNum
	client.MiningCoinbaseStuffNum = miningNum
	// 重新计算区块 hash
	mp.fillCoinbaseMsg(ncb.Coinbase, miningNum)
	ncb.Block.SetMrklRoot(blocks.CalculateMrklRoot(ncb.Block.GetTransactions()))
	mlstuff := blocks.CalculateBlockHashBaseStuff(ncb.Block)
	// 发送消息
	hei := ncb.Block.GetHeight()
	loopnum := hei/50000 + 1
	if loopnum > 16 {
		loopnum = 16
	}
	stuffmining := &x16rs.MiningPoolStuff{
		BlockHeight:   hei,
		MiningIndex:   miningNum,
		Loopnum:       uint8(loopnum),
		TargetHash:    difficulty.Uint32ToHash(hei, ncb.Block.GetDifficulty()),
		BlockHeadMeta: mlstuff,
	}
	//
	// 信息发送给客户端
	x16rs.MiningPoolWriteTcpMsgBytes(client.Conn, 1, stuffmining.Serialize())
	return nil
}

// 定时移除非活跃地址
func (mp *MiningPool) removeDeadConn() {
	for {
		time.Sleep(time.Minute * 5)
		nnn := time.Now()
		mp.AllActiveConns.Range(func(key interface{}, val interface{}) bool {
			client := key.(*Client)
			if nnn.Unix()-client.ActiveTimestamp.Unix() > 7*60*1000 {
				client.Conn.Close()           // 关闭
				mp.AllActiveConns.Delete(key) // 移除
			}
			return true
		})
	}
}

//////////////////////////////////////////////////////

// 客户端连接的抽象
type Client struct {
	// tcp 连接
	Conn net.Conn
	// 当前使用的挖矿
	MiningCoinbaseStuffNum uint64
	// 奖励地址
	RewordAddress *fields.Address
	// 上次活跃的时间戳
	ActiveTimestamp time.Time
}

func listenAndServe(mp *MiningPool, address string) {
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
		go handle(mp, conn)
	}
}

func handle(mp *MiningPool, conn net.Conn) {
	if mp.currentPoolPeriodStateData == nil {
		return // 没准备好，下次重连
	}

	client := &Client{
		Conn:                   conn,
		MiningCoinbaseStuffNum: 0,
		RewordAddress:          nil,
		ActiveTimestamp:        time.Now(),
	}
	mp.AllActiveConns.Store(client, 1)
	// 使用 bufio 标准库提供的缓冲区功能
	reader := bufio.NewReader(conn)
	for {
		// ReadString 会一直阻塞直到遇到分隔符 '\n'
		// 遇到分隔符后会返回上次遇到分隔符或连接建立后收到的所有数据, 包括分隔符本身
		// 若在遇到分隔符之前遇到异常, ReadString 会返回已收到的数据和错误信息
		msgbytes, err := x16rs.MiningPoolReadTcpMsgBytes(reader)
		if err != nil {
			// 通常遇到的错误是连接中断或被关闭，用io.EOF表示
			if err == io.EOF {
				// log.Println("connection close")
				mp.AllActiveConns.Delete(conn)
			} else {
				log.Println(err)
			}
			return
		}

		if msgbytes[0] == 11 {
			// 心跳活跃时间
			client.ActiveTimestamp = time.Now()
			continue
		}

		//fmt.Println(msgbytes)
		//fmt.Println(string(msgbytes))

		// 注册奖励地址
		if msgbytes[0] == 0 {

			addr, err := fields.CheckReadableAddress(string(msgbytes[1:]))
			if err != nil {
				log.Println("connection error close.")
				mp.AllActiveConns.Delete(conn)
				return
			}
			// 判断是否已经满员
			if mp.currentPoolPeriodStateData.NodeNumber >= mp.addressMaxNum {
				if _, hasex := mp.currentPoolPeriodStateData.MiningPowerStatistics.Load(addr); !hasex {
					// 已经满员，返回错误信息
					x16rs.MiningPoolWriteTcpMsgBytes(client.Conn, 255, []byte(`Pool miner member has exceeded the maximum, Please contact the mine service provider\n矿池数量已经超出最大值，暂时无法连接挖矿，请联系你的矿池服务商`))
					return
				}
			}
			client.RewordAddress = addr
			log.Println("link reword address " + client.RewordAddress.ToReadable())
			// 首次发送挖矿信息
			if mp.currentNewCreateBlock != nil {
				mp.sendMiningStuff(client, mp.currentNewCreateBlock)
			}

		}

		if client.RewordAddress == nil {
			continue // 必须首先注册矿工地址
		}

		// 接受区块挖出消息
		if msgbytes[0] == 2 {

			var success x16rs.MiningSuccess
			success.Parse(msgbytes, 1)
			mp.gotSuccessBlock(client, success)

			// 传递算力统计
		} else if msgbytes[0] == 3 {

			//fmt.Println(len(msgbytes), msgbytes)

			if len(msgbytes) >= 1+1+89+4 {
				//fmt.Println("stuff.BlockHeadMeta", msgbytes[2:])
				stuff := msgbytes[2:91]
				nonceNum := len(msgbytes[91:]) / 4
				for i := 0; i < nonceNum; i++ {
					stuff[79] = msgbytes[91+i*4]
					stuff[80] = msgbytes[91+i*4+1]
					stuff[81] = msgbytes[91+i*4+2]
					stuff[82] = msgbytes[91+i*4+3]
					hash := blocks.CalculateBlockHashByStuff(int(msgbytes[1]), stuff)
					// 增加算力统计
					powval := addMinerPowerValue(hash, mp.currentPoolPeriodStateData, client)
					//fmt.Println(">>>>>>>>>>>>>>>>>>>>> msgbytes[0] == 3")
					//fmt.Println("addMinerPowerValue", client.RewordAddress.ToReadable(), powval.String(), hex.EncodeToString(hash))
				}
			}

		}

		//b := []byte(msg)
		// 将收到的信息发送给客户端
		//conn.Write(b)
	}
}

/////////////////////////////////////

func addMinerPowerValue(hash []byte, datacount *PoolPeriodStateData, client *Client) *big.Int {

	if datacount != nil {
		key := client.RewordAddress.ToReadable()
		powvalue := x16rs.CalculateHashPowerValue(hash)
		oldvalue, ld := datacount.MiningPowerStatistics.LoadOrStore(key, powvalue)
		if ld { // 增加
			powvaluetotal := new(big.Int).Add(powvalue, oldvalue.(*big.Int))
			datacount.MiningPowerStatistics.Store(key, powvaluetotal)
		} else {
			datacount.NodeNumber += 1 // 数量统计
		}
		// 总计
		datacount.TotalMiningPower = new(big.Int).Add(datacount.TotalMiningPower, powvalue)
		//fmt.Println(datacount.TotalMiningPower.String())
		return powvalue
	} else {
		return big.NewInt(0)
	}
}

package pool

import (
	"encoding/binary"
	"fmt"
)

var AddressStatisticsStoreItem_bytesize = 4 + 4 + 8 + 8 + 4

// 统计数据 条目
type AddressStatisticsStoreItem struct {
	// 磁盘记录
	// Address fields.Address
	FindBlocks              uint32 // 挖出的区块数量
	FindCoins               uint32 // 挖出的币数量
	CompleteRewards         uint64 // 已完成并打币的奖励  单位： ㄜ240  （10^8）
	DeservedRewards         uint64 // 应得但还没有打币的奖励  单位： ㄜ240  （10^8）
	PrevTransferBlockHeight uint32 // 上一次打币时的区块
}

func (elm *AddressStatisticsStoreItem) Serialize() ([]byte, error) {
	data := make([]byte, AddressStatisticsStoreItem_bytesize)
	binary.BigEndian.PutUint32(data, elm.FindBlocks)
	binary.BigEndian.PutUint32(data[4:], elm.FindCoins)
	binary.BigEndian.PutUint64(data[8:], elm.CompleteRewards)
	binary.BigEndian.PutUint64(data[16:], elm.DeservedRewards)
	binary.BigEndian.PutUint32(data[24:], elm.PrevTransferBlockHeight)
	// fmt.Println(data)
	return data, nil
}

func (elm *AddressStatisticsStoreItem) Parse(buf []byte, seek uint32) (uint32, error) {
	if len(buf)-int(seek) < AddressStatisticsStoreItem_bytesize {
		return 0, fmt.Errorf("buf length error")
	}
	elm.FindBlocks = binary.BigEndian.Uint32(buf[seek:])
	elm.FindCoins = binary.BigEndian.Uint32(buf[seek+4:])
	elm.CompleteRewards = binary.BigEndian.Uint64(buf[seek+8:])
	elm.DeservedRewards = binary.BigEndian.Uint64(buf[seek+16:])
	elm.PrevTransferBlockHeight = binary.BigEndian.Uint32(buf[seek+24:])
	return seek + uint32(AddressStatisticsStoreItem_bytesize), nil
}

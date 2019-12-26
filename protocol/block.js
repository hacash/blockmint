/**
 * 区块数据格式
 */

/*

用 交易hash难度值 防垃圾交易，并奖励代币

用 交易的总平均hash难度 作为，区块难度值的一部分

同一笔交易支持用不同的手续费，重复提交并覆盖较低手续费的交易

需要有各种不同的资产转移支付的方式（不仅仅是A转账给B），能模拟大部分的商业需求，并且是固定写入协议没有安全问题的

PoW 和 PoS 混合共识


5分钟一个区块， 每天12*24=288个区块，每周2016个区块， 每约5年=50万个区块， 调整一次挖矿数量
按黄金分割数列 (21+13+8+5+3+2+1+1 ... +1+1+1+1...) 第一期 2700万 个币，35年后挖矿稳定在一个币


代币名字叫做 Hacash 简称 HCS




*/



/*

设计确稿：

#. 交易也挖矿（交易费为负数时将获得代币！），从区块的总手续费 和 coinbase 里面扣取

#. 挖矿算法避免被 AISC 化（采用各种算法的组合，充分利用一个完整的PC机的各个部分整体性能）

#. 公平的代币分配方式，没有预挖矿，没有基金会，没有超级节点

#. 对侧链结算（闪电网络）、分布式信用有协议层的支持

#. 覆盖现有大部分金融支付方式的协议层支持，能模拟整个金融、股权、权益体系的运行

#. 形成一套区块协议层的标准（非盈利性技术标准委员会）

#. 区块钻石的收藏价值、价值储存体系（Math.pow("WTYUIAHXVMEKBSZN".length, 6) / 288 / 365 = 160年才能挖完钻石）

#. 支持海量小额交易低成本即时到账（闪电网络、高频结算网络）





*/



/*







*/





/**
 * 预定义的区块数据格式
 * 1 byte = 255
 * 2 byte = 65535
 * 3 byte = 16777215
 * 4 byte = 4294967295
 * 8 byte = 18446744073709551615
 */

const PredefinitionBlockFormat = {
    //   1 byte, 0~255, 版本号, 应该极度谨慎升级版本
    version: 1,
    //   5 byte, 区块高度, 一分钟一个块可用8100*255年；一秒一个块可用135*255年
    height: 0,
    //   5 byte, 一秒可用135*255年
    timestamp: 0,
    //   1 byte, 难度目标（用于PoW算力计算）  256个难度梯度
    // difficulty: 6,
    //   4 byte, 随机数（用于PoW算力计算）
    // nonce: 4294967295,
    //  32 byte, 本区块哈希  =  hash( version + height + difficulty + prevHash + mrklRoot + nonce )
    // hash: Buffer.alloc(32),
    //  32 byte, 父区块哈希
    prevMark: Buffer.alloc(32),
    //  32 byte, Merkle树根哈希
    mrklRoot: Buffer.alloc(32),

    // min byte = 2, max byte = 2 + 65535 * 4 = 8192 * 32
    extensionSize: 2,
    extensions: new Buffer(),

    // ----- 可选字段 ----- //
    //   4 byte, 难度目标（用于PoW算力计算）
    // difficulty: 4,
    //   4 byte, 随机数（用于PoW算力计算）
    // nonce: 4294967295,
    // ----- 可选END ----- //

    // 3 byte 交易数量
    transactionCount: 123,

    // 交易
    transactions: [ // length = 4 byte

        ////////  coinbase trs （可选的）  ////////

        {
            //   1 byte, 0~255, 交易类型
            type: 0, // type=0 为 区块 coinbase 奖励
            // 矿工获得的币，必须在200个区块(10个小时)之后，才能花出去
            // 也就是说：挖出区块后10小时内不能花钱，也不能再当矿工
            // 收取奖励的地址
            address: "29aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM",
            // 奖励数量
            ubill: { // length = 1 byte 数量 255 个以内
                //   3 byte, 0 ~ 16777215  数量 无符号整形
                amount: 21,
                //   1 byte, 0~255, 单位（后面跟了几个零）
                unit: 248, // 248 240 232 224 216 208 200
            },
            //  16 byte, 区块寄语（末尾用空格补齐）
            message: "hardertodobetter", // string
            //   5 byte, 区块开始挖掘的<bcc时间戳>（从创世到现在的秒数）
            // timestamp: 0,  // {不接受时间}
        },

        ////////  normal trs  ////////

        {
            //   1 byte, 0~255, 交易类型
            type: 1, // t=1 为 普通交易
            //   5 byte, 交易生成的<bcc时间戳>（从创世到现在的秒数）
            timestamp: 180,
            //  1 + 34 byte  【可选字段，默认单签名或多重签名列表的唯一一个，如果有多个则必须指定，用以区分手续费签名】  本交易签名地址， 默认地址 和 手续费地址
            address: "29aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM",

            // 手续费和附加信息 与 其他账户无关， 改变手续费和附加信息还是同一条交易

            // 手续费可以为负数，从总手续费和 coinbase 里扣除，达成交易的分布式挖矿，让每一笔交易也挖矿，并产生收益
            // 手续费支付者的签名包含了 fee 和 feeUnit 字段，而本交易hash不包含，实现手续费动态出价
            fee: {
                //   2 byte, -32768 ~ 32768  手续费数量（有符号整形）
                amount: 1234,
                //   1 byte, 0~255, 单位（后面跟了几个零）
                unit: 248, // 248 240 232 224 216 208 200
            },
            // 附录、附件、交易附加
            // min byte = 1, max byte = 1 + 255 * 4 = 16 * 64
            appendixSize: 2,
            appendixs: new Buffer(),


            // 1 byte 功能数量统计
            actionCount: 123,

            // 功能、资产列表
            actions: [ // length = 2 byte 数量 65536 个以内

                /////////////////  转账相关  /////////////////

                {
                    //   2 byte, 0~65535, 资产类型
                    kind: 1, // 普通转账
                    //  34 byte, 转账收款方（地址为33位则末尾为空格）
                    address: "1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD",
                    //   1 byte 账单数量统计
                    bill: { // length 数量 255 个以内
                        //   1 byte   1~127为正,   128~255为负
                        dist: 1,
                        //   1~127 byte, 0~16777216   4294967295   1095216660225, 转账金额数量
                        amount: Buffer.alloc(),
                        //   1 byte, 0~255, 转账单位（后面跟了几个零）
                        unit: 248,
                    },
                },
                {
                    kind: 2, // 请求对方转账给自己
                    billCount: 123,
                    bills: [{
                        amount: 1,
                        unit: 8,
                    }],
                    //  34 byte, 转账付款方
                    // 【需要在 signs 字段内加上本地址的签名】
                    address: "1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD",
                },
                {
                    kind: 3, // 请求 fromAddress 转账给 toAddress
                    billCount: 123,
                    bills: [{
                        unit: 8,
                        amount: 1,
                    }],
                    //  34 byte, 转账付款方
                    // 【需要在 signs 字段内加上本地址的签名】
                    fromAddress: "1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD",
                    //  34 byte, 转账收款方
                    toAddress: "1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD",
                },
                {
                    kind: 4, // UTXO 模式转账 请求 inputAddresses 的全部余额转移给 outputs
                    //  34 byte, 转账付款方
                    // 【需要在 signs 字段内加上本地址的签名】
                    // length = 1 byte
                    inputAddressCount: 123,
                    inputAddresses: ["1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD"], // 多个输入的地址
                    //  34 byte, 转账收款方
                    // 1 byte 数量统计
                    outputCount: 123,
                    outputs: [{ // length = 1 byte
                        address: "1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD",
                        billCount: 123,
                        bills: [{
                            unit: 8,
                            amount: 1,
                        }],
                    }],
                },
                {
                    kind: 5, // UTXO 模式转账 请求 inputs 里的转移给 outputs
                    //  34 byte, 转账付款方
                    // 【需要在 signs 字段内加上本地址的签名】
                    // length = 1 byte
                    // 1 byte 数量统计
                    inputCount: 123,
                    inputs: [{ // length = 1 byte
                        address: "1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD",
                        billCount: 123,
                        bills: [{
                            unit: 8,
                            amount: 1,
                        }], // 多个输入的对象
                    }],
                    // 1 byte 数量统计
                    outputCount: 123,
                    // 转账收款方
                    outputs: [{ // length = 1 byte
                        address: "1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD",
                        billCount: 123,
                        bills: [{
                            unit: 8,
                            amount: 1,
                        }],
                    }],
                },

                {
                    // 定额支付混搅
                    kind: 6,
                    fee: { // 每一个地址都要收取的服务手续费，可为零，也可为负
                        amount: 1234,
                        unit: 248,
                    },
                    bill: { // 统一转账数额
                        dist: 1,
                        amount: Buffer.alloc(),
                        unit: 248,
                    },
                    addressCount: 100, // 参与混搅地址数量
                    inputAddresses: ["1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD", "..."], // 多个输入的地址
                    outputAddresses: ["19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM", "..."], // 多个输出的地址
                },



                /////////////////  钻石挖出交易  /////////////////

                {
                    // 上一个钻石区块(或创世区块)的 hash + 自己的 publickey 为底 + 随机数，进行hash再转换成
                    // 类似 00000WTYUIA00000 的16位字符串（前后都是五位0）
                    // 一个区块最多仅含有一枚钻石
                    // 钻石字面值唯一不重复

                    // 钻石挖出声明
                    kind: 16,
                    //   6 byte, WTYUIAHXVMEKBSZN, 钻石字面值
                    diamond: "AAMMKK",
                    //   8 byte, 随机数 用于尝试生成钻石
                    nonce: 289237457,
                    // 钻石首次获得者
                    address: "19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM",
                },
                {
                    // 钻石交易转移（自己必须拥有）
                    kind: 17,
                    diamond: "WWUUYY",
                    // 收钻方
                    address: "19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM",
                },


                /////////////////  防攻击投票  /////////////////

                {
                    // 用来投票选举链条，让矿工强制转换到大家公认的链条上来
                    // 以通道锁定资金为票数，达到一定票数（矿工决定）才开启强制转换
                    // 以此纠正被 51%攻击的分叉，平时不启动，关键时刻才生效
                    // 只有一个月=8064（三个月=24192）个区块前的有效通道才有资格投票
                    kind: 9,
                    // 链条必须包含的区块哈希，一般是分叉起始块
                    // 必须是本链条历史上已经存在的hash
                    targetHash: Buffer.alloc(32),
                    // 通道数量
                    channelIdCount: 3,
                    // 参与投票的通道列表
                    channelIds:[
                        232353,  // 8 byte
                        3847658374,
                        874568376455,
                    ],
                },

                /////////////////  时间要求  /////////////////

                {
                    // 当前区块高度 必须小于等于 height 时，交易才有效。（交易就算签名了，也会在一定时间内过期）
                    kind: 6,
                    //   5 byte, 区块高度
                    height: 123,
                },
                {
                    // 当前区块高度 必须大于等于 height 时，交易才有效。（交易的锁定期，一定时间内暂时不能到账）
                    kind: 7,
                    //   5 byte, 区块高度
                    height: 123,
                },
                {
                    // 当前区块高度 必须 在一定区间内（可以等于） 时，交易才有效。交易生效区间
                    kind: 8,
                    //   5 byte, 区块高度
                    startHeight: 123,
                    endHeight: 789,
                },

                /////////////////  交易判断  /////////////////

                {
                    // 区块链历史上必须已存在目标交易，此比交易才能成功（交易后行）
                    kind: 9,
                    afterTransactionHash: Buffer.alloc(32),
                },
                {
                    // 区块链历史上必须排除、没有这一笔交易（交易先行）
                    kind: 10,
                    beforeTransactionHash: Buffer.alloc(32),
                },
                {
                    // 在一些交易生效之后，而在另一些交易生效之前，此交易可行
                    kind: 11,
                    // 1 + N*32 byte
                    afterTransactionHashs: [Buffer.alloc(32)],
                    // 1 + N*32 byte
                    beforeTransactionHashs: [Buffer.alloc(32)],
                },

                /////////////////  结算网络支持  /////////////////

                {
                    // 开启一条结算通道，fund1和fund2为出资双方
                    // 结算链一旦开启将扣除双方的余额，直到通道关闭
                    // 本交易的hash作为结算通道的 channelTransactionHash
                    kind: 12,
                    // 2 type 最高220天，结算通道单方终结的锁定期限（区块数量）
                    parteEndLockHeightNum: 2016, // 约等于于一周
                    // 8 type 结算通道自定义id
                    channelId: 232353253456,
                    fund1: {
                        address: '1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD',
                        billCount: 1,
                        bills: [{
                            unit: 8,
                            amount: 1234,
                        }],
                    },
                    fund2: {
                        address: '19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM',
                        billCount: 1,
                        bills: [{
                            unit: 8,
                            amount: 1234,
                        }],
                    },
                },

                {
                    // 批量开启结算通道， 以对等出资的方式，锁定期 一周 2016 个区块
                    kind: 13,
                    // 各自要锁定的通道金额
                    bill: {
                        unit: 8,
                        amount: 1234,
                    },
                    // 2 type 最高220天，结算通道单方终结的锁定期限（区块数量）
                    parteEndLockHeightNum: 2016, // 约等于于一周
                    // 1 type 通道数量
                    channelCount: 255,
                    // 通道
                    channels: [{
                        id: 232353253456,  // 8 type
                        address1: '1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD',
                        address2: '19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM',
                    }],
                },

                {
                    // 开启结算通道， 以单向结算出资（addr1单向支付给addr2）的方式，锁定期 两周 4038 个区块
                    kind: 14,
                    // address1 要单方面锁定的通道金额
                    bill: {
                        unit: 8,
                        amount: 1234,
                    },
                    // 2 type 最高220天，结算通道单方终结的锁定期限（区块数量）
                    parteEndLockHeightNum: 2016, // 约等于于一周
                    // 1 type 通道数量
                    channelCount: 255,
                    // 通道
                    channels: [{
                        id: 232353253456, // 8 type
                        address1: '1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD',
                        address2: '19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM',
                    }],
                },

                {
                    // 向通道补充资金 1+4+5+1+1+(2+8+4)+2+33+64
                    kind: 16,
                    channelId: 232353253456,
                    address: '1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD',
                    bill: {
                        unit: 8,
                        amount: 1234,
                    },
                },

                {
                    // 双方确认余额，关闭结算通道
                    kind: 15, // 最终确认双方的通道余额分配（立即生效，没有锁定期）
                    // 8 type 结算通道自定义id
                    channelId: 232353253456,
                    // 确认余额分配 diffConfirm 为两者相差余额额度
                    diffConfirm: {
                        address: '19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM',
                        billCount: 1,
                        bills: [{
                            unit: 8,
                            amount: 1234,
                        }],
                    },
                },


                {
                    // 开启多方清算体系
                    kind: 16,
                    // 8 type 结算id
                    liquidationId: 232353253456,
                    // 2 type 最高220天，结算单方终结的锁定期限（区块数量）
                    parteEndLockHeightNum: 2016, // 约等于于一周
                    // 1 type 数量必须大于等于3个
                    bankrollCount: 12,
                    bankrolls: [{
                        address: '19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM',
                        bill: {
                            unit: 8,
                            amount: 1234,
                        },
                    }],
                },

                /**********************  链下结算支持  **********************/

                {
                    // 【链下结算】阶段性余额分配确认
                    // 如果在链上提交阶段性余额分配，地址将被锁定约定的时间，
                    kind: 13,
                    // 8 type 结算通道自定义id
                    channelId: 232353253456,
                    //  上一笔（或多笔）确认的通道交易hash（首笔的话本字段为交易通道开启所在交易的hash）
                    prevTransactionHashs: Buffer.alloc(32),
                    // 8 type 通道交易序号，自动增量
                    autoincrement: 123135,
                    diffConfirm: {
                        side: 1, // 正差比方
                        bill: {
                            dict: 1,
                            amount: 1234,
                            unit: 8,
                        },
                    }
                },
                {
                    // 通道转账交易
                    kind: 14,
                    bills: [{
                        unit: 8,
                        amount: 1234,
                    }],
                    // 通道数量
                    channelCount: 4,
                    channels: [{
                        // 对账类型（1.实时对账  2.延迟对账）
                        type: 1,
                        // 资金流向
                        side: 1, // fund2 => fund1
                        // 8 type 结算通道自定义id
                        channelId: 232353253456,
                        //  上一笔双方确认的交易hash
                        prevTransactionHash: Buffer.alloc(32),
                        // 8 type 通道交易序号，自动增量
                        autoincrement: 123135,
                        // 通道手续费
                        fee: {
                            unit: 8,
                            amount: 1234,
                        },
                        // 本笔交易完成时的通道差额确认
                        diffConfirm: {
                            side: 1, // 资金正比额方
                            bill: {
                                dict: 1,
                                amount: 1234,
                                unit: 8,
                            },
                        },
                    },{
                        // 对账类型（1.实时对账  2.延迟对账）
                        type: 2,
                        // 资金流向
                        side: 2, // fund1 => fund2
                        // 8 type 结算通道id
                        channelId: 232353253456,
                        // 8 type 通道交易流水号，自动增量
                        autoincrement: 123135,
                        // 通道手续费
                        fee: {
                            unit: 8,
                            amount: 1234,
                        },
                    }],
                },


                /****************** 链下结算支持 END ******************/


                {
                    // 申请单方面结束状态通道， 首次申请时，将通道置于锁定状态
                    // 每个通道每一方只能申请一次
                    kind: 19,
                    // 8 type 结算通道自定义id
                    // 此处的 channelId 要与引用的 transaction 内的 channelId 一致
                    channelId: 232353253456,
                    // 提请人
                    address: '19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM',
                    // 引用的交易（ 只可能为 通道支付 和 阶段对账 两个交易 ）（不可递归）
                    transaction: {/****/}

                },


                {
                    // 【惩罚】单方面用历史（非最新）余额确认来结束锁定通道，仲裁结果为最终态
                    // 需要提供 transaction 的 nonce 值更大的通道交易或余额确认交易
                    // 如果另一方能提供更新的通道余额交易，那么将惩罚性的获取到对方账号下的所有余额（除了开启的其他通道内的额度）
                    kind: 20,
                    // 8 type 结算通道自定义id
                    // 此处的 channelId 要与引用的 transaction 内的 channelId 一致
                    channelId: 232353253456,
                    // 提请人
                    address: '1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD',
                    // 引用的交易（ 只可能为 通道支付 和 阶段对账 两种交易 ）（不可递归）
                    transaction: {/****/}

                },

                /////////////////  组合地址  /////////////////

                {
                    // 生成的地址以 3 版本号 开头
                    // 生成方式为 hash(trsTime validRightsRatio address1 rights1 address2 rights2 ... )

                    // 生成组合地址
                    kind: 6,
                    //    2 byte, 1~10000 满足票数有效比例（万分比）（必须等于或大于此万分比值即可操作账户）
                    validRightsRatio: 30,
                    // 组成列表
                    // 1 byte 数量统计
                    formCount: 123,
                    forms: [ // length = 1 byte 数量 200 个以内
                        // 投票权 和 受益权 可以不等比（即同股不同权）
                        {
                            address: '19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM',
                            //   4 byte, 0~4294967295, 权益数
                            rights: 1,
                            //   4 byte, 0~4294967295, 投票权
                            votes: 3,
                        },
                        {
                            // 成员可以为组合地址，条件是【组合地址已经注册】
                            address: '29aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM',
                            rights: 23465,
                            votes: 38347,
                        },
                        {
                            address: '19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM',
                            rights: 2,
                            votes: 3,
                        },
                    ]
                },
                {
                    // 交易发起者必须为 3 版本号开头的组合地址
                    // 组合地址，添加（或覆盖(改票数)）地址
                    kind: 7,
                    // 1 byte 数量统计
                    formCount: 123,
                    forms: [ // length = 1 byte 限制数量始终在 255 个以内
                        {
                            // 成员可以为组合地址，条件是【组合地址已经注册】
                            address: '19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM',
                            rights: 3,
                            votes: 3,
                        },
                    ]
                },
                {
                    // 交易发起者必须为 3 版本号开头的组合地址
                    // 组合地址，删除地址 （删除不需要抵押费用）
                    kind: 8,
                    // 1 byte 数量统计
                    formCount: 123,
                    forms: [ // length = 1 byte 限制数量始终在 255 个以内
                        {
                            address: '19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM',
                        },
                    ]
                },
                {
                    // 交易发起者必须为 3 版本号开头的组合地址
                    // 组合地址，修改生效票数
                    kind: 9,
                    // 要修改的投票生效比例，单位为万分之
                    validRightsRatio: 3456,
                },
                {
                    // 交易发起者必须为 3 版本号开头的组合地址
                    // 注销地址，提取并转移所有金额
                    // 注意：注销将删除矿工状态数据库内的股权关系的所有数据
                    //       将导致所有下游的决议无法产生，如果占股影响决策权的话
                    kind: 10,
                    // 锁定的账户保底金额，最终转移的地址
                    address: '29aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM',
                },
                {
                    // 提出账户保护模式申请，只要有现存的在控制树内的任意一个私钥即可，必须提供1%的锁定资金，锁定半年
                    // 此时账户即被锁定，如果另外的私钥也来申请保护模式，则保护失效
                    // 如此即保证所有剩余地址都同意申请保护模式，才能最终成功
                    // 如果锁定半年后，还是没有人解锁，就可以将资金提取到这个地址上来
                    // 如果期间有满足投票权的资金操作，则保护自动失效
                    // 如果期间内没有满足投票权的资金操作，则资金即可提取至这个地址
                    kind: 11,
                    // 资金的最终转出的地址
                    address: '29aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM',
                },
                {
                    // 提取此前已经锁定股权账户的资金
                    kind: 12,
                    // 原本被锁定的股权账户的地址
                    address: '29aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM',
                },

                /////////////////  数据操作  /////////////////

                {
                    kind: 10,
                    //  16 byte, like MD5 要存根、声明、签署的文本哈希（用于达成共识、签合同、交易内容等等）
                    hash: "00000000000000000000000000000000", // string buffer
                },
                {
                    kind: 11,
                    //  32 byte, SHA256 要存根、声明、签署的文本哈希（用于达成共识、签合同、交易内容等等）
                    hash: "0000000000000000000000000000000000000000000000000000000000000000", // string buffer
                },
                {
                    kind: 12,
                    //   4 byte, 要存储声明的数字
                    number: 12347890,
                },
                {
                    kind: 13,
                    //   8 byte, 要存储声明的数字
                    number: 12347890834765387,
                },
                {
                    kind: 13,
                    // 255 byte max -变长字段-, 申明的文本明文（UTF8编码），最长支持 255 byte
                    // 1 byte 数量统计
                    stringLength: 123,
                    string: "asda sdaggh fghdf hfgh 23556 df", // string
                },
                {
                    kind: 14,
                    // 65535 byte max -变长字段-, buffer数据
                    bufferLength: 123,
                    buffer: "03950954c4c5cf11ee1e1f08269efe13dc6e3af754e082b539f5ad243781180003950954c4c5cf11ee1e1f08269efe13dc6e3af754e082b539f5ad243781180003950954c4c5cf11ee1e1f08269efe13dc6e3af754e082b539f5ad243781180003950954c4c5cf11ee1e1f08269efe13dc6e3af754e082b539f5ad2437811800", // buffer
                },
                {
                    kind: 15,
                    // 255 byte max -变长字段-, 申明的 title
                    titleLength: 123,
                    title: "asda sdaggh fghdf hfgh 23556 df", // string
                    // 65535 byte max -变长字段-, buffer数据
                    contentLength: 123,
                    content: "03950954c4c5cf11ee1e1f08269efe13dc6e3af754e082b539f5ad243781180003950954c4c5cf11ee1e1f08269efe13dc6e3af754e082b539f5ad243781180003950954c4c5cf11ee1e1f08269efe13dc6e3af754e082b539f5ad243781180003950954c4c5cf11ee1e1f08269efe13dc6e3af754e082b539f5ad2437811800", // buffer
                },

                /////////////////  利益分配  /////////////////

                {
                    // 分红（将金额按 rights 比例分配给成员账户，余数小数部分将存留在组合地址）
                    kind: 15,
                    billCount: 123,
                    bills: [{
                        unit: 8,
                        amount: 1,
                    }],
                },

                /////////////////  签名支持  /////////////////

                {
                    // 请求对方对本交易联合签名（合同签署，信息证明等）
                    kind: 21,
                    address: "19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM",
                },


                /////////////////  身份认证  /////////////////

                {
                    // 声明自己是谁
                    kind: 21,
                    name: "XXXX", //  255 byte 变长 名称
                    identity: "1234567890000123", //  255 byte 变长  身份id标识
                },
                {
                    // 指出、作证别人是谁
                    kind: 22,
                    address: "19aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM",
                    name: "XXXX", //  255 byte 变长 名称
                    identity: "1234567890000123", //  255 byte 变长  身份id标识

                }


            ],


            // 2 byte 签名数量统计
            signCount: 123,
            // 签名列表
            signs: [ // length = 2 byte
                {
                    //  33 byte, 公钥
                    publicKey: "0000000000000000000000000000000000000", // string buffer
                    //  64 byte, 签名
                    signature: "000000000000000000000000000000000000000000000000000000000000000000000000", // string buffer
                }
            ],


            // 多重签名表
            multisignCount: 123,
            multisigns: [ // length = 1 byte
                {
                    publicKeyScript: Buffer.from("23xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx0000000000000000000000000000000000000kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk"),
                    signatureScript: Buffer.from("12000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
                }
            ],

        }

    ],



}




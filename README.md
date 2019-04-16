# Hacash：一种大规模支付实时结算的加密货币系统

Hacash是一种运用通道链有序多签名实时冲销结算方法，可无上限扩充每秒交易量，能掠夺性惩罚不诚实一方从而确保资金安全和实时到账的加密电子货币发行支付及金融体系。内建复合签名地址和层级股权控制地址、多方签署交易结构、全类别支付协议及资产变更协议，满足现代金融、企业及个人的绝大部分支付需求。结合工作量证明、资金抵押及分叉投票的抢占式记账权奖励分配方式，能有效规避双重支付、防范算力攻击和降低马太效应。并采取符合经济规律的货币发行、公账及通道手续费、通道利息和区块钻石等激励机制，无需信任委托于任何机构即可维持整个系统长期有效运行。

通道链结算网络的基本原理是：每两个账户各自锁定若干资金，组成一个支付通道，私下可双方签名多次支付，期间无需向全网广播确认交易，只需最后一次向主网提交最终的余额分配即可取回各自正确拥有的资金，从而极大地自由扩充整个系统的每秒交易数量；如果单方结束通道，则其资金会被锁定一段时间，如果另一方在此期间内向主网举证更新的余额分配而证实对方作假，则揭露方将会夺取对方全部资金，从而迫使双方保持诚实；只需将多个支付通道连接起来，并从收款方开始让所有资金流转方有序签名直至付款者最后签署，则所有相关方都将同时收到和支出钱款，从而确保支付完整、实时到账和资金安全；通道可收取微量手续费激励其提供稳定服务。

相关链接：[技术白皮书](https://github.com/hacash/paper/blob/master/draft/whitepaper.cn.md)


### 运行挖矿节点

##### 一、准备
  1. 一台 CentOS 7.6 服务器，并可 root 账户登录
  2. 安装 Golang 语言 1.12 版本，GOPATH 设置为 `/root/go`
  3. 安装 git 版本控制器 2.21.0 版本

##### 二、下载代码并修改配置

  1. 下载代码，执行命令：
  ```
  mkdir -p /root/go/src/github.com/hacash
  cd /root/go/src/github.com/hacash
  git clone https://github.com/hacash/blockmint.git
  ```
  
  2. 拷贝并修改配置文件，执行命令：
  ```
  cd blockmint/
  cp config.example.yml hacash.config.yml
  vim hacash.config.yml
  ```
  修改 hacash.config.yml 文件中 miner.rewards 下的地址（与比特币压缩地址一致），作为挖矿奖励的收款账户；修改 hacash.config.yml 文件中 p2p.myname 为你自定义的节点识别名称（不支持中文）。
  
  3. 下载依赖、编译代码，并运行：
  ```
  chmod 777 goget.sh restartnode.sh
  ./restartnode.sh
  ```
  如此便成功运行了挖矿节点。使用 `tail -100 output.log` 可查看挖矿日志输出。如果有代码更新，则执行 `./restartnode.sh` 将自动下载依赖，并重新编译后运行。
  
### 进行转账

运行了挖矿节点（或者按上述步骤安装软件环境和下载代码后），可进行转账交易：

  1. 下载代码，执行命令：
  ```
  mkdir -p /root/go/src/github.com/hacash
  cd /root/go/src/github.com/hacash
  git clone https://github.com/hacash/blockmint.git
  ```
  编译代码，生成创建交易的工具：
  ```
  cd blockmint/
  ./goget.sh
  cp ../x16rs/libx16rs_hash.a ./
  go build -o toolshell_hacash run/toolshell/main/main.go
  ```
  编译完成，运行工具，并发送交易
  ```
  ./toolshell_hacash    
  ```
  运行以上编译生成的可执行文件，即进入交互式命令行工具，将显示一下欢迎和帮助信息，并显示读取输入的提示符`>>`。通过在其中输入一些简单的命令，即可完成创建账户、创建交易和发送交易等事项：
  ```
  >>passwd XXXXXXXXXX  ##  passwd 指令表示通过密码（或助记词，不包含空格）生成一个账户地址，并登录至工具中，可用于后续签名交易
  ```
  输出结果：
  ```
  Ok, has loaded your account private key: 0x5b09369749b5240d619e70883c4c89030708917c1b2f5f81e2dc1094c451fff9 address: 12ra7bS1ZGLbXSApbbPPbWmX7jaMe4ajcE
  ## 0x5b09369749b5240d619e70883c4c89030708917c1b2f5f81e2dc1094c451fff9 即为通过密码 XXXXXXXXXX 生成的账户私钥（用于签名交易、付款），12ra7bS1ZGLbXSApbbPPbWmX7jaMe4ajcE 为你的账户地址（用于付款或收款）
  ```
  生成一笔交易：
  ```
  >>gentx sendcash 12ra7bS1ZGLbXSApbbPPbWmX7jaMe4ajcE 14qc3pDBYL43Q5HvKag32gqhgffRLSudg8 HCX1:248 HCX1:244 
  ```
  输出结果为：
  ```
  transaction create success! 
  hash: <f78852cc862c57d11f5b4e9d1be1adba0a68476a4fe9e7f2e0c273051f39e718>, hash_with_fee: <6287fd40304b1ef78b2e70be362a4031259dfefc4a64ff93ee6b7b53c18bc203>
  body length 159 bytes, hex body is:
  -------- TRANSACTION BODY START --------
  01005c88a9f5001458265ac2630f4e7ebf36032f72312ebe2fe735f4010100010001002a1999bd5a61eb7802d2c6549d851bd8d52f2d6ff80101000103256dd3294096dadcc2959031d46aa6163cfd73873e0dc27d2a147718e6d21e9de7ed3219130079635db1c2114a2f06d7ca79103ed551a3adb3d6b8ac2eb955d96aa28241e9644cfacbb94c339779969b0078c203a15081236bf78d8d220bdba10000
  -------- TRANSACTION BODY END   --------
  ## 以上命令执行结果代表生成了一笔交易（交易hash为f78852cc862c57d11f5b4e9d1be1adba0a68476a4fe9e7f2e0c273051f39e718）并自动用账户12ra7bS1ZGLbXSApbbPPbWmX7jaMe4ajcE签名了交易。
  ## HCX1:248 代表你将要转账给地址 14qc3pDBYL43Q5HvKag32gqhgffRLSudg8 的金额（一枚HCX币），而后面的 HCX1:244 表示将要支付给矿工的手续费（万分之一枚HCX币，具体记账规则详见技术白皮书）
  ```
  以上 `gentx sendcash` 命令表示创建一笔发送现金的交易，12ra7bS1ZGLbXSApbbPPbWmX7jaMe4ajcE 为付款账户，14qc3pDBYL43Q5HvKag32gqhgffRLSudg8 为收款账户， HCX1:248 为付款数额， HCX1:244 为手续费数量。 生成的交易hash为：f78852cc862c57d11f5b4e9d1be1adba0a68476a4fe9e7f2e0c273051f39e718，现在你可以向矿工发送这笔交易：
  ```
  >>sendtx f78852cc862c57d11f5b4e9d1be1adba0a68476a4fe9e7f2e0c273051f39e718 hacash.org:3338
  ```
  输出结果为：
  ```
  add tx to hacash.org:3338, the response is:
  Transaction f78852cc862c57d11f5b4e9d1be1adba0a68476a4fe9e7f2e0c273051f39e718 Add to MemTxPool success !
  ```
  以上信息表示交易已经发送给矿工 hacash.org 端口为 3338（也可以采用IP+PORT形式，例如 47.244.26.14:3338 ）
  如果账户 12ra7bS1ZGLbXSApbbPPbWmX7jaMe4ajcE 内拥有充足余额的话，这笔转账交易将被确认并打包、广播验证。如果余额不足，则交易将被简单地忽略丢弃。
  ```
  >>exit  ## 退出、关闭交易生成工具
  ```
  toolshell 为 Hacash 的专有交易创建、签名、发送的工具，类似钱包。更多功能及命令请参见文档。
  转账后，可以通过数据接口查看账户余额，访问： http://hacash.org:3338/query?action=balance&address=14qc3pDBYL43Q5HvKag32gqhgffRLSudg8 ，结果为
  ```
  {
    amount: "ㄜ0:0"   ## 表示余额为零
  }
  ```




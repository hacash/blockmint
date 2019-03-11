# Hacash node


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

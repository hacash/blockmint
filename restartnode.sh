#!/bin/bash

./goget.sh

cp ../x16rs/libx16rs_hash.a ./

go build -o miner_node_hacash run/miner/main/main.go

# kill
kill -s 9 `ps -aux | grep miner_node_hacash | awk '{print $2}'`

nohup ./miner_node_hacash > output.log &

# finish
echo "tail -50 output.log # to see the logs"


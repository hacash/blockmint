#!/bin/bash

cd /root/go/src/github.com/hacash/blockmint

./goget.sh

cp ../x16rs/libx16rs_hash.a ./

go build -o miner_node_hacash run/miner/main/main.go

# kill
ps -ef | grep miner_node_hacash | grep -v grep | awk '{print $2}' | xargs --no-run-if-empty kill

# start
nohup ./miner_node_hacash > output.log &

# finish
echo "tail -50 output.log # to see the logs"


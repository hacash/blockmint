#!/bin/bash


cp ../x16rs/libx16rs_hash.a ./

go build run/miner/main/main.go -o ./hacash_node


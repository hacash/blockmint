#!/bin/bash

# go get all

# p2p
go get -x github.com/ethereum/go-ethereum/p2p

# config
go get -x github.com/jinzhu/configor

# x16rs
go get -x github.com/hacash/x16rs

# hash
# go get -x golang.org/x/crypto/ripemd160
# go get -x golang.org/x/crypto/sha3

# bitcoin
go get -x github.com/hacash/bitcoin

# database
go get -x github.com/golang/snappy
go get -x github.com/syndtr/goleveldb

# gpu miner
go get -x github.com/samuel/go-opencl

# utils
go get -x github.com/deckarep/golang-set

# finish
echo "all pkg go get finish."


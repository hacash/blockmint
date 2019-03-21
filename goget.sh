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

# utils
go get -x github.com/deckarep/golang-set

# finish
echo "all pkg go get finish."


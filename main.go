package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"
)

type Blockchain struct {
	blocks []*Block
}

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
}

func (blockchain *Blockchain) AddBlock(data string) {
	tipBlock := blockchain.blocks[len(blockchain.blocks)-1]
	newBlock := NewBlock(data, tipBlock.Hash)
	blockchain.blocks = append(blockchain.blocks, newBlock)
}

func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewBlock("Genesis Block", []byte{})}}
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}}
	block.SetHash()
	return block
}

func (block *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(block.Timestamp, 10))
	contents := bytes.Join([][]byte{timestamp, block.Data, block.PrevBlockHash}, []byte{})
	hash := sha256.Sum256(contents)
	block.Hash = hash[:]
}

func main() {
	blockchain := NewBlockchain()

	blockchain.AddBlock("Send 50 BTC to Satoshi")
	blockchain.AddBlock("Send 25 more BTC to Nick Szabo")
	blockchain.AddBlock("Send 12 more BTC to Luke Jones")

	for _, block := range blockchain.blocks {
		fmt.Printf("Prev: %x\n", block.PrevBlockHash)
		fmt.Printf("Time: %d\n", block.Timestamp)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}

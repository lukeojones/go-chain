package main

type Blockchain struct {
	blocks []*Block
}

func (blockchain *Blockchain) AddBlock(data string) {
	tipBlock := blockchain.blocks[len(blockchain.blocks)-1]
	newBlock := NewBlock(data, tipBlock.Hash)
	blockchain.blocks = append(blockchain.blocks, newBlock)
}

func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewBlock("Genesis Block", []byte{})}}
}

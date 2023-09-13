package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"time"
)

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func (block *Block) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	if err := encoder.Encode(block); err != nil {
		panic("Unable to serialize Block")
	}

	return buffer.Bytes()
}

// HashTransactions Produce a single hash representing all transactions in the Block
// The final hash is the hash of the concatenated transaction hashes (IDs).
// NB: Bitcoin is more sophisticated than this and uses Merkle Trees.
func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte
	for _, tx := range block.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	combinedTxHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return combinedTxHash[:]
}

func DeserializeBlock(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&block); err != nil {
		fmt.Println(err)
		panic("Unable to deserialize data")
	}
	return &block
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce
	return block
}

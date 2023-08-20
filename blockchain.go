package main

import (
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"io/fs"
	"os"
)

const dbFile = "blockchain.db"
const blocksBucketName = "blocks"
const genesisData = "Hello Blockchain!"

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (iterator *BlockchainIterator) Next() *Block {
	var block *Block
	// retrieve block
	iterator.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucketName))
		blockBytes := bucket.Get(iterator.currentHash)
		block = DeserializeBlock(blockBytes)
		return nil
	})
	iterator.currentHash = block.PrevBlockHash
	return block
}

func (blockchain *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{blockchain.tip, blockchain.db}
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return true
}

func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, _ := bolt.Open(dbFile, 0600, nil)
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucketName))
		if bucket != nil {
			tip = bucket.Get([]byte("l"))
		} else {
			println("Creating Coinbase Tx")
			coinbaseTx := NewCoinbaseTx(address, genesisData)
			genesisBlock := NewGenesisBlock(coinbaseTx)
			bucket, _ := tx.CreateBucket([]byte(blocksBucketName))
			bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
			bucket.Put([]byte("l"), genesisBlock.Hash)
			tip = genesisBlock.Hash
		}
		return nil
	})
	return &Blockchain{tip, db}
}

func NewBlockchain() *Blockchain {
	// Open the DB
	//Create an update transaction
	//Read from the block bucket
	//If bucket exists, read last hash ("l") value and assign to tip
	//If not exists, create bucket, create genesis block, put genesis block, put genesis hash @ l and assign to tip
	var tip []byte
	db, _ := bolt.Open(dbFile, 0600, nil)
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucketName))
		if bucket != nil {
			tip = bucket.Get([]byte("l"))
		} else {
			coinbaseTx := NewCoinbaseTx("Lukoshi", "Hello Blockchain!")
			genesisBlock := NewGenesisBlock(coinbaseTx)
			bucket, _ := tx.CreateBucket([]byte(blocksBucketName))
			bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
			bucket.Put([]byte("l"), genesisBlock.Hash)
			tip = genesisBlock.Hash
		}
		return nil
	})
	return &Blockchain{tip, db}
}

func NewGenesisBlock(coinbaseTx *Transaction) *Block {
	return NewBlock([]*Transaction{coinbaseTx}, []byte{})
}

package main

import "github.com/boltdb/bolt"

const dbFile = "blockchain.db"
const blocksBucketName = "blocks"

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

func (blockchain *Blockchain) AddBlock(data string) {
	newBlock := NewBlock(data, blockchain.tip)
	blockchain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucketName))
		bucket.Put(newBlock.Hash, newBlock.Serialize())
		bucket.Put([]byte("l"), newBlock.Hash)
		blockchain.tip = newBlock.Hash

		return nil
	})
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
			genesisBlock := NewBlock("Genesis Block", []byte{})
			bucket, _ := tx.CreateBucket([]byte(blocksBucketName))
			bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
			bucket.Put([]byte("l"), genesisBlock.Hash)
			tip = genesisBlock.Hash
		}
		return nil
	})
	return &Blockchain{tip, db}
}

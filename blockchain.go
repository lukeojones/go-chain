package main

import (
	"encoding/hex"
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

func containsIndex(indexToFind int, indices []int) bool {
	if indices == nil {
		return false
	}
	for _, index := range indices {
		if index == indexToFind {
			return true
		}
	}
	return false
}

func (blockchain *Blockchain) FindTxsWithUnspentOutputs(address string) []Transaction {
	var txsWithUtxos []Transaction
	spentTxos := make(map[string][]int) // txid -> []offset

	// Blocks
	bci := blockchain.Iterator()
	for {
		block := bci.Next()

		//Transactions
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

			// Transaction Outputs
			for txoIndex, txo := range tx.Outputs {
				spentOutputIndices := spentTxos[txID]
				if containsIndex(txoIndex, spentOutputIndices) {
					continue
				}
				//if spentOutputIndices != nil {
				//	for _, indexOfSpentOutput := range spentOutputIndices {
				//		if txoIndex == indexOfSpentOutput {
				//			// go to next transaction output
				//		}
				//	}
				//}

				// if here means there is a transaction output that isn't spent yet
				// so we need to check if it for our address/key
				if txo.CanBeUnlockedWith(address) {
					txsWithUtxos = append(txsWithUtxos, *tx)
				}
			}

			// now inspect the inputs of the block to mark spent outputs
			// coinbase can be ignored because they reference no inputs
			if tx.IsCoinbase() == false {
				for _, input := range tx.Inputs {
					if input.CanUnlockOutputWith(address) {
						outputTxID := hex.EncodeToString(input.TxOutputID)
						spentTxos[outputTxID] = append(spentTxos[outputTxID], input.TxOutputIndex)
					}
				}
			}
		}
	}
	// Loop backwards (newest first) through the blockchain
	// 		For every transaction (tx) in the block
	// 			For every transaction output (txo) in the tx
	// 			See if it has been spent (check map of spent txns by txid)
	// 			if it has,
	//				break
	// 			if it hasn't, check if the txo is unlockable by address/key string
	// 				if it is unlockable by the address, add to unspent txo (utxo) array
	// 		For every transaction input (txi) in the tx
	// 		Exclude coinbases
	// 		If the txi unlocks coins sent to address string
	// 			Add txo index to map (by txid) since you already have txid in the map key it's only necessary to store the tx offset

}

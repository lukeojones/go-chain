package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"golang.org/x/exp/slices"
	"io/fs"
	"log"
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

func (blockchain *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte
	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucketName))
		lastHash = bucket.Get([]byte("l"))
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)
	blockchain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucketName))
		err = bucket.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = bucket.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		blockchain.tip = newBlock.Hash
		return nil
	})
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
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucketName))
		if bucket != nil {
			tip = bucket.Get([]byte("l"))
		} else {
			println("Creating Coinbase Tx")
			coinbaseTx := NewCoinbaseTx(address, genesisData)
			genesisBlock := NewGenesisBlock(coinbaseTx)
			bucket, err := tx.CreateBucket([]byte(blocksBucketName))

			if err != nil {
				panic(err)
			}

			err = bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
			if err != nil {
				panic(err)
			}

			err = bucket.Put([]byte("l"), genesisBlock.Hash)
			if err != nil {
				panic(err)
			}

			tip = genesisBlock.Hash
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return &Blockchain{tip, db}
}

func NewBlockchain(address string) *Blockchain {
	var tip []byte
	db, _ := bolt.Open(dbFile, 0600, nil)
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucketName))
		if bucket != nil {
			tip = bucket.Get([]byte("l"))
		} else {
			coinbaseTx := NewCoinbaseTx(address, "Hello Blockchain!")
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
				if slices.Contains(spentOutputIndices, txoIndex) {
					continue
				}

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

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return txsWithUtxos
}

func (blockchain *Blockchain) FindUtxos(address string) []TxOutput {
	txsWithUtxos := blockchain.FindTxsWithUnspentOutputs(address)
	var utxos []TxOutput
	for _, tx := range txsWithUtxos {
		for _, txo := range tx.Outputs {
			if txo.CanBeUnlockedWith(address) {
				utxos = append(utxos, txo)
			}
		}
	}
	return utxos
}

func (blockchain *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentTxs := blockchain.FindTxsWithUnspentOutputs(address)
	spendableOutputs := make(map[string][]int)
	acc := 0

	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for offset, output := range tx.Outputs {
			if output.CanBeUnlockedWith(address) && acc < amount {
				acc = acc + output.Value
				spendableOutputs[txID] = append(spendableOutputs[txID], offset)

				if acc >= amount {
					return acc, spendableOutputs
				}
			}
		}
	}

	return acc, spendableOutputs
}

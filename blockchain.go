package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"golang.org/x/exp/slices"
	"io/fs"
	"log"
	"os"
)

const dbFile = "blockchain_%s.db"
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
	err := iterator.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucketName))
		blockBytes := bucket.Get(iterator.currentHash)
		block = DeserializeBlock(blockBytes)
		return nil
	})
	if err != nil {
		return nil
	}
	iterator.currentHash = block.PrevBlockHash
	return block
}

func (blockchain *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{blockchain.tip, blockchain.db}
}

func (blockchain *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	// Verify the transaction before adding them to the block
	for _, tx := range transactions {
		if blockchain.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid Transaction!")
		}
	}

	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucketName))
		lastHash = bucket.Get([]byte("l"))
		blockData := bucket.Get(lastHash)
		block := DeserializeBlock(blockData)
		lastHeight = block.Height
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash, lastHeight+1)
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

	return newBlock
}

func (blockchain *Blockchain) GetBestHeight() int {
	var lastBlock Block
	err := blockchain.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucketName))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return lastBlock.Height
}

func (blockchain *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block
	err := blockchain.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucketName))
		blockData := b.Get(blockHash)
		if blockData == nil {
			return errors.New("Block is not found")
		}
		block = *DeserializeBlock(blockData)
		return nil
	})
	if err != nil {
		return block, err
	}
	return block, nil
}

func (blockchain *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := blockchain.Iterator()
	for {
		block := bci.Next()
		blocks = append(blocks, block.Hash)
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return blocks
}

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return true
}

func CreateBlockchain(address string, nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) {
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

func NewBlockchain(address string, nodeID string) *Blockchain {
	var tip []byte
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}
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

func (blockchain *Blockchain) AddBlock(block *Block) {
	err := blockchain.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucketName))
		blockInDb := b.Get(block.Hash)
		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			blockchain.tip = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (blockchain *Blockchain) FindTxsWithUnspentOutputs(pubKeyHash []byte) []Transaction {
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
				if txo.IsLockedWithKey(pubKeyHash) {
					txsWithUtxos = append(txsWithUtxos, *tx)
				}
			}

			// now inspect the inputs of the block to mark spent outputs
			// coinbase can be ignored because they reference no inputs
			if tx.IsCoinbase() == false {
				for _, input := range tx.Inputs {
					if input.UsesKey(pubKeyHash) {
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

func (blockchain *Blockchain) BuildTransactionUtxoMap() map[string]TxOutputs {
	spentTxos := make(map[string][]int) // txid -> []offset
	utxoMap := make(map[string]TxOutputs)

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
				utxoForTx := utxoMap[txID]
				utxoForTx.Outputs = append(utxoForTx.Outputs, txo)
				utxoMap[txID] = utxoForTx
			}

			// now inspect the inputs of the block to mark spent outputs
			// coinbase can be ignored because they reference no inputs
			if tx.IsCoinbase() == false {
				for _, input := range tx.Inputs {
					outputTxID := hex.EncodeToString(input.TxOutputID)
					spentTxos[outputTxID] = append(spentTxos[outputTxID], input.TxOutputIndex)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return utxoMap
}

// FindTx iterates through all blocks to find the transaction with provided ID
func (blockchain *Blockchain) FindTx(ID []byte) (Transaction, error) {
	bci := blockchain.Iterator()
	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("transaction not found")
}

// SignTransaction takes a transaction, finds all transactions it references and signs it
func (blockchain *Blockchain) SignTransaction(tx *Transaction, key ecdsa.PrivateKey) {
	prevTxs := make(map[string]Transaction)
	for _, input := range tx.Inputs {
		prevTx, err := blockchain.FindTx(input.TxOutputID)
		if err != nil {
			log.Panic(err)
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}
	tx.Sign(key, prevTxs)
}

// VerifyTransaction takes a transaction, finds all transactions it references and verifies the signature
func (blockchain *Blockchain) VerifyTransaction(tx *Transaction) bool {
	prevTxs := make(map[string]Transaction)
	for _, input := range tx.Inputs {
		prevTx, err := blockchain.FindTx(input.TxOutputID)
		if err != nil {
			log.Panic(err)
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}
	return tx.Verify(prevTxs)
}

package main

import (
	"encoding/hex"
	"github.com/boltdb/bolt"
	"log"
)

const utxoBucketName = "chainstate"

type UTXOSet struct {
	Blockchain *Blockchain
}

func (us UTXOSet) Reindex() {
	db := us.Blockchain.db
	bucketName := []byte(utxoBucketName)

	db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket(bucketName)
		_, _ = tx.CreateBucket(bucketName)
		return nil
	})

	utxoMap := us.Blockchain.BuildTransactionUtxoMap()

	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		for txID, utxos := range utxoMap {
			key, _ := hex.DecodeString(txID)
			bucket.Put(key, utxos.Serialize())
		}
		return nil
	})
}

func (us UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	spendableOutputs := make(map[string][]int)
	acc := 0

	db := us.Blockchain.db
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucketName))
		cursor := bucket.Cursor()

		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			txID := hex.EncodeToString(key)
			outputs := DeserializeOutputs(value)

			for offset, output := range outputs.Outputs {
				if output.IsLockedWithKey(pubKeyHash) && acc < amount {
					acc = acc + output.Value
					spendableOutputs[txID] = append(spendableOutputs[txID], offset)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return acc, spendableOutputs
}
func (us UTXOSet) FindUtxos(pubKeyHash []byte) []TxOutput {
	var utxos []TxOutput
	db := us.Blockchain.db
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucketName))
		cursor := bucket.Cursor()

		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			outputs := DeserializeOutputs(value)

			for _, output := range outputs.Outputs {
				if output.IsLockedWithKey(pubKeyHash) {
					utxos = append(utxos, output)
				}
			}
		}
		return nil
	})

	return utxos
}

// Update Having the UTXO set means that our data (transactions) are now split into two storages:
//
//	actual transactions are stored in the blockchain,
//	and unspent outputs are stored in the UTXO set.
//
// Such separation requires solid synchronization mechanism because we want the UTXO set to
// always be updated and store outputs of most recent transactions
func (us UTXOSet) Update(block *Block) {
	db := us.Blockchain.db

	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucketName))
		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, input := range tx.Inputs {
					var updatedOutputs TxOutputs
					data := bucket.Get(input.TxOutputID)
					outputs := DeserializeOutputs(data)

					// carry over all outputs that have not been referenced by an input
					for offset, output := range outputs.Outputs {
						if input.TxOutputIndex != offset {
							updatedOutputs.Outputs = append(updatedOutputs.Outputs, output)
						}
					}

					// If there are no unspent outputs for a tx, then remove them from the utxo chainstate
					if len(updatedOutputs.Outputs) == 0 {
						bucket.Delete(input.TxOutputID)
					} else {
						bucket.Put(input.TxOutputID, updatedOutputs.Serialize())
					}
				}
			}

			// Now add the outputs from the latest tx (being added in this block)
			outputsForNewTx := TxOutputs{}
			for _, output := range tx.Outputs {
				outputsForNewTx.Outputs = append(outputsForNewTx.Outputs, output)
			}
			bucket.Put(tx.ID, outputsForNewTx.Serialize())
		}
		return nil
	})
}

package main

import (
	"encoding/hex"
	"github.com/boltdb/bolt"
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
	db.View(func(tx *bolt.Tx) error {
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

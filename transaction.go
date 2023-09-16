package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func (tx *Transaction) SetId() {
	var encoded bytes.Buffer
	encoder := gob.NewEncoder(&encoded)

	if err := encoder.Encode(tx); err != nil {
		log.Panic(err)
	}

	sum256 := sha256.Sum256(encoded.Bytes())
	tx.ID = sum256[:]
}

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].TxOutputID) == 0 && tx.Inputs[0].TxOutputIndex == -1
}

// NewCoinbaseTx Creates a Coinbase Transaction
// Use a dummy input since these coins are "mined" and have no origin transaction
// Dummy input can have arbirtrary data (like Satoshi's Chancellor data in first ever Coinbase)
// Output contains the block reward sent straight to the recipient (miner)
func NewCoinbaseTx(recipient, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coinbase Reward to: %s", recipient)
	}

	dummyTxInput := TxInput{[]byte{}, -1, nil, []byte(data)}
	//output := TxOutput{blockSubsidy, recipient}
	output := NewTXOutput(blockSubsidy, recipient)

	tx := Transaction{
		ID:      nil,
		Inputs:  []TxInput{dummyTxInput},
		Outputs: []TxOutput{*output},
	}

	tx.SetId()
	return &tx
}

func NewUtxoTransaction(from, to string, amount int, blockchain *Blockchain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}

	wallet := wallets.GetWallet(from)
	pubKeyHash := HashPubKey(wallet.PublicKey)
	available, spendableOutputs := blockchain.FindSpendableOutputs(pubKeyHash, amount)
	fmt.Printf("Found available funds of [%d] in [%s]\n", available, from)
	if available < amount {
		log.Panic("ERROR Not enough funds!")
	}

	for txID, outputIndices := range spendableOutputs { //outputs are offsets here (since we have the txID)
		txID, _ := hex.DecodeString(txID)
		for _, outputIndex := range outputIndices {
			input := TxInput{
				TxOutputID:    txID,
				TxOutputIndex: outputIndex,
				Signature:     nil,
				PubKey:        wallet.PublicKey,
			}
			inputs = append(inputs, input)
		}
	}

	// Build the outputs (one to receiver and one to sender as change)
	fmt.Printf("Creating main txo [%d to %s]\n", amount, to)
	outputs = append(outputs, *NewTXOutput(amount, to))

	if available > amount {
		fmt.Printf("Creating change txo [%d to %s]\n", available-amount, from)
		outputs = append(outputs, *NewTXOutput(available-amount, from))
	}

	tx := Transaction{
		ID:      nil,
		Inputs:  inputs,
		Outputs: outputs,
	}
	tx.SetId()
	return &tx
}

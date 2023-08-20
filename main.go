package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)

const blockSubsidy = 50

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxInput struct {
	TxID          []byte
	TxOutputIndex int
	ScriptSig     string
}

type TxOutput struct {
	Value        int
	ScriptPubKey string
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

// Creates a Coinbase Transaction
// Use a dummy input since these coins are "mined" and have no origin transaction
// Dummy input can have arbirtrary data (like Satoshi's Chancellor data in first ever Coinbase)
// Output contains the block reward sent straight to the recipient (miner)
func NewCoinbaseTx(recipient, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coinbase Reward to: %s", recipient)
	}

	dummyTxInput := TxInput{[]byte{}, -1, data}
	output := TxOutput{blockSubsidy, recipient}

	tx := Transaction{
		ID:      nil,
		Inputs:  []TxInput{dummyTxInput},
		Outputs: []TxOutput{output},
	}

	tx.SetId()
	return &tx
}

func main() {
	cli := CLI{}
	cli.Run()
}

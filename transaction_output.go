package main

import "bytes"

type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// Sets the PubKeyHash to that of the recipient's address (pub key)
func (out *TxOutput) Lock(address []byte) {
	pubKeyHashRecipient := ConvertBase58BytesToPubKeyHash(address)
	out.PubKeyHash = pubKeyHashRecipient
}

// NewTXOutput create a new TXOutput and lock to recipients address
func NewTXOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}

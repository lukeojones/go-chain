package main

import "bytes"

type TxInput struct {
	TxOutputID    []byte
	TxOutputIndex int
	Signature     []byte
	PubKey        []byte
}

// UsesKey checks if transaction input was initiated by the provided address (pub key)
func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

package main

import (
	"bytes"
	"encoding/gob"
	"log"
)

type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

type TxOutputs struct {
	Outputs []TxOutput
}

func (outputs TxOutputs) Serialize() []byte {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	if err := encoder.Encode(outputs); err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func DeserializeOutputs(data []byte) TxOutputs {
	var outputs TxOutputs
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&outputs); err != nil {
		log.Panic(err)
	}
	return outputs
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// Lock Sets the PubKeyHash to that of the recipient's address (pub key)
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

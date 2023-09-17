package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	encoder := gob.NewEncoder(&encoded)

	if err := encoder.Encode(tx); err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
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

// Sign constructs a trimmed version of a transaction including just:
//   - the inputs
//   - the outputs referenced by the inputs
//   - the outputs
func (tx *Transaction) Sign(key ecdsa.PrivateKey, prevTxs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	txTrimmed := tx.TrimmedCopy()
	for index, input := range txTrimmed.Inputs {
		prevTx := prevTxs[hex.EncodeToString(input.TxOutputID)]
		txTrimmed.Inputs[index].Signature = nil                                         // Blank the signature - it's not needed in the current tx sig
		txTrimmed.Inputs[index].PubKey = prevTx.Outputs[input.TxOutputIndex].PubKeyHash // PubKeyHash of referenced output

		txTrimmed.ID = txTrimmed.Hash()
		txTrimmed.Inputs[index].PubKey = nil // reset so this doesn't affect further iterations

		r, s, err := ecdsa.Sign(rand.Reader, &key, txTrimmed.ID) // Get signature of the hashed transaction
		if err != nil {
			log.Panic(err)
		}

		sig := append(r.Bytes(), s.Bytes()...)
		tx.Inputs[index].Signature = sig // store the signature on the input
	}

}

// TrimmedCopy generates a lightweight version of a transction for signing purposes
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, input := range tx.Inputs {
		inputs = append(inputs, TxInput{
			TxOutputID:    input.TxOutputID,
			TxOutputIndex: input.TxOutputIndex,
			Signature:     nil, //blank the sig
			PubKey:        nil, //blank the pub key
		})
	}

	for _, output := range tx.Outputs {
		outputs = append(outputs, TxOutput{
			Value:      output.Value,
			PubKeyHash: output.PubKeyHash,
		})
	}

	return Transaction{
		ID:      tx.ID,
		Inputs:  tx.Inputs,
		Outputs: tx.Outputs,
	}
}

// Verify all inputs match the mandatory signed data
func (tx *Transaction) Verify(prevTxs map[string]Transaction) bool {
	txTrimmed := tx.TrimmedCopy()
	curve := elliptic.P256()

	for index, input := range tx.Inputs {
		// this is the same preparation we do in the signing
		prevTx := prevTxs[hex.EncodeToString(input.TxOutputID)]
		txTrimmed.Inputs[index].Signature = nil
		txTrimmed.Inputs[index].PubKey = prevTx.Outputs[input.TxOutputIndex].PubKeyHash
		txTrimmed.ID = txTrimmed.Hash()
		txTrimmed.Inputs[index].PubKey = nil

		// Extract the signature
		r := big.Int{}
		s := big.Int{}
		sigLen := len(input.Signature)
		r.SetBytes(input.Signature[:(sigLen / 2)])
		s.SetBytes(input.Signature[(sigLen / 2):])

		//Extract the public key parts
		x := big.Int{}
		y := big.Int{}
		keyLen := len(input.PubKey)
		x.SetBytes(input.PubKey[:(keyLen / 2)])
		y.SetBytes(input.PubKey[(keyLen / 2):])

		// Verify that the provided signature matches by verifying data (tx ID + pubkey)
		// If any input doesn't have valid signature, return false for the transaction
		pubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&pubKey, txTrimmed.ID, &r, &s) == false {
			return false
		}
	}

	return true
}

// Hash generates the SHA256 of the serialized transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
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

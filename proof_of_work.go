package main

import (
	"bytes"
	"crypto/sha256"
	"math/big"
)

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join([][]byte{
		Int64ToBytes(pow.block.Timestamp),
		pow.block.Data,
		pow.block.PrevBlockHash,
		Int64ToBytes(int64(nonce)),
		Int64ToBytes(int64(difficulty)),
	}, []byte{})
	return data
}

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, 256-difficulty)
	return &ProofOfWork{block, target}
}

func (pow *ProofOfWork) Run() (nonce int, solvedHash []byte) {
	var hashAsInt big.Int
	var hash [32]byte

	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashAsInt.SetBytes(hash[:])

		if hashAsInt.Cmp(pow.target) == -1 {
			break
		}
		nonce++
	}

	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() (valid bool) {
	var hashAsInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashAsInt.SetBytes(hash[:])

	valid = hashAsInt.Cmp(pow.target) == -1
	return valid
}

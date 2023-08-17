package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"
)

const difficulty = 16
const maxNonce = math.MaxInt64

type Blockchain struct {
	blocks []*Block
}

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, 256-difficulty)
	return &ProofOfWork{block, target}
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

func (blockchain *Blockchain) AddBlock(data string) {
	tipBlock := blockchain.blocks[len(blockchain.blocks)-1]
	newBlock := NewBlock(data, tipBlock.Hash)
	blockchain.blocks = append(blockchain.blocks, newBlock)
}

func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewBlock("Genesis Block", []byte{})}}
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce
	return block
}

func (block *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(block.Timestamp, 10))
	contents := bytes.Join([][]byte{timestamp, block.Data, block.PrevBlockHash}, []byte{})
	hash := sha256.Sum256(contents)
	block.Hash = hash[:]
}

func (pow *ProofOfWork) Validate() (valid bool) {
	var hashAsInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashAsInt.SetBytes(hash[:])

	valid = hashAsInt.Cmp(pow.target) == -1
	return valid
}

func main() {
	blockchain := NewBlockchain()

	blockchain.AddBlock("Send 50 BTC to Satoshi")
	blockchain.AddBlock("Send 25 more BTC to Nick Szabo")
	blockchain.AddBlock("Send 12 more BTC to Luke Jones")

	for _, block := range blockchain.blocks {
		fmt.Printf("Prev: %x\n", block.PrevBlockHash)
		fmt.Printf("Time: %d\n", block.Timestamp)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoWo: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}

/*
 * Utils
 */
func Int64ToBytes(i int64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, i)
	if err != nil {
		fmt.Println("Failed to write int64 to buffer:", err)
	}
	return buf.Bytes()
}

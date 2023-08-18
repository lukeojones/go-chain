package main

import (
	"fmt"
	"math"
	"strconv"
)

const difficulty = 16
const maxNonce = math.MaxInt64

func main() {
	blockchain := NewBlockchain()

	blockchain.AddBlock("Send 50 BTC to Satoshi")
	blockchain.AddBlock("Send 25 more BTC to Nick Szabo")
	blockchain.AddBlock("Send 12 more BTC to Luke Jones")

	it := blockchain.Iterator()
	for {
		block := it.Next()
		fmt.Printf("Prev: %x\n", block.PrevBlockHash)
		fmt.Printf("Time: %d\n", block.Timestamp)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoWo: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

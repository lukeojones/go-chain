package main

import (
	"fmt"
	"strconv"
)

func (cli *CLI) PrintChain(nodeID string) {
	// Just exec NewBlockChain here (which actually loads the thing)
	bc := NewBlockchain(nodeID)
	defer bc.db.Close()
	it := bc.Iterator()
	for {
		block := it.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

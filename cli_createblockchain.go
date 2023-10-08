package main

import "fmt"

func (cli *CLI) CreateBlockchain(address, nodeID string) {
	println("1. Creating Chain")
	blockchain := CreateBlockchain(address, nodeID)
	defer blockchain.db.Close()

	utxoSet := UTXOSet{blockchain}
	utxoSet.Reindex()
	fmt.Println("Blockchain Created")
}

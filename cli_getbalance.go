package main

import "fmt"

func (cli *CLI) GetBalance(address, nodeID string) {
	bc := NewBlockchain(nodeID)
	utxo := UTXOSet{bc}
	defer bc.db.Close()

	balance := 0
	utxos := utxo.FindUtxos(ConvertBase58AddressToPubKeyHash(address))

	for _, utxo := range utxos {
		balance = balance + utxo.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

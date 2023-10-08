package main

func (cli *CLI) reindexUTXO(nodeID string) {
	bc := NewBlockchain(nodeID)
	utxoSet := UTXOSet{bc}
	utxoSet.Reindex()
}

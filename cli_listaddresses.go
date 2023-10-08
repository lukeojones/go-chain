package main

import (
	"fmt"
	"log"
)

func (cli *CLI) listAddresses(nodeID string) {
	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(nodeID)
	addresses := wallet.GetAddress()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

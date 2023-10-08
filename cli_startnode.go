package main

import "fmt"

func (cli *CLI) startNode(nodeID, minerAddress string) {
	fmt.Printf("Starting node %s\n", nodeID)
	if len(minerAddress) > 0 {
		fmt.Printf("Mining is on. Address to receive rewards: %s\n", minerAddress)
	}
	//todo - StartServer
}

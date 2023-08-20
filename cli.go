package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type CLI struct {
	bc *Blockchain
}

func (cli *CLI) PrintChain() {
	it := cli.bc.Iterator()
	for {
		block := it.Next()
		fmt.Printf("Prev: %x\n", block.PrevBlockHash)
		fmt.Printf("Time: %d\n", block.Timestamp)
		fmt.Printf("Data: %s\n", block.Transactions)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoWo: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) CreateChain(address string) {
	blockchain := CreateBlockchain(address)
	blockchain.db.Close()
	fmt.Println("Blockchain Created")
}

func (cli *CLI) PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  createchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createChainCmd := flag.NewFlagSet("createchain", flag.ExitOnError)
	createChainAddress := createChainCmd.String("address", "", "The address to send genesis block reward to")

	switch os.Args[1] {
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createchain":
		err := createChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.PrintUsage()
		os.Exit(1)
	}

	if printChainCmd.Parsed() {
		cli.PrintChain()
	}

	if createChainCmd.Parsed() {
		if *createChainAddress == "" {
			createChainCmd.Usage()
			os.Exit(1)
		}
		cli.CreateChain(*createChainAddress)
	}
}

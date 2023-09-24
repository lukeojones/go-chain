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
	// Just exec NewBlockChain here (which actually loads the thing)
	bc := NewBlockchain("Luketoshi")
	defer bc.db.Close()
	it := bc.Iterator()
	for {
		block := it.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
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

func (cli *CLI) CreateChain(address string) {
	println("1. Creating Chain")
	blockchain := CreateBlockchain(address)
	defer blockchain.db.Close()

	utxoSet := UTXOSet{blockchain}
	utxoSet.Reindex()
	fmt.Println("Blockchain Created")
}

func (cli *CLI) PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  createchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
}

func (cli *CLI) GetBalance(address string) {
	bc := NewBlockchain(address)
	utxo := UTXOSet{bc}
	defer bc.db.Close()

	balance := 0
	utxos := utxo.FindUtxos(ConvertBase58AddressToPubKeyHash(address))

	for _, utxo := range utxos {
		balance = balance + utxo.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) Send(from string, to string, amount int) {
	blockchain := NewBlockchain(from)
	defer blockchain.db.Close()
	utxoSet := UTXOSet{blockchain}

	tx := NewUtxoTransaction(from, to, amount, &utxoSet)
	coinbaseTx := NewCoinbaseTx(from, "")
	block := blockchain.MineBlock([]*Transaction{coinbaseTx, tx})
	utxoSet.Update(block)
	fmt.Println("Success")
}

func (cli *CLI) CreateWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
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
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	createChainAddress := createChainCmd.String("address", "", "The address to send genesis block reward to")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to check balance for")
	sendFromAddress := sendCmd.String("from", "", "The address to send from")
	sendToAddress := sendCmd.String("to", "", "The address to send to")
	sendAmount := sendCmd.Int("amount", 0, "The amount to send")

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
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
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

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.GetBalance(*getBalanceAddress)
	}

	if sendCmd.Parsed() {
		if *sendFromAddress == "" || *sendToAddress == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.Send(*sendFromAddress, *sendToAddress, *sendAmount)
	}

	if createWalletCmd.Parsed() {
		cli.CreateWallet()
	}
}

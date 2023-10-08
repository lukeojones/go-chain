package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
)

type Version struct {
	Version    int
	BestHeight int
	AddrFrom   string
}
type GetBlocks struct {
	AddrFrom string
}
type Inventory struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}
type GetData struct {
	AddrFrom string
	Type     string
	ID       []byte
}
type BlockData struct {
	AddrFrom string
	Block    []byte
}
type TxData struct {
	AddrFrom    string
	Transaction []byte
}

const nodeVersion = 1
const protocol = "tcp"
const commandLength = 12

var nodeAddress string
var knownNodes = []string{"localhost:3000"}
var blocksInTransit = [][]byte{}
var miningAddress string // only set on mining nodes
var mempool = make(map[string]Transaction)

func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	listener, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer listener.Close()

	bc := NewBlockchain(nodeID)

	// All nodes (excluding the central one) send a version to the central
	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)
	}
}

func handleConnection(conn net.Conn, bc *Blockchain) {
	req, err := io.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(req[:commandLength])
	fmt.Printf("Received [%s] command\n", command)

	switch command {
	case "inventory":
		handleInventory(req, bc)
	case "version":
		handleVersion(req, bc)
	case "getblocks":
		handleGetBlocks(req, bc)
	case "blockdata":
		handleBlockData(req, bc)
	case "getdata":
		handleGetData(req, bc)
	case "txdata":
		handleTxData(req, bc)
	default:
		fmt.Println("Unknown Command!")
	}

	conn.Close()
}

func handleInventory(req []byte, bc *Blockchain) {
	var buffer bytes.Buffer
	var inv Inventory

	buffer.Write(req[commandLength:])
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&inv)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Received inventory with %d %s\n", len(inv.Items), inv.Type)
	if inv.Type == "block" {
		// Record the block hashes from the incoming message and mark them for later download
		blocksInTransit = inv.Items

		// Immediately download the first block (in reality, blocks would be downloaded from different nodes)
		blockHash := inv.Items[0]
		sendGetData(inv.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, block := range blocksInTransit {
			if bytes.Compare(block, blockHash) != 0 {
				newInTransit = append(newInTransit, block)
			}
		}
		blocksInTransit = newInTransit
	}

	if inv.Type == "tx" {
		//do tx inventory
	}
}

func handleVersion(req []byte, bc *Blockchain) {
	var buffer bytes.Buffer
	var version Version

	buffer.Write(req[commandLength:])
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&version)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := bc.GetBestHeight()
	otherBestHeight := version.BestHeight

	if myBestHeight < otherBestHeight {
		sendGetBlocks(version.AddrFrom)
	} else if myBestHeight > otherBestHeight {
		//send version back
		sendVersion(version.AddrFrom, bc)
	}

	if !nodeIsKnown(version.AddrFrom) {
		knownNodes = append(knownNodes, version.AddrFrom)
	}
}

func handleGetBlocks(req []byte, bc *Blockchain) {
	var buffer bytes.Buffer
	var getblocks GetBlocks

	buffer.Write(req[commandLength:])
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&getblocks)
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.GetBlockHashes()
	sendInventory(getblocks.AddrFrom, "block", blocks)
}

func handleGetData(req []byte, bc *Blockchain) {
	var buffer bytes.Buffer
	var getdata GetData

	buffer.Write(req[commandLength:])
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&getdata)
	if err != nil {
		log.Panic(err)
	}

	if getdata.Type == "block" {
		block, err := bc.GetBlock([]byte(getdata.ID))
		if err != nil {
			log.Panic(err)
		}
		sendBlock(getdata.AddrFrom, &block)
	}

	if getdata.Type == "tx" {
		txID := hex.EncodeToString(getdata.ID)
		tx := mempool[txID]
		sendTx(getdata.AddrFrom, &tx)
	}
}

func handleBlockData(request []byte, bc *Blockchain) {
	var buffer bytes.Buffer
	var blockdata BlockData

	buffer.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&blockdata)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("Received a new block!")
	block := DeserializeBlock(blockdata.Block)
	bc.AddBlock(block)

	fmt.Printf("Added block %x\n\n", block.Hash)
	// If there are more blocks to download, then request them now (from the node that just sent us this one)
	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(blockdata.AddrFrom, "block", blockHash)
		blocksInTransit = blocksInTransit[1:]
	} else {
		// If we have all the blocks, reindex the utxo set
		utxoSet := UTXOSet{bc}
		utxoSet.Reindex()
	}
}

func handleTxData(request []byte, bc *Blockchain) {
	var buffer bytes.Buffer
	var txdata TxData

	buffer.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buffer)
	decoder.Decode(&txdata)

	// First things first, add the tx to the mempool
	txbytes := txdata.Transaction
	tx := DeserializeTransaction(txbytes)
	mempool[hex.EncodeToString(tx.ID)] = tx

	// If this node is the central node, just propagate the transactions
	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress && node != txdata.AddrFrom {
				sendInventory(node, "tx", [][]byte{tx.ID})
			}
		}
	}
	// If this is the miner node, mine transactions
	if len(miningAddress) > 0 && len(mempool) > 2 {
	MineTransactions:
		var txs []*Transaction
		// verify all the transactions
		for _, tx := range mempool {
			if bc.VerifyTransaction(&tx) {
				txs = append(txs, &tx)
			}
		}

		if len(txs) == 0 {
			fmt.Println("All transactions are invalid! Waiting for new ones...")
			return
		}

		// Create coinbase txn and add to block
		coinbaseTx := NewCoinbaseTx(miningAddress, "")
		txs = append(txs, coinbaseTx)
		newBlock := bc.MineBlock(txs)

		// Re-index now that we've added the new block
		utxoSet := UTXOSet{bc}
		utxoSet.Reindex()

		fmt.Println("New block has been mined!")

		// Remove mined txs from mempool
		for _, tx := range txs {
			txID := hex.EncodeToString(tx.ID)
			delete(mempool, txID)
		}

		// Inform the other nodes that a new block exists
		for _, node := range knownNodes {
			sendInventory(node, "block", [][]byte{newBlock.Hash})
		}

		// Repeat until the mempool is clear
		if len(mempool) > 0 {
			goto MineTransactions
		}
	}

}
func sendTx(addr string, tx *Transaction) {
	data := TxData{nodeAddress, tx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("txdata"), payload...)
	sendData(addr, request)
}

func sendBlock(addr string, b *Block) {
	data := BlockData{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("blockdata"), payload...)
	sendData(addr, request)
}

func sendInventory(addr string, tipe string, items [][]byte) {
	payload := gobEncode(Inventory{nodeAddress, tipe, items})
	request := append(commandToBytes("inventory"), payload...)
	sendData(addr, request)
}

func sendGetBlocks(addr string) {
	payload := gobEncode(GetBlocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)
	sendData(addr, request)
}

func sendGetData(addr string, tipe string, id []byte) {
	payload := gobEncode(GetData{nodeAddress, tipe, id})
	request := append(commandToBytes("getdata"), payload...)
	sendData(addr, request)
}

func sendVersion(addr string, bc *Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(Version{nodeVersion, bestHeight, nodeAddress})
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}

func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string
		for _, node := range knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}
		knownNodes = updatedNodes
		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if addr == node {
			return true
		}
	}
	return false
}

func bytesToCommand(data []byte) string {
	var command []byte
	for _, b := range data {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte
	for i, c := range command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}

func gobEncode(data interface{}) []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buffer.Bytes()
}

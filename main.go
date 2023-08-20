package main

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxInput struct {
	TxID          []byte
	TxOutputIndex int
	ScriptSig     string
}

type TxOutput struct {
	Value        int
	ScriptPubKey string
}

const blockSubsidy = 50

func main() {
	blockchain := NewBlockchain()
	defer blockchain.db.Close()

	cli := CLI{blockchain}
	cli.Run()
}

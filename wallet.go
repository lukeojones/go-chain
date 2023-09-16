package main

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	_ "golang.org/x/crypto/ripemd160"
	"log"
	"math/big"
	"os"
)

const addressGenerationVersion = byte(0x00)
const addressChecksumLen = 4
const walletFile = "wallet.dat"

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

type WalletData struct {
	PublicKeyX, PublicKeyY *big.Int
	PrivateKeyD            *big.Int
}

type Wallets struct {
	WalletDatas map[string]*WalletData
}

func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	file, err := os.ReadFile(walletFile)
	if err != nil {
		return err
	}

	var wallets Wallets
	//gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(file))
	err = decoder.Decode(&wallets)
	if err != nil {
		return err
	}
	ws.WalletDatas = wallets.WalletDatas
	return nil
}

// SaveToFile saves wallets to a file
func (ws Wallets) SaveToFile() {
	var content bytes.Buffer
	//gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	if err := encoder.Encode(ws); err != nil {
		log.Panic(err)
	}

	if err := os.WriteFile(walletFile, content.Bytes(), 0644); err != nil {
		log.Panic(err)
	}
}

func (ws Wallets) GetWallet(address string) Wallet {
	walletData := *ws.WalletDatas[address]
	return *walletData.GetWallet()
}

func (ws Wallets) CreateWallet() string {
	walletData := NewWalletData()
	address := fmt.Sprintf("%s", walletData.GetWallet().GetAddress())
	ws.WalletDatas[address] = walletData
	return address
}

func NewWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.WalletDatas = make(map[string]*WalletData)

	err := wallets.LoadFromFile()
	return &wallets, err
}

func NewWalletData() *WalletData {
	private, _ := newKeyPair()
	return &WalletData{
		PublicKeyX:  private.PublicKey.X,
		PublicKeyY:  private.PublicKey.Y,
		PrivateKeyD: private.D,
	}
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...) // In ECDSA, public keys are X,Y co-ordinates on a curve
	return *private, pubKey
}

func (wallet *Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(wallet.PublicKey)

	versionedPayload := append([]byte{addressGenerationVersion}, pubKeyHash...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	return Base58Encode(fullPayload)
}

func (walletData *WalletData) GetWallet() *Wallet {
	curve := elliptic.P256()
	publicKey := ecdsa.PublicKey{
		Curve: curve,
		X:     walletData.PublicKeyX,
		Y:     walletData.PublicKeyY,
	}
	privateKey := ecdsa.PrivateKey{
		PublicKey: publicKey,
		D:         walletData.PrivateKeyD,
	}
	pubKeyBytes := append(publicKey.X.Bytes(), publicKey.Y.Bytes()...)
	return &Wallet{privateKey, pubKeyBytes}
}

func checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:addressChecksumLen]
}

func HashPubKey(pubKey []byte) []byte {
	pubKeySHA256 := sha256.Sum256(pubKey)
	hasher := crypto.RIPEMD160.New()
	if _, err := hasher.Write(pubKeySHA256[:]); err != nil {
		log.Panic(err)
	}
	pubKeyRIPEMD160 := hasher.Sum(nil)
	return pubKeyRIPEMD160

}

func ConvertBase58AddressToPubKeyHash(address string) []byte {
	return ConvertBase58BytesToPubKeyHash([]byte(address))
}
func ConvertBase58BytesToPubKeyHash(address []byte) []byte {
	pubKeyHashRecipient := Base58Decode(address)
	pubKeyHashRecipient = pubKeyHashRecipient[1 : len(pubKeyHashRecipient)-addressChecksumLen] // strip off version and checksum
	return pubKeyHashRecipient
}

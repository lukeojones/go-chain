package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"
)

const addressGenerationVersion = byte(0x00)
const addressChecksumLen = 4

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallet() *Wallet {
	private, public := newKeyPair()
	return &Wallet{
		PrivateKey: private,
		PublicKey:  public,
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

package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"

	"golang.org/x/crypto/ripemd160"
)

const (
	version            = byte(0x00)
	walletFile        = "wallet.dat"
	addressChecksumLen = 4
)

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// SerializedWallet is used for storage
type SerializedWallet struct {
	PrivateKeyD    *big.Int
	PrivateKeyX    *big.Int
	PrivateKeyY    *big.Int
	PrivateKeyCurve string
	PublicKey      []byte
}

// Wallets stores a collection of wallets
type Wallets struct {
	Wallets map[string]*Wallet
}

// NewWallet creates and returns a new Wallet
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

// NewWallets creates Wallets and fills it from a file if it exists
func NewWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile()

	return &wallets, err
}

// CreateWallet adds a Wallet to Wallets
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())

	ws.Wallets[address] = wallet

	return address
}

// GetAddresses returns an array of addresses stored in the wallet file
func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet returns a Wallet by its address
func (ws *Wallets) GetWallet(address string) *Wallet {
	return ws.Wallets[address]
}

// LoadFromFile loads wallets from the file
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return nil
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return err
	}

	var serializedWallets map[string]*SerializedWallet
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&serializedWallets)
	if err != nil {
		return err
	}

	ws.Wallets = make(map[string]*Wallet)
	for addr, sWallet := range serializedWallets {
		privateKey := ecdsa.PrivateKey{
			D: sWallet.PrivateKeyD,
			PublicKey: ecdsa.PublicKey{
				Curve: elliptic.P256(),
				X:     sWallet.PrivateKeyX,
				Y:     sWallet.PrivateKeyY,
			},
		}
		ws.Wallets[addr] = &Wallet{
			PrivateKey: privateKey,
			PublicKey:  sWallet.PublicKey,
		}
	}

	return nil
}

// SaveToFile saves wallets to a file
func (ws *Wallets) SaveToFile() error {
	var content bytes.Buffer

	// Convert Wallets to SerializedWallets for storage
	serializedWallets := make(map[string]*SerializedWallet)
	for addr, wallet := range ws.Wallets {
		serializedWallets[addr] = &SerializedWallet{
			PrivateKeyD:    wallet.PrivateKey.D,
			PrivateKeyX:    wallet.PrivateKey.PublicKey.X,
			PrivateKeyY:    wallet.PrivateKey.PublicKey.Y,
			PrivateKeyCurve: "P-256",
			PublicKey:      wallet.PublicKey,
		}
	}

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(serializedWallets)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

// GetAddress returns wallet address
func (w *Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := Base58Encode(fullPayload)

	return address
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// ValidateAddress check if address if valid
func ValidateAddress(address string) bool {
	if len(address) == 0 {
		return false
	}

	pubKeyHash := Base58Decode([]byte(address))
	if len(pubKeyHash) < addressChecksumLen {
		return false
	}

	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Equal(actualChecksum, targetChecksum)
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

// newKeyPair creates a new key pair
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}

// Base58Encode encodes a byte array to Base58
func Base58Encode(input []byte) []byte {
	var result []byte

	x := new(big.Int).SetBytes(input)

	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}

	// https://en.bitcoin.it/wiki/Base58Check_encoding\#Version_bytes
	if input[0] == 0x00 {
		result = append(result, b58Alphabet[0])
	}

	ReverseBytes(result)

	return result
}

// Base58Decode decodes Base58-encoded data
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)

	for _, b := range input {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()

	if input[0] == b58Alphabet[0] {
		decoded = append([]byte{0x00}, decoded...)
	}

	return decoded
}

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

// GetStringAddress returns the string representation of the address
func (w *Wallet) GetStringAddress() string {
	return hex.EncodeToString(w.GetAddress())
}

// ExportPrivateKey exports the private key in a format suitable for signing
func (w *Wallet) ExportPrivateKey() *ecdsa.PrivateKey {
	return &w.PrivateKey
}

// SignData signs the provided data with the wallet's private key
func (w *Wallet) SignData(data []byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, &w.PrivateKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %v", err)
	}
	
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

// VerifySignature verifies that a signature is valid for the given data and public key
func VerifySignature(pubKey []byte, data []byte, signature []byte) bool {
	if len(signature) == 0 || len(pubKey) == 0 {
		return false
	}

	curve := elliptic.P256()
	
	r := big.Int{}
	s := big.Int{}
	sigLen := len(signature)
	r.SetBytes(signature[:(sigLen / 2)])
	s.SetBytes(signature[(sigLen / 2):])

	x := big.Int{}
	y := big.Int{}
	keyLen := len(pubKey)
	x.SetBytes(pubKey[:(keyLen / 2)])
	y.SetBytes(pubKey[(keyLen / 2):])

	rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
	return ecdsa.Verify(&rawPubKey, data, &r, &s)
}

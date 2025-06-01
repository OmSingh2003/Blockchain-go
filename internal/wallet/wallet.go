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
    "path/filepath"

    "golang.org/x/crypto/ripemd160"
)

const (
    version            = byte(0x00)
    walletFile        = "wallet.dat"
    addressChecksumLen = 4
)

// Wallet stores private and public keys
type Wallet struct {
    PrivateKey ecdsa.PrivateKey
    PublicKey  []byte
}

// walletSerializable is used for wallet serialization
type walletSerializable struct {
    PrivateKeyD    []byte
    PrivateKeyX    []byte
    PrivateKeyY    []byte
    PublicKey      []byte
}

func init() {
    gob.Register(elliptic.P256())
}

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
    private, public := newKeyPair()
    wallet := Wallet{private, public}
    
    // Save the wallet immediately after creation
    SaveWallet(wallet.GetAddress(), &wallet)
    
    return &wallet
}

// LoadWallet loads a wallet from a file
func LoadWallet(address string) *Wallet {
    if !ValidateAddress(address) {
        return nil
    }

    walletDir := getWalletDir()
    walletPath := filepath.Join(walletDir, fmt.Sprintf("%s.wallet", address))

    if _, err := os.Stat(walletPath); os.IsNotExist(err) {
        return nil
    }

    fileContent, err := ioutil.ReadFile(walletPath)
    if err != nil {
        log.Panic(err)
    }

    var ws walletSerializable
    decoder := gob.NewDecoder(bytes.NewReader(fileContent))
    err = decoder.Decode(&ws)
    if err != nil {
        log.Panic(err)
    }

    curve := elliptic.P256()
    x := new(big.Int).SetBytes(ws.PrivateKeyX)
    y := new(big.Int).SetBytes(ws.PrivateKeyY)
    d := new(big.Int).SetBytes(ws.PrivateKeyD)

    privateKey := ecdsa.PrivateKey{
        PublicKey: ecdsa.PublicKey{
            Curve: curve,
            X:     x,
            Y:     y,
        },
        D: d,
    }

    return &Wallet{privateKey, ws.PublicKey}
}

// SaveWallet saves the wallet to a file
func SaveWallet(address string, wallet *Wallet) {
    walletDir := getWalletDir()
    if err := os.MkdirAll(walletDir, 0700); err != nil {
        log.Panic(err)
    }

    walletPath := filepath.Join(walletDir, fmt.Sprintf("%s.wallet", address))

    ws := walletSerializable{
        PrivateKeyD: wallet.PrivateKey.D.Bytes(),
        PrivateKeyX: wallet.PrivateKey.X.Bytes(),
        PrivateKeyY: wallet.PrivateKey.Y.Bytes(),
        PublicKey:   wallet.PublicKey,
    }

    var content bytes.Buffer
    encoder := gob.NewEncoder(&content)
    err := encoder.Encode(ws)
    if err != nil {
        log.Panic(err)
    }

    err = ioutil.WriteFile(walletPath, content.Bytes(), 0600)
    if err != nil {
        log.Panic(err)
    }
}

// ListAddresses returns a list of addresses of all wallets
func ListAddresses() []string {
    var addresses []string
    walletDir := getWalletDir()
    
    files, err := ioutil.ReadDir(walletDir)
    if err != nil && !os.IsNotExist(err) {
        log.Panic(err)
    }

    for _, f := range files {
        if filepath.Ext(f.Name()) == ".wallet" {
            address := f.Name()[:len(f.Name())-7] // Remove .wallet extension
            if ValidateAddress(address) {
                addresses = append(addresses, address)
            }
        }
    }

    return addresses
}

// GetAddress returns wallet address
func (w *Wallet) GetAddress() string {
    pubKeyHash := HashPubKey(w.PublicKey)

    versionedPayload := append([]byte{version}, pubKeyHash...)
    checksum := checksum(versionedPayload)

    fullPayload := append(versionedPayload, checksum...)
    address := hex.EncodeToString(fullPayload)

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
    pubKeyHash, err := hex.DecodeString(address)
    if err != nil {
        return false
    }

    if len(pubKeyHash) < addressChecksumLen+1 {
        return false
    }

    actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
    version := pubKeyHash[0]
    pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
    targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

    return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// SignData signs data using the wallet's private key
func (w *Wallet) SignData(data []byte) ([]byte, error) {
    r, s, err := ecdsa.Sign(rand.Reader, &w.PrivateKey, data)
    if err != nil {
        return nil, err
    }

    signature := append(r.Bytes(), s.Bytes()...)
    return signature, nil
}

// VerifySignature verifies a signature against public key and data
func VerifySignature(pubKey []byte, data []byte, signature []byte) bool {
    curve := elliptic.P256()
    r := new(ecdsa.PublicKey)
    r.Curve = curve
    r.X, r.Y = curve.ScalarBaseMult(pubKey)

    if len(signature) != 64 {
        return false
    }

    rSign := new(big.Int).SetBytes(signature[:32])
    sSign := new(big.Int).SetBytes(signature[32:])

    return ecdsa.Verify(r, data, rSign, sSign)
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
    firstSHA := sha256.Sum256(payload)
    secondSHA := sha256.Sum256(firstSHA[:])

    return secondSHA[:addressChecksumLen]
}

// newKeyPair creates a new cryptographic key pair
func newKeyPair() (ecdsa.PrivateKey, []byte) {
    curve := elliptic.P256()
    private, err := ecdsa.GenerateKey(curve, rand.Reader)
    if err != nil {
        log.Panic(err)
    }
    
    pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
    return *private, pubKey
}

// getWalletDir returns the directory where wallet files are stored
func getWalletDir() string {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        log.Panic(err)
    }
    return filepath.Join(homeDir, ".blockchain-wallets")
}

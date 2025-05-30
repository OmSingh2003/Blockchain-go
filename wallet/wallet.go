package wallet
import (

)
type Wallet struct {
	PrivateKey   ecdsa.PrivateKey
	PublicKey    []byte
}
type Wallets struct  {
	Wallets map[string] *Wallet
}
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private,public}

	return &wallet
}

func newKeyPair() (ecdsa.PrivateKey,[]byte) {
	curve := elliptic.P256()
	private , err := ecdsa.GenerateKey(curve, Rand.Reader)
	pubKey := append(PrivateKey.PublicKey.X.Bytes(),private.PublicKey.Y.Bytes()...)

	return *private,pubKey
}

func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)
	versionPayLoad := append([]byte{version},pubKeyHash...)
	checksum := checksum(versionPayload)

	fullPayload := append(versionPayload, checksum...)
	address := Base58Encode(fullPayload)

	return address
}

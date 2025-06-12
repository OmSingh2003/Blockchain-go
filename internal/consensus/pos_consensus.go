package consensus

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math/big"

	"github.com/OmSingh2003/decentralized-ledger/internal/block"
	"github.com/OmSingh2003/decentralized-ledger/internal/transaction"
	"github.com/OmSingh2003/decentralized-ledger/internal/wallet"
	"go.etcd.io/bbolt"
)

// Constants for PoS specific buckets
const (
	validatorsBucket = "validators"
	stakesBucket     = "stakes"
)

// Validator struct representing a staking entity
type Validator struct {
	Address   string
	PublicKey []byte // Stored in raw form
	Stake     int64
	// Additional fields like LastProposedBlock, JailedStatus etc. can be added
}

// PoSConsensus implements the Consensus interface for Proof-of-Stake.
type PoSConsensus struct {
	db           *bbolt.DB
	validatorSet []Validator // In-memory cache of current validators
}

// NewPoSConsensus creates a new PoSConsensus instance and loads validators.
func NewPoSConsensus(db *bbolt.DB) *PoSConsensus {
	pos := &PoSConsensus{db: db}
	if err := pos.loadValidators(); err != nil {
		fmt.Printf("Warning: Failed to load validators for PoS: %v. Starting with empty set.\n", err)
		// Optionally, log.Panic(err) if validators are critical for startup
	}
	return pos
}

// ProposeBlock for PoS consensus involves selecting a validator and signing the block.
// The `proposerWallet` is the wallet of the node attempting to propose.
func (p *PoSConsensus) ProposeBlock(proposerWallet *wallet.Wallet, transactions []*transaction.Transaction, prevBlockHash []byte, currentTipHash []byte) (*block.Block, error) {
	// 1. Select a validator who is allowed to propose the next block.
	// In a real PoS, this would involve a more sophisticated mechanism (e.g., VRF, turn-based).
	// For now, we use a weighted random selection and assume the `proposerWallet` matches the selected validator.
	selectedValidator, err := p.selectValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to select a validator: %v", err)
	}

	// In a real network, you'd only propose if your local node is the selectedValidator.
	// For this simulation, we'll assume the `proposerWallet` is the selected one if provided.
	// You need to ensure the proposer's public key matches the selected validator.
	if !bytes.Equal(proposerWallet.PublicKey, selectedValidator.PublicKey) {
		return nil, fmt.Errorf("current wallet is not the selected validator (%s vs %s)",
			wallet.HashPubKey(proposerWallet.PublicKey), wallet.HashPubKey(selectedValidator.PublicKey))
	}

	// Create new block
	newBlock := block.NewBlock(transactions, prevBlockHash)
	// Note: Timestamp is already set in NewBlock constructor

	// Set validator's public key in the block header
	newBlock.SetValidatorPubKey(proposerWallet.PublicKey)

	// Calculate the hash of the block's contents (excluding the signature itself)
	// This is the data that the validator will sign.
	hashableData := newBlock.GetHashableDataPoS()
	
	// Hash the data before signing for security
	dataHash := sha256.Sum256(hashableData)
	
	// Sign the hashed data using the proposer's private key
	signature, err := proposerWallet.SignData(dataHash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign block: %v", err)
	}

	// Set the signature in the block
	newBlock.SetSignature(signature)

	// The block's actual hash (ID) is derived from its full content (including signature)
	newBlock.Hash = newBlock.GetPoSHash()

	return newBlock, nil
}

// ValidateBlock for PoS consensus involves verifying validator signatures and stake.
func (p *PoSConsensus) ValidateBlock(b *block.Block, prevTXs map[string]transaction.Transaction) (bool, error) {
	// 1. Basic block structure and transaction validation
	if err := b.ValidateBlock(prevTXs); err != nil {
		return false, fmt.Errorf("block structure/transaction validation failed: %v", err)
	}

	// 2. Verify validator's public key and signature
	if len(b.GetValidatorPubKey()) == 0 || len(b.GetSignature()) == 0 {
		return false, fmt.Errorf("PoS block missing validator public key or signature")
	}

	// Reconstruct the data that was signed
	hashableData := b.GetHashableDataPoS()
	
	// Hash the data (same as in signing)
	dataHash := sha256.Sum256(hashableData)
	
	// Verify the signature using the validator's public key
	isValidSignature := wallet.VerifySignature(b.GetValidatorPubKey(), dataHash[:], b.GetSignature())
	if !isValidSignature {
		return false, fmt.Errorf("invalid validator signature for block %x", b.GetHash())
	}

	// 3. Check if the validator is part of the current active validator set and has enough stake.
	foundValidator := false
	var actualStake int64 = 0
	for _, v := range p.validatorSet {
		if bytes.Equal(v.PublicKey, b.GetValidatorPubKey()) {
			foundValidator = true
			actualStake = v.Stake
			break
		}
	}

	if !foundValidator {
		return false, fmt.Errorf("validator %x not found in active set", b.GetValidatorPubKey())
	}

	// You would define a minimum stake requirement here
	// For example:
	minStake := int64(100) // Example minimum stake
	if actualStake < minStake {
		return false, fmt.Errorf("validator %x has insufficient stake (%d, required %d)", b.GetValidatorPubKey(), actualStake, minStake)
	}

	// 4. Optionally, add more advanced PoS validation (e.g., checking for double-signing, proposer fairness)
	// This would require a network layer and state tracking beyond just the database.

	return true, nil
}

// GetCurrentDifficulty for PoS might return information about the current validator set or next proposer.
func (p *PoSConsensus) GetCurrentDifficulty(blockchainTipHash []byte) (interface{}, error) {
	// For PoS, "difficulty" might be represented by the active validator set.
	// In a more complex system, it could include expected proposer, slot time, etc.
	return p.validatorSet, nil // Returning the in-memory validator set
}

// loadValidators initializes validators from the database
func (p *PoSConsensus) loadValidators() error {
	p.validatorSet = []Validator{} // Clear existing set
	err := p.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(validatorsBucket))
		if b == nil {
			// If bucket doesn't exist, it's fine for initial run, but log it.
			return fmt.Errorf("'%s' bucket not found, no validators loaded", validatorsBucket)
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var validator Validator
			decoder := gob.NewDecoder(bytes.NewReader(v))
			if err := decoder.Decode(&validator); err != nil {
				return fmt.Errorf("failed to decode validator %x: %v", k, err)
			}
			p.validatorSet = append(p.validatorSet, validator)
		}
		return nil
	})
	if err != nil {
		// If the bucket wasn't found, it's not necessarily an error, just means no validators yet.
		if err.Error() == fmt.Sprintf("'%s' bucket not found, no validators loaded", validatorsBucket) {
			return nil // Return nil if just no validators
		}
		return fmt.Errorf("error loading validators: %v", err)
	}
	return nil
}

// SaveValidator saves a validator to the database (NEW helper function)
func (p *PoSConsensus) SaveValidator(validator Validator) error {
	return p.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(validatorsBucket))
		if err != nil {
			return err
		}

		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		if err := encoder.Encode(validator); err != nil {
			return err
		}

		// Use validator's public key hash as the key
		key := wallet.HashPubKey(validator.PublicKey)
		return b.Put(key, buffer.Bytes())
	})
}

// selectValidator selects a validator based on their stake using weighted random choice.
func (p *PoSConsensus) selectValidator() (Validator, error) {
	if len(p.validatorSet) == 0 {
		return Validator{}, fmt.Errorf("no validators available in the set")
	}

	totalStake := big.NewInt(0)
	for _, v := range p.validatorSet {
		if v.Stake > 0 { // Only consider validators with positive stake
			totalStake.Add(totalStake, big.NewInt(v.Stake))
		}
	}

	if totalStake.Cmp(big.NewInt(0)) == 0 {
		return Validator{}, fmt.Errorf("total stake is zero, no validators to select from")
	}

	randNum, err := rand.Int(rand.Reader, totalStake)
	if err != nil {
		return Validator{}, fmt.Errorf("failed to generate random number for validator selection: %v", err)
	}

	var cumulativeStake int64 = 0

	for _, v := range p.validatorSet {
		if v.Stake > 0 {
			cumulativeStake += v.Stake
			if randNum.Cmp(big.NewInt(cumulativeStake)) < 0 {
				return v, nil
			}
		}
	}

	return Validator{}, fmt.Errorf("failed to select a validator (should not happen if totalStake > 0)")
}

// AddStake allows a wallet to add stake to become a validator or increase existing stake (NEW helper)
// This would typically be a transaction that updates the validator's stake.
// For now, it's a direct function for testing.
func (p *PoSConsensus) AddStake(stakeAmount int64, w *wallet.Wallet) error {
	if stakeAmount <= 0 {
		return fmt.Errorf("stake amount must be positive")
	}

	validatorAddress := w.GetAddress()

	var existingValidator *Validator
	for i := range p.validatorSet {
		if bytes.Equal(p.validatorSet[i].PublicKey, w.PublicKey) {
			existingValidator = &p.validatorSet[i]
			break
		}
	}

	if existingValidator != nil {
		existingValidator.Stake += stakeAmount
		fmt.Printf("Updated stake for validator %s to %d\n", validatorAddress, existingValidator.Stake)
		return p.SaveValidator(*existingValidator)
	} else {
		newValidator := Validator{
			Address:   validatorAddress,
			PublicKey: w.PublicKey,
			Stake:     stakeAmount,
		}
		p.validatorSet = append(p.validatorSet, newValidator)
		fmt.Printf("Registered new validator %s with stake %d\n", validatorAddress, stakeAmount)
		return p.SaveValidator(newValidator)
	}
}

func init() {
	// Register Validator struct for gob serialization
	gob.Register(Validator{})
}

package consensus

import (
	"testing"

	"github.com/OmSingh2003/decentralized-ledger/internal/transaction"
	"github.com/OmSingh2003/decentralized-ledger/internal/wallet"
)

// Test PoS Consensus implements Consensus interface
func TestPoSConsensusInterface(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	var consensus Consensus = NewPoSConsensus(db)
	if consensus == nil {
		t.Error("PoSConsensus should implement Consensus interface")
	}
}

// Test validator registration and staking
func TestValidatorStaking(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	pos := NewPoSConsensus(db)
	validatorWallet := wallet.NewWallet()

	// Add stake
	err := pos.AddStake(1000, validatorWallet)
	if err != nil {
		t.Fatalf("Failed to add stake: %v", err)
	}

	// Check validator was added
	if len(pos.validatorSet) != 1 {
		t.Errorf("Expected 1 validator, got %d", len(pos.validatorSet))
	}

	if pos.validatorSet[0].Stake != 1000 {
		t.Errorf("Expected stake 1000, got %d", pos.validatorSet[0].Stake)
	}
}

// Test validator selection
func TestValidatorSelection(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	pos := NewPoSConsensus(db)

	// Add multiple validators with different stakes
	validator1 := wallet.NewWallet()
	validator2 := wallet.NewWallet()

	err := pos.AddStake(500, validator1)
	if err != nil {
		t.Fatalf("Failed to add stake for validator1: %v", err)
	}

	err = pos.AddStake(1500, validator2)
	if err != nil {
		t.Fatalf("Failed to add stake for validator2: %v", err)
	}

	// Test selection multiple times to check randomness
	selections := make(map[string]int)
	for i := 0; i < 100; i++ {
		selected, err := pos.selectValidator()
		if err != nil {
			t.Fatalf("Failed to select validator: %v", err)
		}
		selections[selected.Address]++
	}

	// Validator2 should be selected more often due to higher stake
	if len(selections) != 2 {
		t.Errorf("Expected selections from 2 validators, got %d", len(selections))
	}
}

// Test PoS block proposal
func TestPoSProposeBlock(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	pos := NewPoSConsensus(db)
	validatorWallet := wallet.NewWallet()

	// Add validator with stake
	err := pos.AddStake(1000, validatorWallet)
	if err != nil {
		t.Fatalf("Failed to add stake: %v", err)
	}

	// Create test transaction
	coinbaseTx := createCoinbaseTransaction()
	transactions := []*transaction.Transaction{coinbaseTx}

	// Propose block
	block, err := pos.ProposeBlock(validatorWallet, transactions, []byte{}, []byte{})
	if err != nil {
		t.Fatalf("Failed to propose block: %v", err)
	}

	if block == nil {
		t.Error("Proposed block should not be nil")
	}

	// Check block has validator signature
	if len(block.GetSignature()) == 0 {
		t.Error("Block should have validator signature")
	}

	if len(block.GetValidatorPubKey()) == 0 {
		t.Error("Block should have validator public key")
	}
}

// Test PoS block validation
func TestPoSValidateBlock(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	pos := NewPoSConsensus(db)
	validatorWallet := wallet.NewWallet()

	// Add validator with stake
	err := pos.AddStake(1000, validatorWallet)
	if err != nil {
		t.Fatalf("Failed to add stake: %v", err)
	}

	// Create and propose a valid block
	coinbaseTx := createCoinbaseTransaction()
	transactions := []*transaction.Transaction{coinbaseTx}

	validBlock, err := pos.ProposeBlock(validatorWallet, transactions, []byte{}, []byte{})
	if err != nil {
		t.Fatalf("Failed to propose block: %v", err)
	}

	// Validate the block
	valid, err := pos.ValidateBlock(validBlock, make(map[string]transaction.Transaction))
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if !valid {
		t.Error("Valid PoS block should pass validation")
	}
}

// Test PoS validation fails for invalid signature
func TestPoSValidationFailsInvalidSignature(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	pos := NewPoSConsensus(db)
	validatorWallet := wallet.NewWallet()

	// Add validator with stake
	err := pos.AddStake(1000, validatorWallet)
	if err != nil {
		t.Fatalf("Failed to add stake: %v", err)
	}

	// Create block with valid validator
	coinbaseTx := createCoinbaseTransaction()
	transactions := []*transaction.Transaction{coinbaseTx}

	validBlock, err := pos.ProposeBlock(validatorWallet, transactions, []byte{}, []byte{})
	if err != nil {
		t.Fatalf("Failed to propose block: %v", err)
	}

	// Tamper with signature (simulate attack)
	fakeSignature := make([]byte, 64)
	for i := range fakeSignature {
		fakeSignature[i] = byte(i % 256)
	}
	validBlock.SetSignature(fakeSignature)

	// Validation should fail
	valid, err := pos.ValidateBlock(validBlock, make(map[string]transaction.Transaction))
	if err == nil {
		t.Error("Validation should fail for tampered signature")
	}

	if valid {
		t.Error("Block with invalid signature should not be valid")
	}
}

// Test insufficient stake validation
func TestPoSInsufficientStake(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	pos := NewPoSConsensus(db)
	validatorWallet := wallet.NewWallet()

	// Add validator with insufficient stake (less than minimum)
	err := pos.AddStake(50, validatorWallet) // Minimum is 100 in the code
	if err != nil {
		t.Fatalf("Failed to add stake: %v", err)
	}

	// Create block
	coinbaseTx := createCoinbaseTransaction()
	transactions := []*transaction.Transaction{coinbaseTx}

	validBlock, err := pos.ProposeBlock(validatorWallet, transactions, []byte{}, []byte{})
	if err != nil {
		t.Fatalf("Failed to propose block: %v", err)
	}

	// Validation should fail due to insufficient stake
	valid, err := pos.ValidateBlock(validBlock, make(map[string]transaction.Transaction))
	if err == nil {
		t.Error("Validation should fail for insufficient stake")
	}

	if valid {
		t.Error("Block from validator with insufficient stake should not be valid")
	}
}


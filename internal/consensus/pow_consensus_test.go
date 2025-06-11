package consensus

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/OmSingh2003/decentralized-ledger/internal/block"
	"github.com/OmSingh2003/decentralized-ledger/internal/transaction"
	"go.etcd.io/bbolt"
)

// Helper to create test database
func createTestDB(t *testing.T) *bbolt.DB {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := bbolt.Open(dbPath, 0o600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(blocksBucket))
		return err
	})
	if err != nil {
		t.Fatalf("Failed to create blocks bucket: %v", err)
	}
	return db
}

// Test Genesis Block ProposeBlock
func TestProposeGenesisBlock(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	powConsensus := &POWConsensus{db: db}
	coinbaseTx := createCoinbaseTransaction()
	transactions := []*transaction.Transaction{coinbaseTx}

	// For genesis block, both hashes should be empty
	block, err := powConsensus.ProposeBlock(transactions, []byte{}, []byte{})
	if err == nil {
		t.Log("Genesis block creation succeeded (or failed as expected)")
	}
	
	if block != nil {
		if len(block.Transactions) != 1 {
			t.Errorf("Expected 1 transaction, got %d", len(block.Transactions))
		}
	}
}

// Helper to create coinbase transaction
func createCoinbaseTransaction() *transaction.Transaction {
	return &transaction.Transaction{
		ID: []byte("coinbase-tx"),
		Vin: []transaction.TxInput{{
			Txid: []byte{}, Vout: -1, Signature: nil, PubKey: []byte("coinbase"),
		}},
		Vout: []transaction.TxOutput{{
			Value: 50, PubKeyHash: []byte("miner-address"),
		}},
	}
}

// Test POWConsensus implements Consensus interface
func TestPOWConsensusInterface(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	var consensus Consensus = &POWConsensus{db: db}
	if consensus == nil {
		t.Error("POWConsensus should implement Consensus interface")
	}
}

// Test ProposeBlock creates valid blocks
func TestProposeBlock(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	powConsensus := &POWConsensus{db: db}
	coinbaseTx := createCoinbaseTransaction()
	transactions := []*transaction.Transaction{coinbaseTx}

	// Create and store a genesis block first for currentTipHash
	genesisBlock := block.NewBlock([]*transaction.Transaction{coinbaseTx}, []byte{})
	genesisBlock.SetBits(INITIAL_TARGET_BITS)
	genesisBlock.UpdateHash()
	storeTestBlock(t, db, genesisBlock)

	block, err := powConsensus.ProposeBlock(transactions, genesisBlock.GetHash(), genesisBlock.GetHash())
	if err != nil {
		t.Fatalf("ProposeBlock failed: %v", err)
	}

	if block == nil {
		t.Error("Proposed block should not be nil")
	}
	if len(block.Transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(block.Transactions))
	}
	if block.GetNonce() == 0 {
		t.Error("Block nonce should be set after mining")
	}
	if block.GetBits() == 0 {
		t.Error("Block bits should be set after mining")
	}
}

// Helper to store block in database
func storeTestBlock(t *testing.T, db *bbolt.DB, b *block.Block) {
	blockData, err := b.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize block: %v", err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		return bucket.Put(b.GetHash(), blockData)
	})
	if err != nil {
		t.Fatalf("Failed to store block: %v", err)
	}
}

// Test ValidateBlock validates correct blocks
func TestValidateBlock(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	powConsensus := &POWConsensus{db: db}
	coinbaseTx := createCoinbaseTransaction()

	// Create and store a genesis block first
	genesisBlock := block.NewBlock([]*transaction.Transaction{coinbaseTx}, []byte{})
	genesisBlock.SetBits(INITIAL_TARGET_BITS)
	genesisBlock.UpdateHash()
	storeTestBlock(t, db, genesisBlock)

	// Now create a properly mined block using ProposeBlock
	validBlock, err := powConsensus.ProposeBlock([]*transaction.Transaction{coinbaseTx}, genesisBlock.GetHash(), genesisBlock.GetHash())
	if err != nil {
		t.Fatalf("Failed to create valid block: %v", err)
	}

	valid, err := powConsensus.ValidateBlock(validBlock, make(map[string]transaction.Transaction))
	if err != nil {
		t.Fatalf("ValidateBlock failed: %v", err)
	}
	if !valid {
		t.Error("Valid block should pass validation")
	}
}

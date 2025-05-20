package ProofOfWork
import (
	"github.com/OmSingh2003/blockchain-go/blocks"
)
const targetBits = 24 
type ProofOFWork struct {
	block      *blocks.block
	tartget    *big.int
}
func NewProofOfWork(b *blocks.Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	
	pow :=&ProofOfWork{b,target}
}


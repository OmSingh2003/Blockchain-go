package transcations
import (
	"fmt"
)

func NewCoinbaseTX(to , data string) *Transcations {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	txin := TXInput{[]Byte{}, -1, data}
	txout := TXOutput{subsidy, to }
	tx := Transactions{nil, []TXInput{txin},[]TXOutput{txout}}
	tx.SetID()

	return &tx 
}

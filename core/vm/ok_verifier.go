package vm

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// TODO: move to core/okex
type ContractVerifier interface {
	Verify(stateDB StateDB, op OpCode, from, to common.Address, input []byte, value *big.Int) error
}

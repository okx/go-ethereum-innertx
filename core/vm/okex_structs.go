package vm

import "math/big"

const (
	CREATE_INNER_TX_CALL_TYPE      = "evm-create"
	CREATE2_INNER_TX_CALL_TYPE     = "evm-create2"
	CALL_INNER_TX_CALL_TYPE        = "evm-call"
	STATIC_CALL_INNER_TX_CALL_TYPE = "evm-staticcall"
	CALL_CODE_INNER_TX_CALL_TYPE   = "evm-callcode"
	DELEGATE_INNER_TX_CALL_TYPE    = "evm-delegatecall"
	SUICIDE_INNER_TX_CALL_TYPE     = "evm-suicide"
)

type InnerTx struct {
	Dept          big.Int `json:"dept"`
	InternalIndex big.Int `json:"internal_index"`
	CallType      string  `json:"call_type"`
	Name          string  `json:"name"`
	TraceAddress  string  `json:"trace_address"`
	CodeAddress   string  `json:"code_address"`
	From          string  `json:"from"`
	To            string  `json:"to"`
	Input         string  `json:"input"`
	Output        string  `json:"output"`
	IsError       bool    `json:"is_error"`
	GasUsed       uint64  `json:"gas_used"`
	Value         string  `json:"value"`
	ValueWei      string  `json:"value_wei"`
	Error         string  `json:"error"`
}

func (tx *InnerTx) SetError(err string) {
	tx.Error = err
	tx.IsError = true
}

type TokenInitInfo struct {
	ContractAddr string `json:"contract_addr"`
	Name         string `json:"name"`
	Symble       string `json:"symble"`
	Decimals     string `json:"decimals"`
	TotalSupply  string `json:"total_supply"`
	OwnerBalance string `json:"owner_init_balance"`
	OwnerAddr    string `json:"owner_addr"`
	Type         string `json:"type"`
}

type ERC20Contract struct {
	ContractAddr []byte `json:"contract_addr"`
	ContractCode []byte `json:"contract_code"`
}

type BlockInnerData struct {
	BlockHash    string
	TxHashes     []string
	TxMap        map[string][]*InnerTx
	ContractList []*ERC20Contract
}

func NewInnerTx(depth int64, from, to, valueWei, callType, name string) *InnerTx {
	return &InnerTx{
		Dept:     *big.NewInt(depth),
		From:     from,
		IsError:  false,
		To:       to,
		CallType: callType,
		Name:     name,
		ValueWei: valueWei,
	}
}

func NewInnerTxs() []*InnerTx {
	return make([]*InnerTx, 0)
}

package vm

import "math/big"

const (
	// TODO: unified constant in exchain/libs/cosmos-sdk/types/innertx/innerTx.go
	CosmosCallType = "cosmos"
	EvmCreateName  = "cosmos-create"
)

type InnerTxExport struct {
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

type ContractCreationInfoExport struct {
	Creator         string `json:"creator"`
	CreateType      string `json:"create_type"`
	NonceOnCreation string `json:"nonce_on_creation"`
	Salt            string `json:"salt"`
	CodeHash        string `json:"code_hash"`
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
	BlockHash           string
	TxHashes            []string
	TxMap               map[string][]*InnerTxExport
	ContractCreationMap map[string]*ContractCreationInfoExport
	ContractList        []*ERC20Contract
}

type InnerTxInternal struct {
	InnerTxExport

	FromNonce       string
	Create2Salt     string
	Create2CodeHash string
}

func BuildInnerTxExports(innerTxs []*InnerTxInternal) []*InnerTxExport {
	results := make([]*InnerTxExport, 0, len(innerTxs))

	for _, innertx := range innerTxs {
		results = append(results, &innertx.InnerTxExport)
	}

	return results
}

func BuildContractCreationInfos(innerTxs []*InnerTxInternal) map[string]*ContractCreationInfoExport {
	results := make(map[string]*ContractCreationInfoExport)

	for _, tx := range innerTxs {
		if !tx.isCreateContract() {
			continue
		}

		createType := tx.CallType
		if tx.CallType == CosmosCallType && tx.Name == EvmCreateName {
			// TODO: define calltype constant
			createType = "create"
		}
		cci := &ContractCreationInfoExport{
			Creator:         tx.From,
			CreateType:      createType,
			NonceOnCreation: tx.FromNonce,
			Salt:            tx.Create2Salt,
			CodeHash:        tx.Create2CodeHash,
		}
		results[tx.To] = cci
	}

	return results
}

func (intx *InnerTxInternal) isCreateContract() bool {
	// TODO: define calltype constant
	return intx.CallType == "create" ||
		intx.CallType == "create2" ||
		intx.Name == EvmCreateName
}

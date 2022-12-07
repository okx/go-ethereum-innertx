package okex

import (
	"math/big"
)

type InnerTxCallType string
type InnerTxNameType string
type ContractCreateType string

const (
	CosmosDepth = 0

	InnerTxCall_Cosmos       InnerTxCallType = "cosmos"
	InnerTxCall_Create       InnerTxCallType = "create"
	InnerTxCall_Create2      InnerTxCallType = "create2"
	InnerTxCall_Call         InnerTxCallType = "call"
	InnerTxCall_Callcode     InnerTxCallType = "callcode"
	InnerTxCall_Delegatecall InnerTxCallType = "delegatecall"
	InnerTxCall_Staticcall   InnerTxCallType = "staticcall"
	InnerTxCall_Suicide      InnerTxCallType = "suicide"

	InnerTxName_SendCall       InnerTxNameType = "cosmos-send"
	InnerTxName_DelegateCall   InnerTxNameType = "cosmos-delegate"
	InnerTxName_MultiCall      InnerTxNameType = "cosmos-multi-send"
	InnerTxName_UndelegateCall InnerTxNameType = "cosmos-undelegate"
	InnerTxName_EvmCall        InnerTxNameType = "cosmos-call"
	InnerTxName_EvmCreate      InnerTxNameType = "cosmos-create"

	ContractCreateType_Create  = ContractCreateType(InnerTxCall_Create)
	ContractCreateType_Create2 = ContractCreateType(InnerTxCall_Create2)
)

type ContractCreationInfoExport struct {
	Creator         string             `json:"creator"`
	CreateType      ContractCreateType `json:"create_type"`
	NonceOnCreation string             `json:"nonce_on_creation"`
	Salt            string             `json:"salt"`
	CodeHash        string             `json:"code_hash"`
}

type InnerTxInternal struct {
	InnerTxExport

	FromNonce       string
	Create2Salt     string
	Create2CodeHash string
}
type InnerTxExport struct {
	Dept          big.Int         `json:"dept"`
	InternalIndex big.Int         `json:"internal_index"`
	CallType      InnerTxCallType `json:"call_type"`
	Name          InnerTxNameType `json:"name"`
	TraceAddress  string          `json:"trace_address"`
	CodeAddress   string          `json:"code_address"`
	From          string          `json:"from"`
	To            string          `json:"to"`
	Input         string          `json:"input"`
	Output        string          `json:"output"`
	IsError       bool            `json:"is_error"`
	GasUsed       uint64          `json:"gas_used"`
	Value         string          `json:"value"`
	ValueWei      string          `json:"value_wei"`
	Error         string          `json:"error"`
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

func CreateInnerTxInternal(dept int64, from, to string, callType InnerTxCallType, name InnerTxNameType, valueWei *big.Int, err error, nonce, create2Salt, create2CodeHash string) *InnerTxInternal {
	callTx := &InnerTxInternal{
		InnerTxExport: InnerTxExport{
			Dept:     *big.NewInt(dept),
			From:     from,
			IsError:  false,
			To:       to,
			CallType: callType,
			Name:     name,
			ValueWei: valueWei.String(),
		},
		FromNonce:       nonce,
		Create2Salt:     create2Salt,
		Create2CodeHash: create2CodeHash,
	}
	if err != nil {
		callTx.Error = err.Error()
		callTx.IsError = true
	}

	return callTx
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
		if tx.CallType == InnerTxCall_Cosmos && tx.Name == InnerTxName_EvmCreate {
			createType = InnerTxCall_Create
		}
		cci := &ContractCreationInfoExport{
			Creator:         tx.From,
			CreateType:      ContractCreateType(createType),
			NonceOnCreation: tx.FromNonce,
			Salt:            tx.Create2Salt,
			CodeHash:        tx.Create2CodeHash,
		}
		results[tx.To] = cci
	}

	return results
}

func (intx *InnerTxInternal) isCreateContract() bool {
	return intx.CallType == InnerTxCall_Create ||
		intx.CallType == InnerTxCall_Create2 ||
		intx.Name == InnerTxName_EvmCreate
}

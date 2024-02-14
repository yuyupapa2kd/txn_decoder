package core

import "github.com/the-medium/tx-decoder/abis"

type ContainerConn struct {
	AbiDatas []abis.AbiData `json:"abi_datas"`
}

type TxDecoded struct {
	Status uint64 `json:"status"` //txReceipt : receipt.Status
	Nonce  uint64 `json:"nonce"`  //tx: nonce()
	// swag 에서 cannot find type definition: big.Int 라는 오류 발생해서, uint64로 변경함.
	//Block     *big.Int `json:"blockNo"`     //txReceipt : receipt.BlockNumber
	//GasTipCap *big.Int `json:"gas_tip_cap"` //tx: gasTipCap()
	//GasFeeCap *big.Int `json:"gas_fee_cap"` //tx: gasFeeCap()
	Block     string `json:"blockNo"`     //txReceipt : receipt.BlockNumber
	GasTipCap uint64 `json:"gas_tip_cap"` //tx: gasTipCap()
	GasFeeCap uint64 `json:"gas_fee_cap"` //tx: gasFeeCap()
	//Timestamp	uint64			// timestamp는 블럭에서 가져올건데, 이 api 에서는 tx 바로 호출하니깐, 이것까지 주려면 따로 block 호출해야되서 pass, 나중에 블럭기준으로 처리하는 db 만들때 넣어주자.
	From           string `json:"from"`      //tx.from
	To             string `json:"to"`        //tx: to()
	Value          string `json:"value"`     //tx: value()
	TransactionFee uint64 `json:"tx_fee"`    // gasPrice * gasUsed
	GasPrice       uint64 `json:"gas_price"` //tx: gasPrice()
	//GasLimit       *big.Int  `json:"gas_limit"`      // TraceTx 에서 얻을 수 있더라... gasUsed도 나오고, traceTx 에도 나온다면... 이걸 굳이 여기서 보여줄필요가 있나??
	GasUsed          uint64                 `json:"gas_used"`      //txReceipt: receipt.GasUsed
	InputData        []byte                 `json:"input_data"`    // 이게 굳이 필요한지는 기획에서 결정할 사항... 필요 없을 것 같기는 한데...
	ContType         string                 `json:"contract_type"` // erc20, erc721, erc1155, etc
	MethodName       string                 `json:"method_name"`
	DecodedInput     string                 `json:"decoded_input_data"`
	DecodedLogTopics []string               `json:"decoded_log_topics"`
	OutputDataMap    map[string]interface{} `json:"output_data_map"`
	//OutputDataMap string `json:"output_data_map"` // 그냥 string 으로도 전환 가능. rdb 구성 때문에 column 형식이 문제되면 그냥 string 으로 때려 넣는 것도 방법
}

// 이하는 InternalTxn 과 StateDiff 정보를 parsing 에 사용되는 내용
// trace_replayBlockTransactions api 가 반환하는 result 를 담는 용도
type ReplayBlockBody struct {
	JsonRpc string              `json:"jsonrpc"`
	Id      int8                `json:"id"`
	Result  []ReplayBlockResult `json:"result"`
}

type ReplayBlockResult struct {
	OutPut          string      `json:"output"`
	StateDiff       interface{} `json:"stateDiff"`
	Trace           []Trace     `json:"trace"`
	TransactionHash string      `json:"transactionHash"`
	VmTrace         []byte      `json:"vmTrace"`
}

// InternalTxn 관련 내용을 처리하는데 사용
type Trace struct {
	Action       *Action `json:"action"`
	Result       *Result `json:"result,omitempty"`
	SubTraces    int64   `json:"subtraces,omitempty"`
	TraceAddress []int64 `json:"traceAddress,omitempty"`
	Type         string  `json:"type"`
}

type Action struct {
	CallType string `json:"callType"`
	From     string `json:"from"`
	Gas      string `json:"gas"`
	Input    string `json:"input"`
	To       string `json:"to"`
	Value    string `json:"value"`
}

type Result struct {
	GasUsed string `json:"gasUsed"`
	OutPut  string `json:"output"`
}

// trace 필드에서 파싱한 정보들을 담아서 반환하기 위해 사용
type InternalTxn struct {
	TypeTrace string `json:"type_trace"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     string `json:"value"`
	GasLimit  string `json:"gas_limit"`
}

// state_diff 의 storage field 의 값을 decimals 로 변환한 내용을 담을 구조체
type StorageDecimal struct {
	StorageAddress string `json:"storage_address"`
	Before         string `json:"before"`
	After          string `json:"after"`
}

// state_diff 필드에서 파싱한 정보들을 담아서 반환하기 위해 사용
type ParsedStateDiff struct {
	Address         string           `json:"address"`
	BalanceBefore   string           `json:"balance_before"`
	BalanceAfter    string           `json:"balance_after"`
	NonceBefore     string           `json:"nonce_before"`
	NonceAfter      string           `json:"nonce_after"`
	CodeBefore      string           `json:"code_before"`
	CodeAfter       string           `json:"code_after"`
	Storage         interface{}      `json:"storage"`
	StorageDecimals []StorageDecimal `json:"storage_decimal"`
}

// getAdvancedTxnsDataOfBlock api 에서 반환해줄 결과를 담기 위해 사용
type AdvancedTxData struct {
	TxHash       string            `json:"tx_hash"`
	InternalTxns []InternalTxn     `json:"internal_txn"`
	StateDiff    []ParsedStateDiff `json:"state_diff"`
}

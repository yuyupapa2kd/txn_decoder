package resource

import "github.com/the-medium/tx-decoder/core"

type ReqTxDecoded struct {
	TxHash string `uri:"txHash"`
}

type ResTxDecoded struct {
	TxDecoded core.TxDecoded `json:"txDecoded"`
}

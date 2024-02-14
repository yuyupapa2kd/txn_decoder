package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/the-medium/tx-decoder/cfg"
)

func GetAdvancedTxnsDataOfBlock(blockNo string) ([]AdvancedTxData, error) {
	var result []AdvancedTxData
	var advancedTxData AdvancedTxData

	// scrapping block info about internalTxn and stateDiff
	replayBlockDatas, err := getReplayBlockTx(blockNo)
	if err != nil {
		fmt.Println("getReplayBlockTx Error : ", err)
		return result, err
	}
	fmt.Println("successfully get replayBlockData!!!")
	//fmt.Println("replayBlockDatas : ", replayBlockDatas)

	var replayBlockBody ReplayBlockBody
	err = json.Unmarshal(replayBlockDatas, &replayBlockBody)
	if err != nil {
		return result, err
	}

	// loop to parsing replayBlockData
	for _, replayBlockData := range replayBlockBody.Result {
		advancedTxData.TxHash = replayBlockData.TransactionHash

		// parsing replayBlockData for InternalTxn Field
		advancedTxData.InternalTxns, err = parsingCallTypeFromTrace(replayBlockData.Trace)
		if err != nil {
			return result, err
		}

		// parsing replayBlockData for StateDiff Field
		advancedTxData.StateDiff, err = parsingStateDiff(replayBlockData.StateDiff)
		if err != nil {
			return result, err
		}

		result = append(result, advancedTxData)
	}

	return result, nil
}

// scrapping block info about internalTxn and stateDiff
func getReplayBlockTx(blockNo string) ([]byte, error) {
	// 특히 블럭넘버 다음의 [] 가 옵션 부분인데, 이 부분이 핵심임.
	// Ref. https://besu.hyperledger.org/en/stable/public-networks/reference/api/#trace_replayblocktransactions
	json := `{"id": 1, "method": "trace_replayBlockTransactions", "params": ["` + blockNo + `", ["trace", "stateDiff"]]}`
	fmt.Println("Request with:", json)
	jsonByte := []byte(json)
	req, _ := http.NewRequest("POST", cfg.RpcEndpoint, bytes.NewBuffer(jsonByte))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("Request info: " + resp.Status)
	return body, err
}

// parsing replayBlockData for InternalTxn Field
func parsingCallTypeFromTrace(trace []Trace) ([]InternalTxn, error) {
	var internalTxn InternalTxn
	var internalTxns []InternalTxn

	for _, traceData := range trace {
		if traceData.Type == "call" {
			valueBigInt := new(big.Int)
			valueBigInt.SetString(strings.TrimPrefix(traceData.Action.Value, "0x"), 16)
			gasDecimal, err := strconv.ParseInt(strings.TrimPrefix(traceData.Action.Gas, "0x"), 16, 64)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			internalTxn.TypeTrace = traceData.Action.CallType
			internalTxn.From = traceData.Action.From
			internalTxn.To = traceData.Action.To
			internalTxn.Value = valueBigInt.String()
			internalTxn.GasLimit = strconv.FormatInt(gasDecimal, 10)
			internalTxns = append(internalTxns, internalTxn)
		}
	}
	return internalTxns, nil
}

// parsing replayBlockData for StateDiff Field
func parsingStateDiff(stateDiff interface{}) ([]ParsedStateDiff, error) {
	var parsedStateDiff ParsedStateDiff
	var parsedStateDiffs []ParsedStateDiff
	var storageDecimal StorageDecimal
	var nilStorage []StorageDecimal

	// state_diff field 의 내용을 parsing
	// 타입이 * 이면 map[string]map, +, - 이면 map[string]string, = 이면 string 의 3가지여서, typeAssertion 을 사용하여 처리함.
	// Ref. https://besu.hyperledger.org/en/stable/public-networks/reference/trace-types/
	for k, v := range stateDiff.(map[string]interface{}) {

		parsedStateDiff.Address = k
		parsedStateDiff.Storage = v.(map[string]interface{})["storage"]
		storageFeildData := v.(map[string]interface{})["storage"]

		// storage field 를 순회하면서 hex to decimal 작업 진행
		if len(storageFeildData.(map[string]interface{})) > 0 {
			parsedStateDiff.StorageDecimals = nilStorage
			for ks, vs := range storageFeildData.(map[string]interface{}) {
				storageDecimal.StorageAddress = ks
				switch vs.(type) {
				case map[string]interface{}:
					if _, ester := vs.(map[string]interface{})["*"]; ester {
						beforeBigInt := new(big.Int)
						beforeBigInt.SetString(strings.TrimPrefix(vs.(map[string]interface{})["*"].(map[string]interface{})["from"].(string), "0x"), 16)
						storageDecimal.Before = beforeBigInt.String()

						afterBigInt := new(big.Int)
						afterBigInt.SetString(strings.TrimPrefix(vs.(map[string]interface{})["*"].(map[string]interface{})["to"].(string), "0x"), 16)
						storageDecimal.After = afterBigInt.String()
					} else if _, plus := vs.(map[string]interface{})["+"]; plus {
						storageDecimal.Before = ""

						afterBigInt := new(big.Int)
						afterBigInt.SetString(strings.TrimPrefix(vs.(map[string]interface{})["+"].(string), "0x"), 16)
						storageDecimal.After = afterBigInt.String()
					} else if _, minus := vs.(map[string]interface{})["-"]; minus {
						beforeBigInt := new(big.Int)
						beforeBigInt.SetString(strings.TrimPrefix(vs.(map[string]interface{})["-"].(string), "0x"), 16)
						storageDecimal.Before = beforeBigInt.String()

						storageDecimal.After = ""
					}
				case string:
					storageDecimal.Before = ""
					storageDecimal.After = ""

				default:
					fmt.Printf("type of storage is %T\n", vs)
				}
				parsedStateDiff.StorageDecimals = append(parsedStateDiff.StorageDecimals, storageDecimal)
			}
		} else {
			parsedStateDiff.StorageDecimals = nilStorage
		}

		// balance field 를 파싱
		switch stateDiffBalance := v.(map[string]interface{})["balance"].(type) {
		case map[string]interface{}:
			if _, ester := stateDiffBalance["*"]; ester {
				balanceBeforeBigInt := new(big.Int)
				balanceBeforeBigInt.SetString(strings.TrimPrefix(stateDiffBalance["*"].(map[string]interface{})["from"].(string), "0x"), 16)
				parsedStateDiff.BalanceBefore = balanceBeforeBigInt.String()

				balanceAfterBigInt := new(big.Int)
				balanceAfterBigInt.SetString(strings.TrimPrefix(stateDiffBalance["*"].(map[string]interface{})["to"].(string), "0x"), 16)
				parsedStateDiff.BalanceAfter = balanceAfterBigInt.String()
			} else if _, plus := stateDiffBalance["+"]; plus {
				parsedStateDiff.BalanceBefore = ""
				balanceAfterBigInt := new(big.Int)
				balanceAfterBigInt.SetString(strings.TrimPrefix(stateDiffBalance["+"].(string), "0x"), 16)
				parsedStateDiff.BalanceAfter = balanceAfterBigInt.String()

			} else if _, minus := stateDiffBalance["-"]; minus {
				balanceBeforeBigInt := new(big.Int)
				balanceBeforeBigInt.SetString(strings.TrimPrefix(stateDiffBalance["-"].(string), "0x"), 16)
				parsedStateDiff.BalanceBefore = balanceBeforeBigInt.String()
				parsedStateDiff.BalanceAfter = ""
			}
		case string:
			fmt.Println("balance type is string")
			parsedStateDiff.BalanceBefore = ""
			parsedStateDiff.BalanceAfter = ""
		default:
			fmt.Printf("type of balance is %T\n", stateDiffBalance)
		}

		// code field 를 파싱
		switch stateDiffCode := v.(map[string]interface{})["code"].(type) {
		case map[string]interface{}:
			if _, ester := stateDiffCode["*"]; ester {
				parsedStateDiff.CodeBefore = stateDiffCode["*"].(map[string]interface{})["from"].(string)
				parsedStateDiff.CodeAfter = stateDiffCode["*"].(map[string]interface{})["to"].(string)
			} else if _, plus := stateDiffCode["+"]; plus {
				parsedStateDiff.CodeBefore = ""
				parsedStateDiff.CodeAfter = stateDiffCode["+"].(string)
			} else if _, minus := stateDiffCode["-"]; minus {
				parsedStateDiff.CodeBefore = stateDiffCode["-"].(string)
				parsedStateDiff.CodeAfter = ""
			}
		case string:
			fmt.Println("code type is string")
			parsedStateDiff.CodeBefore = ""
			parsedStateDiff.CodeAfter = ""
		default:
			fmt.Println("cannot fix type!!!")
			fmt.Printf("type of code is %T\n", stateDiffCode)
		}

		// nonce field 를 파싱
		switch stateDiffNonce := v.(map[string]interface{})["nonce"].(type) {
		case map[string]interface{}:
			if _, ester := stateDiffNonce["*"]; ester {
				nonceBeforeDecimal, err := strconv.ParseInt(strings.TrimPrefix(stateDiffNonce["*"].(map[string]interface{})["from"].(string), "0x"), 16, 64)
				if err != nil {
					fmt.Println(err)
					return parsedStateDiffs, err
				}
				parsedStateDiff.NonceBefore = strconv.FormatInt(nonceBeforeDecimal, 10)

				nonceAfterDecimal, err := strconv.ParseInt(strings.TrimPrefix(stateDiffNonce["*"].(map[string]interface{})["to"].(string), "0x"), 16, 64)
				if err != nil {
					fmt.Println(err)
					return parsedStateDiffs, err
				}
				parsedStateDiff.NonceAfter = strconv.FormatInt(nonceAfterDecimal, 10)
			} else if _, plus := stateDiffNonce["+"]; plus {
				parsedStateDiff.NonceBefore = ""
				nonceAfterDecimal, err := strconv.ParseInt(strings.TrimPrefix(stateDiffNonce["+"].(string), "0x"), 16, 64)
				if err != nil {
					fmt.Println(err)
					return parsedStateDiffs, err
				}
				parsedStateDiff.NonceAfter = strconv.FormatInt(nonceAfterDecimal, 10)

			} else if _, minus := stateDiffNonce["-"]; minus {
				nonceBeforeDecimal, err := strconv.ParseInt(strings.TrimPrefix(stateDiffNonce["-"].(string), "0x"), 16, 64)
				if err != nil {
					fmt.Println(err)
					return parsedStateDiffs, err
				}
				parsedStateDiff.NonceBefore = strconv.FormatInt(nonceBeforeDecimal, 10)
				parsedStateDiff.NonceAfter = ""
			}
		case string:
			fmt.Println("nonce type is string")
			parsedStateDiff.NonceBefore = ""
			parsedStateDiff.NonceAfter = ""
		default:
			fmt.Println("cannot fix type!!!")
			fmt.Printf("type of nonce is %T\n", stateDiffNonce)

		}

		parsedStateDiffs = append(parsedStateDiffs, parsedStateDiff)
	}
	return parsedStateDiffs, nil
}

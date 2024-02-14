package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/the-medium/tx-decoder/abis"
	"github.com/the-medium/tx-decoder/cfg"
)

// api 에서 호출하는 메서드
// TransactionHash 에 해당하는 Tx 의 기본정보와 InputData 및 LogDetail 을 분석해서 제공
func GetTxDecoded(txHash common.Hash) (TxDecoded, error) {
	var result TxDecoded
	// 블록체인 노드가 제공하는 rpcEnpoint 에 연결된 client object 생성
	fmt.Println("ethclient dialing to rpcEndpoint")
	client, err := ethclient.Dial(cfg.RpcEndpoint)
	if err != nil {
		return result, err
	}
	fmt.Println("ethclient get connection to recEndpoint successfully")

	// on-chain 에서 chainID 조회
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		fmt.Println("err : ", err)
		return result, err
	}
	fmt.Println("chainID : ", chainID)
	fmt.Println("get Tx sequence will start!!")

	// scrapping tx from on-chain
	tx, err := getTransaction(client, txHash)
	if err != nil {
		return result, err
	}
	fmt.Println("successfully get Tx!!!")

	/*
		// from 값을 얻기 위해서는 msg 를 꺼내야 하는데, tx 에 따라서 유형이 다르다는 오류 발생
		// 아마도 types.NewEIP155Signer() 를 다른 것으로 바꾸면 될것 같기는 한데,
		// 현재 기능에서 중요한 부분이 아니라 그냥 skip 하였음.
		msg, err := tx.AsMessage(types.NewEIP155Signer(chainID), big.NewInt(0))
		if err != nil {
			return result, err
		}
	*/

	// scrapping txReceipt from on-chain
	fmt.Println("get TxReceipt swquence will start!!")
	txReceipt, err := getTransactionReceipt(client, txHash)
	if err != nil {
		return result, err
	}
	fmt.Println("successfully get TxReceipt!!!")
	fmt.Println("Injecting data to result struct")

	// parsing tx datas to result struct
	result.Status = txReceipt.Status
	result.Nonce = tx.Nonce()
	result.Block = txReceipt.BlockNumber.String()
	//result.From = msg.From().Hex()
	result.To = tx.To().Hex()
	result.Value = tx.Value().String()
	result.InputData = tx.Data()
	result.GasPrice = tx.GasPrice().Uint64()
	result.GasUsed = txReceipt.GasUsed
	result.GasTipCap = tx.GasTipCap().Uint64()
	result.GasFeeCap = tx.GasFeeCap().Uint64()
	result.TransactionFee = uint64(result.GasUsed * result.GasPrice)
	fmt.Println("get abi sequence will start")

	// abiContainer 에서 필요한 abi 값을 contract address 로 조회
	abiJson, contType, err := abis.GetABIfromContractAddress(result.To)
	if err != nil {
		errAbiMsg := "The following error occurred while using abi. Please check the registered abi.json!!! Error :" + string(err.Error())
		return result, errors.New(errAbiMsg)
	} else if contType == "" {
		nilAbiMsg := "There is no such contract address in abi DB"
		return result, errors.New(nilAbiMsg)
	}
	result.ContType = contType
	fmt.Println("successfully get abi!!!")

	// inputData 필드 decoding 해서 methodName 가 decoding 된 value 제공
	fmt.Println("decoding TxInputData sequence will start")
	methodName, decodedInput, err := decodeTransactionInputData(abiJson, tx.Data())
	if err != nil {
		return result, err
	}
	fmt.Println("successfully decoding TxInputData!!!")
	result.MethodName = methodName
	result.DecodedInput = decodedInput
	//서비스개발 쪽에서 erc20 의 Transfer 작업할 때, 이거 말고 log topics 의 내용 쓰도록 가이드!
	//형식 보면 이놈은 대소문자 구별 못하고 다 소문자 만들어버림. 문제 생길 수 있음.
	fmt.Println("methodName : ", result.MethodName)
	fmt.Println("DecodedInput : ", result.DecodedInput)
	fmt.Println("decoding TxLogs sequence will start")
	//fmt.Println("txReceipt.Logs : ", txReceipt.Logs)

	// txReceipt.logs 의 내용을 decoding 해서 logTopics 과 decoding 된 value 제공
	logTopics, outputDataMap, err := decodeTransactionLogs(txReceipt)
	if err != nil {
		return result, err
	}
	fmt.Println("successfully decoding TxLogs!!!")
	result.DecodedLogTopics = logTopics
	result.OutputDataMap = outputDataMap

	/*
		// erc20 등에 대한 정보제공은 이 데이터를 받아서 처리하는 explorer 쪽에서 진행하는 것이 흐름상 자연스러울듯...
		// 대신 type 은 제공해줘서 그거 기준으로 바로 작업할 수 있도록 처리.
		if checkTypeErc20(contType) && checkMethodTransfer(methodName) {
			result.Erc20Trnf = append(result.Erc20Trnf, logTopics[1])
			result.Erc20Trnf = append(result.Erc20Trnf, logTopics[2])
			//result.Erc20Trnf = append(result.Erc20Trnf, outputDataMap.Value)
		} else if checkTypeErc721(contType) && checkMethodTransfer(methodName) {
			result.Erc721Trnf = append(result.Erc721Trnf, logTopics[1])
			result.Erc721Trnf = append(result.Erc721Trnf, logTopics[2])
			//result.Erc721Trnf = append(result.Erc721Trnf, outputDataMap.value)
		} else if checkTypeErc1155(contType) && checkMethodTransfer(methodName) {
			result.Erc1155Trnf = append(result.Erc1155Trnf, logTopics[1])
			result.Erc1155Trnf = append(result.Erc1155Trnf, logTopics[2])
			//result.Erc1155Trnf = append(result.Erc1155Trnf, outputDataMap.value)
		}
	*/
	return result, nil
}

// get Tx
func getTransaction(client *ethclient.Client, txHash common.Hash) (*types.Transaction, error) {
	tx, _, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return tx, nil
}

// get TxReceipt
func getTransactionReceipt(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return receipt, nil
}

// decoding InputData from Tx
func decodeTransactionInputData(contractABI abi.ABI, data []byte) (string, string, error) {
	// The first 4 bytes of the t represent the ID of the method in the ABI
	// https://docs.soliditylang.org/en/v0.5.3/abi-spec.html#function-selector
	methodSigData := data[:4]
	method, err := contractABI.MethodById(methodSigData)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	// methodName 을 추출한 뒤쪽의 bytes 를 형식에 맞게 쪼개서 map 으로 생성
	// UnpackIntoMap 은 메서드로 선언되어 있음. 해당 내용은 아래의 링크에서 확인 가능
	// https://github.com/ethereum/go-ethereum/blob/master/accounts/abi/abi.go line.128, line.84
	inputsSigData := data[4:]
	inputsMap := make(map[string]interface{})
	if err := method.Inputs.UnpackIntoMap(inputsMap, inputsSigData); err != nil {
		fmt.Println(err)
		return "", "", err
	}

	// string 변환을 위해 map 을 보기 편한 json 으로 marshaling
	inputsJson, err := json.Marshal(inputsMap)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	return method.Name, string(inputsJson), nil
}

// decoding Tx Logs from TxReceipt
func decodeTransactionLogs(receipt *types.Receipt) ([]string, map[string]interface{}, error) {
	var decodedLogTopics []string
	var outputDataMap map[string]interface{}
	if receipt.Logs == nil {
		err := errors.New("There is no Log data in this TxReceipt")
		return nil, nil, err
	}

	for _, vLog := range receipt.Logs {
		// get abi for log
		logAddress := vLog.Address

		// log 에서 to 로 세팅되어 있는 contract address 로 abi 조회
		abiJson, contType, err := abis.GetABIfromContractAddress(logAddress.Hex())
		if err != nil {
			errAbiMsg := "The following error occurred while using abi. Please check the registered abi.json!!! Error :" + string(err.Error())
			return nil, nil, errors.New(errAbiMsg)
		} else if contType == "" {
			nilAbiMsg := "There is no such contract address in abi DB"
			return nil, nil, errors.New(nilAbiMsg)
		}

		// topic[0] is the event name
		// Ref. https://github.com/ethereum/go-ethereum/blob/master/accounts/abi/abi.go line. 202
		event, err := abiJson.EventByID(vLog.Topics[0])
		if err != nil {
			fmt.Println("contractABI.EventByID Error : ", err)
			return nil, nil, err
		}
		decodedLogTopics = append(decodedLogTopics, event.Name)
		fmt.Printf("Event Name: %s\n", event.Name)

		// topic[1:] is other indexed params in event
		if len(vLog.Topics) > 1 {
			fmt.Println("length of log topics")
			for i, param := range vLog.Topics[1:] {
				fmt.Printf("Indexed params %d in hex: %s\n", i, param)
				fmt.Printf("Indexed params %d decoded %s\n", i, common.HexToAddress(param.Hex()))
				decodedLogTopics = append(decodedLogTopics, common.HexToAddress(param.Hex()).String())
			}
		}
		fmt.Println("decoding log topics sequence was successfully clear!!")

		if len(vLog.Data) > 0 {
			fmt.Println("unpack log datas squence will be start")
			outputDataMapResource := make(map[string]interface{})
			err = abiJson.UnpackIntoMap(outputDataMapResource, event.Name, vLog.Data)
			if err != nil {
				fmt.Println("contractABI.UnpackUntoMap Error : ", err)
				return nil, nil, err
			}
			//fmt.Printf("Event outputs: %v\n", outputDataMapResource)
			outputDataMap = outputDataMapResource
		}
	}
	return decodedLogTopics, outputDataMap, nil
}

/*
func checkTypeErc20(contType string) bool {
	if contType == "erc20" {
		return true
	}
	return false
}

func checkTypeErc721(contType string) bool {
	if contType == "erc721" {
		return true
	}
	return false
}

func checkTypeErc1155(contType string) bool {
	if contType == "erc1155" {
		return true
	}
	return false
}

func checkMethodTransfer(methodName string) bool {
	if methodName == "transfer" {
		return true
	}
	return false
}
*/

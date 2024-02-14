package abis

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func StoreAbiToABIContainer(filePath string) error {
	// json 파일 읽기
	jsonBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	fmt.Println("store abi to container, successfully read file from ", filePath)

	// json parsing
	var abiData AbiData
	abiData.AbiJson = string(jsonBytes)
	filename := filepath.Base(filePath)
	fileNameWithoutExtension := strings.TrimSuffix(filename, filepath.Ext(filename))
	fmt.Println("filename without extension : ", fileNameWithoutExtension)
	splitFileName := strings.Split(fileNameWithoutExtension, "_")
	abiData.AbiType = splitFileName[0]
	abiData.ContractAddress = splitFileName[1]
	fmt.Println("abiData.ContractAddress : ", abiData.ContractAddress)
	fmt.Println("abiData.AbiType : ", abiData.AbiType)

	PublicAbiDatas = append(PublicAbiDatas, abiData)

	return nil
}

func GetABIfromContractAddress(address string) (abi.ABI, string, error) {
	var contractABI abi.ABI
	var contType string

	for _, abiData := range PublicAbiDatas {
		if abiData.ContractAddress == address {
			contractABI, err := abi.JSON(strings.NewReader(abiData.AbiJson))
			if err != nil {
				fmt.Println(err)
				return contractABI, contType, err
			}
			contType = abiData.AbiType
			return contractABI, contType, nil
		}
	}
	fmt.Println("There is no such contract address in ABIContainer")
	return contractABI, contType, nil
}

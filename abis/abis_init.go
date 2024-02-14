package abis

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/the-medium/tx-decoder/cfg"
)

var abiFilePath = "tx-decoder/abis/registered"

type AbiData struct {
	ContractAddress string `json:"address"`
	AbiJson         string `json:"abi_json"`
	AbiType         string `json:"type"` //erc20, erc721, erc1155, etc
}

type ABIContainer struct {
	AbiDatas []AbiData `json:"abi_datas"`
}

// abis 패키지와 core 패키지에서 사용할 abi 관련 내용이 저장될 변수
var PublicAbiDatas []AbiData

func InitABIContainer() *ABIContainer {
	fmt.Println("InitABIContainer process")
	return &ABIContainer{AbiDatas: PublicAbiDatas}
}

func StoreDefaultABIToABIContABIContainer() error {
	folderPath := filepath.Join(cfg.ProjectFolder, abiFilePath)
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return err
	}
	fmt.Println("read dir for store default abi was done")
	for _, file := range files {
		filePath := filepath.Join(folderPath, file.Name())
		fmt.Println("ready to store abi to container with ", filePath)

		err = StoreAbiToABIContainer(filePath)
		if err != nil {
			return err
		}
	}
	return nil
}

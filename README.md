# tx-decoder

## How To Setting
### 1. ./cfg/cfg.go 파일의 변수들을 배포 상황에 맞게 설정하고 저장
### 2. ./abis/registered 폴더에 abi 파일들을 저장
#### 1) abi 파일의 형식 : {type}_{contractAddress}.json
#### 2) type : erc-20 이면 "erc20", erc-721 이면 "erc721", erc-1155 이면 "erc1155", 어디에도 해당하지 않으면 "etc"로 설정
### 3. ./go build ./main.go 로 build 수행
### 4. ./main 으로 서버 실행

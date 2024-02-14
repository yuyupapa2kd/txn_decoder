package api

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/ethereum/go-ethereum/common"
	"github.com/the-medium/tx-decoder/core"
	"github.com/the-medium/tx-decoder/docs"
)

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func setupSwagger(r *gin.Engine) {

	localAddr := getOutboundIP()
	IPNPortString := localAddr.String() + ":8080"

	docs.SwaggerInfo.Title = "GNDChain Tx Decoder API"
	docs.SwaggerInfo.Description = "This is a Tool for developer on GNDChain."
	docs.SwaggerInfo.Version = "0.1"
	docs.SwaggerInfo.Host = IPNPortString
	docs.SwaggerInfo.BasePath = ""
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})

	swaggerUrlString := "http://" + IPNPortString + "/swagger/doc.json"
	fmt.Println("swaggerUrlString : ", swaggerUrlString)
	url := ginSwagger.URL(swaggerUrlString)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
}

func SetRouter() *gin.Engine {
	r := gin.Default()

	setupSwagger(r)

	r.GET("/api/txDecoder/:txHash", getTxDecoded)
	r.GET("/api/advancedTxnsDataOfBlock/:blockNo", getAdvancedTxnsDataOfBlock)

	return r
}

// TxDecoder	godoc
// @Tags         TxDecoder
// @Summary      Get decoded information about Tx
// @Description  Get decoded information of Transacion : {txHash}
// @Produce      json
// @Param        txHash  path      string  true  "txHash"
// @Success      200      {object}  resource.ResJSON{data=core.TxDecoded}
// @Failure      400      {object}  resource.ResJSON{data=resource.ResErr}
// @Router       /api/txDecoder/{txHash} [get]
func getTxDecoded(c *gin.Context) {
	txHashStr := c.Param("txHash")
	txHash := common.HexToHash(txHashStr)
	fmt.Println("param : ", txHash)
	fmt.Println("GetTxDecoded Method is progressing...")
	var res core.TxDecoded
	res, err := core.GetTxDecoded(txHash)
	if err != nil {
		c.JSON(400, gin.H{"result": res, "error": string(err.Error())})
		return
	}
	c.JSON(200, gin.H{"result": res, "error": ""})
}

// TxDecoder	godoc
// @Tags         GetAdvancedTxnsDataOfBlock
// @Summary      Get Advanced information about Txns
// @Description  Get InternalTxn and StateDiff Information about Txns of blockNo {blockNo}
// @Produce      json
// @Param        blockNo  path      string  true  "blockNo"
// @Success      200      {object}  resource.ResJSON{data=[]core.AdvancedTxData}
// @Failure      400      {object}  resource.ResJSON{data=resource.ResErr}
// @Router       /api/advancedTxnsDataOfBlock/{blockNo} [get]
func getAdvancedTxnsDataOfBlock(c *gin.Context) {
	blockNo := c.Param("blockNo")
	fmt.Println("param : ", blockNo)
	fmt.Println("GetAdvancedTxData Method is progressing...")
	var res []core.AdvancedTxData
	res, err := core.GetAdvancedTxnsDataOfBlock(blockNo)
	if err != nil {
		c.JSON(400, gin.H{"result": res, "error": string(err.Error())})
		return
	}
	c.JSON(200, gin.H{"result": res, "error": ""})
}

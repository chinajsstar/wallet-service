package handler

import (
	"net/http"
	l4g "github.com/alecthomas/log4go"
	"blockchain_server/service"
	"bastionpay_tools/function"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

////////////////////////////////////////////////////////////////////////////////
type WebRes struct {
	Err int 		`json:"err"`
	ErrMsg string 	`json:"errmsg"`
	Value string 	`json:"value"`
}

type Web struct{
	*function.Functions

	filesMap map[string]string
}

func (self *Web)Init(clientManager *service.ClientManager, dataDir string) error {
	// functions
	self.Functions = &function.Functions{}
	err := self.Functions.Init(clientManager, dataDir)
	if err != nil {
		return err
	}

	self.filesMap = make(map[string]string)

	return nil
}

// start http server
func (self *Web) StartHttpServer(port string) error {
	// http
	l4g.Info("Start http server on: %s", port)

	engine := gin.Default()
	engine.Use(cors.New(cors.Config{
		AllowAllOrigins:true,
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Authorization", "X-Requested-With", "X_Requested_With", "Content-Type", "Access-Token", "Accept-Language"},
		//AllowOrigins:     []string{"*"},
		//AllowCredentials: true,
		//AllowOriginFunc: func(origin string) bool {
		//	return true;//origin == "https://github.com"
		//},
		//MaxAge: 12 * time.Hour,
	}))

	//router := engine.Group("/offline", func(ctx *gin.Context) {
	//
	//})

	//engine.GET("/", self.handle404)

	engine.Static("", "template")
	//router.Static("/js", "template")
	//
	//var path string
	//path = "index"
	//router.Static("/" + path, "template/html/" + path + ".html")
	//
	//path = "newaddress"
	//router.Static("/" + path, "template/html/" + path + ".html")
	//
	//path = "signtx"
	//router.Static("/" + path, "template/html/" + path + ".html")

	engine.POST("/newaddressact", self.handleNewAddressAct)
	engine.POST("/signtxact", self.handleSigntxAct)

	engine.Run(":" + port)

	return nil
}

func (self *Web) handle404(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, "/index")
}

func (self *Web) handleNewAddressAct(ctx *gin.Context) {
	rb := WebRes{Err:1, ErrMsg:""}

	err := func() error {
		coinType := ctx.PostForm("cointype")
		count := ctx.PostForm("count")
		newAddressSaveDir := ctx.PostForm("newaddresssavedir")

		// new address
		l4g.Info("newaddress: %s-%s-%s", coinType, count,  newAddressSaveDir)

		c, err := strconv.Atoi(count)
		if err != nil {
			return err
		}

		dstUniDir, err := self.NewAddress(coinType, uint32(c), newAddressSaveDir)
		if err != nil {
			return err
		}

		rb.Err = 0
		rb.Value = "Db file save as：" + dstUniDir
		return nil
	}()

	if err != nil {
		rb.Err = 1
		rb.ErrMsg = err.Error()
		l4g.Error("handleNewAddressAct: %s", err.Error())
	}

	ctx.JSON(http.StatusOK, rb)
	return
}

func (self *Web) handleSigntxAct(ctx *gin.Context) {
	rb := WebRes{Err:1, ErrMsg:""}

	err := func()error{
		txFilePath := ctx.PostForm("txfilepath")
		txSignedSaveDir := ctx.PostForm("txsignedsavedir")

		// signtx
		l4g.Info("signtx: %s-%s", txFilePath, txSignedSaveDir)

		uniPath, err := self.SignTx(txFilePath, txSignedSaveDir)
		if err != nil {
			return err
		}

		rb.Err = 0
		rb.Value = "txsigned save as：" + uniPath
		return nil
	}()

	if err != nil {
		rb.Err = 1
		rb.ErrMsg = err.Error()
		l4g.Error("handleSigntxAct: %s", err.Error())
	}

	ctx.JSON(http.StatusOK, rb)

	return
}
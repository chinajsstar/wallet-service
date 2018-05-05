package handler

import (
	"net/http"
	"html/template"
	l4g "github.com/alecthomas/log4go"
	"encoding/json"
	"blockchain_server/service"
	"bastionpay_tools/function"
	"strconv"
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

	http.Handle("/",http.HandlerFunc(self.handle404))
	http.Handle("/css/", http.FileServer(http.Dir("template")))
	http.Handle("/js/", http.FileServer(http.Dir("template")))

	var path string
	// files
	path = "index"
	self.filesMap["/" + path] = path + ".html"
	http.Handle("/" + path,http.HandlerFunc(self.handlePathFile))

	path = "newaddress"
	self.filesMap["/" + path] = path + ".html"
	http.Handle("/" + path,http.HandlerFunc(self.handlePathFile))

	path = "signtx"
	self.filesMap["/" + path] = path + ".html"
	http.Handle("/" + path,http.HandlerFunc(self.handlePathFile))

	// act post
	http.Handle("/newaddressact", http.HandlerFunc(self.handleNewAddressAct))
	http.Handle("/signtxact", http.HandlerFunc(self.handleSigntxAct))

	go func() {
		l4g.Info("Http server routine running... ")
		err := http.ListenAndServe(":" + port, nil)
		if err != nil {
			l4g.Crashf("", err)
		}
	}()

	return nil
}

func (self *Web) handle404(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/index", http.StatusFound)
	}

	t, err := template.ParseFiles("template/html/404.html")
	if (err != nil) {
		l4g.Error("%s", err.Error())
		w.Write([]byte(err.Error()))
		return
	}

	t.Execute(w, nil)
}

func (self *Web) handlePathFile(w http.ResponseWriter, req *http.Request) {
	filename, ok := self.filesMap[req.URL.Path]
	if ok == false {
		filename = "404.html"
	}

	t, err := template.ParseFiles("template/html/" + filename)
	if err != nil {
		l4g.Error("%s", err.Error())
		w.Write([]byte(err.Error()))
		return
	}

	t.Execute(w, nil)
}

func (self *Web) handleNewAddressAct(w http.ResponseWriter, req *http.Request) {
	rb := WebRes{Err:1, ErrMsg:""}

	err := func() error {
		coinType := req.FormValue("cointype")
		count := req.FormValue("count")
		newAddressSaveDir := req.FormValue("newaddresssavedir")
		//newAddressBackDir := req.FormValue("newaddressbackdir")

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

	b, _ := json.Marshal(rb)
	w.Write(b)
	return
}

func (self *Web) handleSigntxAct(w http.ResponseWriter, req *http.Request) {
	rb := WebRes{Err:1, ErrMsg:""}

	err := func()error{
		txFilePath := req.FormValue("txfilepath")
		txSignedSaveDir := req.FormValue("txsignedsavedir")

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

	b, _ := json.Marshal(rb)
	w.Write(b)
	return
}
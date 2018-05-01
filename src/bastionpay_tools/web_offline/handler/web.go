package handler

import (
	"net/http"
	"html/template"
	l4g "github.com/alecthomas/log4go"
	"fmt"
	"encoding/json"
	"os"
	"io"
	"bastionpay_tools/db"
	"blockchain_server/service"
	"bastionpay_tools/function"
	"strconv"
)

// copy file
func CopyFile(src, dst string)(w int64, err error){
	srcFile,err := os.Open(src)
	if err!=nil{
		fmt.Println(err.Error())
		return
	}
	defer srcFile.Close()

	dstFile,err := os.Create(dst)
	if err!=nil{
		fmt.Println(err.Error())
		return
	}

	defer dstFile.Close()

	return io.Copy(dstFile,srcFile)
}

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
			return
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
		return
	}

	t.Execute(w, nil)
	return
}

func (self *Web) handleNewAddressAct(w http.ResponseWriter, req *http.Request) {
	rb := WebRes{Err:1, ErrMsg:""}

	err := func() error {
		cointype := req.FormValue("cointype")
		count := req.FormValue("count")
		newaddresssavedir := req.FormValue("newaddresssavedir")
		newaddressfilepath := ""

		// new address
		fmt.Println(cointype, "--", count, "--", newaddresssavedir)

		c, err := strconv.Atoi(count)
		if err != nil {
			return err
		}

		uniName, err := self.NewAddress(cointype, uint32(c))
		if err != nil {
			return err
		}else{
			onlineDBPath := self.GetAddressDataDir() + "/" + db.GetOnlineUniDBName(uniName)
			newaddressfilepath = newaddresssavedir + "/" + db.GetOnlineUniDBName(uniName)
			_, err = CopyFile(onlineDBPath, newaddressfilepath)
			if err != nil {
				return err
			}
		}

		rb.Err = 0
		rb.Value = "Db file save as：" + newaddressfilepath
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
		txfilepath := req.FormValue("txfilepath")
		txsignedfilepath := req.FormValue("txsignedfilepath")

		// signtx
		fmt.Println(txfilepath, "--", txsignedfilepath)

		err := self.SignTx(txfilepath, txsignedfilepath)
		if err != nil {
			return err
		}

		rb.Err = 0
		rb.Value = "txsigned save ad：" + txsignedfilepath
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
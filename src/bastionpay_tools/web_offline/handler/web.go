package handler

import (
	"net/http"
	"html/template"
	l4g "github.com/alecthomas/log4go"
	"fmt"
	"encoding/json"
	"bastionpay_tools/tools"
	"bastionpay_tools/handler"
	"os"
	"io"
	"bastionpay_tools/db"
)

////////////////////////////////////////////////////////////////////////////////
type Web struct{
	offlineTool *tools.OffLine
}

func NewWeb() *Web {
	w := &Web{}
	return w
}

func (self *Web)Init(ol *tools.OffLine) error {
	if err := self.startHttpServer(); err != nil{
		return err
	}

	self.offlineTool = ol
	return nil
}

// start http server
func (self *Web) startHttpServer() error {
	// http
	l4g.Debug("Start http server on 8066")

	http.Handle("/",http.HandlerFunc(self.handle404))
	http.Handle("/index",http.HandlerFunc(self.handleIndex))
	http.Handle("/newaddress", http.HandlerFunc(self.handleNewAddress))
	http.Handle("/signtx", http.HandlerFunc(self.handleSigntx))

	http.Handle("/newaddressact", http.HandlerFunc(self.handleNewAddressAct))
	http.Handle("/signtxact", http.HandlerFunc(self.handleSigntxAct))

	http.Handle("/css/", http.FileServer(http.Dir("template")))
	http.Handle("/js/", http.FileServer(http.Dir("template")))

	go func() {
		l4g.Info("Http server routine running... ")
		err := http.ListenAndServe(":8066", nil)
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

// http handler
func (self *Web) handleIndex(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles("template/html/index.html")
	if err != nil {
		return
	}

	t.Execute(w, nil)
	return
}

type ResBack struct {
	Err int `json:"err"`
	ErrMsg string `json:"errmsg"`
}

// http handler
func (self *Web) handleNewAddress(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles("template/html/newaddress.html")
	if err != nil {
		return
	}

	t.Execute(w, nil)
	return
}

func CopyFile(src,dst string)(w int64, err error){
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

// http handler
func (self *Web) handleNewAddressAct(w http.ResponseWriter, req *http.Request) {
	cointype := req.FormValue("cointype")
	count := req.FormValue("count")
	newaddressdir := req.FormValue("newaddressdir")
	newaddressfilepath := ""

	// new address
	fmt.Println(cointype, "--", count, "--", newaddressdir)

	var argv []string
	argv = append(argv, "newaddress")
	argv = append(argv, cointype)
	argv = append(argv, count)
	res, err := self.offlineTool.Execute(argv)

	rb := ResBack{}
	if err != nil {
		rb.Err = 1
		rb.ErrMsg = err.Error()
	}else{
		naRes := handler.NewAddressRes{}
		err = json.Unmarshal([]byte(res), &naRes)
		if err == nil {
			newaddressfilepath = newaddressdir + "/" + db.GetOnlineUniDBName(naRes.UniName)
			_, err = CopyFile(naRes.OnlineDBPath, newaddressfilepath)
		}
	}

	if err != nil {
		rb.Err = 1
		rb.ErrMsg = err.Error()
	}else{
		rb.ErrMsg = "文件保存在：" + newaddressfilepath
	}

	b, _ := json.Marshal(rb)
	w.Write(b)

	return
}

// http handler
func (self *Web) handleSigntx(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles("template/html/signtx.html")
	if err != nil {
		return
	}

	t.Execute(w, nil)
	return
}

// http handler
func (self *Web) handleSigntxAct(w http.ResponseWriter, req *http.Request) {
	txfilepath := req.FormValue("txfilepath")
	txsignedfilepath := req.FormValue("txsignedfilepath")

	// signtx
	fmt.Println(txfilepath, "--", txsignedfilepath)
	var argv []string
	argv = append(argv, "signtx")
	argv = append(argv, txfilepath)
	argv = append(argv, txsignedfilepath)
	_, err := self.offlineTool.Execute(argv)

	rb := ResBack{}
	if err != nil {
		rb.Err = 1
		rb.ErrMsg = err.Error()
	}else{
		rb.ErrMsg = "文件保存在：" + txsignedfilepath
	}

	b, _ := json.Marshal(rb)
	w.Write(b)
	return
}
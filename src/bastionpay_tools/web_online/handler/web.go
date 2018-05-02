package handler

import (
	"net/http"
	"html/template"
	l4g "github.com/alecthomas/log4go"
	"fmt"
	"encoding/json"
	"os"
	"io"
	"blockchain_server/service"
	"bastionpay_tools/function"
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

	// 下载交易文件
	http.Handle("/data/", http.FileServer(http.Dir(".")))

	var path string
	// files
	path = "index"
	self.filesMap["/" + path] = path + ".html"
	http.Handle("/" + path,http.HandlerFunc(self.handlePathFile))

	path = "uploadaddress"
	self.filesMap["/" + path] = path + ".html"
	http.Handle("/" + path,http.HandlerFunc(self.handlePathFile))

	path = "uploadtx"
	self.filesMap["/" + path] = path + ".html"
	http.Handle("/" + path,http.HandlerFunc(self.handlePathFile))

	path = "uploadsignedtx"
	self.filesMap["/" + path] = path + ".html"
	http.Handle("/" + path,http.HandlerFunc(self.handlePathFile))

	// 上传地址文件
	http.Handle("/uploadaddressact", http.HandlerFunc(self.handleUploadAddressAct))

	// 上传交易文件
	http.Handle("/uploadtxact", http.HandlerFunc(self.handleUploadTxAct))

	// 上传签名交易文件
	http.Handle("/uploadsignedtxact", http.HandlerFunc(self.handleUploadSignedtxAct))

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

func (self *Web) handleUploadAddressAct(w http.ResponseWriter, req *http.Request) {
	rb := WebRes{Err:1, ErrMsg:""}

	err := func() error {
		req.ParseMultipartForm(32 << 20)
		file, handler, err := req.FormFile("uploadfile")
		if err != nil {
			return err
		}
		defer file.Close()

		//fmt.Fprintf(w, "%v", handler.Header)

		//saveFilePath := "./data/" + common.AddressDirName + handler.Filename
		saveFilePath := self.GetAddressDataDir() + "/" + handler.Filename
		f, err := os.OpenFile(saveFilePath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		rb.Err = 0
		rb.Value = "Address db file upload ok!"
		return nil
	}()

	if err != nil {
		rb.Err = 1
		rb.ErrMsg = err.Error()
		l4g.Error("handleUploadAddressAct: %s", err.Error())
	}

	b, _ := json.Marshal(rb)
	w.Write(b)
	return
}

func (self *Web) handleUploadTxAct(w http.ResponseWriter, req *http.Request) {
	rb := WebRes{Err:1, ErrMsg:""}

	err := func() error {
		req.ParseMultipartForm(32 << 20)
		file, handler, err := req.FormFile("uploadfile")
		if err != nil {
			return err
		}
		defer file.Close()

		//fmt.Fprintf(w, "%v", handler.Header)

		//saveFilePath := "./data/" + common.TxDirName + handler.Filename
		//fmt.Println(handler.Filename)
		saveFilePath := self.GetTxDataDir() + "/" + handler.Filename
		f, err := os.OpenFile(saveFilePath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		rb.Err = 0
		rb.Value = "tx file upload ok!"
		return nil
	}()

	if err != nil {
		rb.Err = 1
		rb.ErrMsg = err.Error()
		l4g.Error("handleUploadTxAct: %s", err.Error())
	}

	b, _ := json.Marshal(rb)
	w.Write(b)
	return
}

func (self *Web) handleUploadSignedtxAct(w http.ResponseWriter, req *http.Request) {
	rb := WebRes{Err:1, ErrMsg:""}

	err := func() error {
		req.ParseMultipartForm(32 << 20)
		file, handler, err := req.FormFile("uploadfile")
		if err != nil {
			return err
		}
		defer file.Close()

		//fmt.Fprintf(w, "%v", handler.Header)

		//saveFilePath := "./data/" + common.TxDirName + handler.Filename
		saveFilePath := self.GetTxDataDir() + "/" + handler.Filename
		f, err := os.OpenFile(saveFilePath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		rb.Err = 0
		rb.Value = "tx file upload ok!"
		return nil
	}()

	if err != nil {
		rb.Err = 1
		rb.ErrMsg = err.Error()
		l4g.Error("handleUploadSignedtxAct: %s", err.Error())
	}

	b, _ := json.Marshal(rb)
	w.Write(b)
	return
}
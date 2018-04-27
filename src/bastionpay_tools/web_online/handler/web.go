package handler

import (
	"net/http"
	"html/template"
	l4g "github.com/alecthomas/log4go"
	"fmt"
	"encoding/json"
	"bastionpay_tools/tools"
	"os"
	"io"
	"github.com/satori/go.uuid"
	"time"
	"crypto/md5"
	"strconv"
)

////////////////////////////////////////////////////////////////////////////////
type Web struct{
	onlineTool *tools.OnLine
}

func NewWeb() *Web {
	w := &Web{}
	return w
}

func (self *Web)Init(ol *tools.OnLine) error {
	if err := self.startHttpServer(); err != nil{
		return err
	}

	self.onlineTool = ol
	return nil
}

// start http server
func (self *Web) startHttpServer() error {
	// http
	l4g.Debug("Start http server on 8055")

	http.Handle("/",http.HandlerFunc(self.handle404))
	http.Handle("/index",http.HandlerFunc(self.handleIndex))

	// 生成交易
	http.Handle("/buildtx", http.HandlerFunc(self.handleBuildtx))
	http.Handle("/buildtxact", http.HandlerFunc(self.handleBuildtxAct))

	// 下载交易文件
	http.Handle("/data/", http.FileServer(http.Dir(".")))

	// 上传签名交易
	http.Handle("/uploadsignedtx", http.HandlerFunc(self.handleUploadSignedtx))
	http.Handle("/uploadsignedtxact", http.HandlerFunc(self.handleUploadSignedtxAct))

	// 发送签名交易
	http.Handle("/sendsignedtx", http.HandlerFunc(self.handleSendSignedtx))
	http.Handle("/sendsignedtxact", http.HandlerFunc(self.handleSendSignedtxAct))

	//http.Handle("/newaddressact", http.HandlerFunc(self.handleNewAddressAct))
	//http.Handle("/signtxact", http.HandlerFunc(self.handleSigntxAct))

	http.Handle("/css/", http.FileServer(http.Dir("template")))
	http.Handle("/js/", http.FileServer(http.Dir("template")))

	go func() {
		l4g.Info("Http server routine running... ")
		err := http.ListenAndServe(":8055", nil)
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
func (self *Web) handleBuildtx(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles("template/html/buildtx.html")
	if err != nil {
		return
	}

	t.Execute(w, nil)
	return
}

// http handler
func (self *Web) handleBuildtxAct(w http.ResponseWriter, req *http.Request) {
	cointype := req.FormValue("cointype")
	chiperprikey := req.FormValue("chiperprikey")
	fromaddr := req.FormValue("fromaddr")
	toaddr := req.FormValue("toaddr")
	count := req.FormValue("count")

	// fmt.Println("正确格式：buildtxcmd 类型 加密私钥 从地址 去地址 数量 交易文件路径")
	buildtxdir := self.onlineTool.GetDataDir()
	// uuid
	uniName, err := func()(string, error) {
		uuidv4, err := uuid.NewV4()
		if err != nil {
			return "", err
		}
		uuid := uuidv4.String()

		datetime := time.Now().UTC().Format(time.RFC3339)
		return datetime + uuid, nil
	}()

	buildtxfilepath := buildtxdir + "/" + uniName + ".tx"

	var argv []string
	argv = append(argv, "buildtxcmd")
	argv = append(argv, cointype)
	argv = append(argv, chiperprikey)
	argv = append(argv, fromaddr)
	argv = append(argv, toaddr)
	argv = append(argv, count)
	argv = append(argv, buildtxfilepath)
	_, err = self.onlineTool.Execute(argv)

	rb := ResBack{}
	if err != nil {
		rb.Err = 1
		rb.ErrMsg = err.Error()
	}else{
		rb.ErrMsg = "操作成功， 请到下载页面下载交易文件"
	}

	b, _ := json.Marshal(rb)
	w.Write(b)

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
func (self *Web) handleUploadSignedtx(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles("template/html/uploadsignedtx.html")
	if err != nil {
		return
	}

	t.Execute(w, nil)
	return
}

// http handler
func (self *Web) handleUploadSignedtxAct(w http.ResponseWriter, req *http.Request) {
	err := self.upload(w, req)

	rb := ResBack{}
	if err != nil {
		rb.Err = 1
		rb.ErrMsg = err.Error()
	}else{
		rb.ErrMsg = "上传成功"
	}

	b, _ := json.Marshal(rb)
	w.Write(b)
	return
}

// 处理/upload 逻辑
func (self *Web)upload(w http.ResponseWriter, r *http.Request) error{
	fmt.Println("method:", r.Method) //获取请求的方法
	var err error
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))
		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
		//r.ParseForm()
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println("1:", err)
			return err
		}
		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header)
		f, err := os.OpenFile("./data/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println("2", err)
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, file)
	}

	return err
}

// http handler
func (self *Web) handleSendSignedtx(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles("template/html/sendsignedtx.html")
	if err != nil {
		return
	}

	t.Execute(w, nil)
	return
}

// http handler
func (self *Web) handleSendSignedtxAct(w http.ResponseWriter, req *http.Request) {
	txfilename := req.FormValue("txsignedfilename")

	// signtx
	fmt.Println("--", txfilename)
	var argv []string
	argv = append(argv, "sendsignedtx")
	buildtxdir := self.onlineTool.GetDataDir()
	argv = append(argv, buildtxdir + "/" + txfilename)
	_, err := self.onlineTool.Execute(argv)

	rb := ResBack{}
	if err != nil {
		rb.Err = 1
		rb.ErrMsg = err.Error()
	}else{
		rb.ErrMsg = "发送完成"
	}

	b, _ := json.Marshal(rb)
	w.Write(b)
	return
}
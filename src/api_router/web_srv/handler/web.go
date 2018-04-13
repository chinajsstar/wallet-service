package handler

import (
	"../../data"
	"../../base/service"
	"encoding/json"
	"net/http"
	"html/template"
	"../../base/utils"
	"io/ioutil"
	"net/url"
	"strings"
	l4g "github.com/alecthomas/log4go"
	"fmt"
	"../../account_srv/user"
)

var G_henly_prikey []byte
var G_henly_pubkey []byte
var G_henly_licensekey string

var G_server_pubkey []byte
func loadRsaKeys() error {
	var err error
	G_henly_prikey, err = ioutil.ReadFile("/Users/henly.liu/workspace/private_henly.pem")
	if err != nil {
		return err
	}

	G_henly_pubkey, err = ioutil.ReadFile("/Users/henly.liu/workspace/public_henly.pem")
	if err != nil {
		return err
	}

	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	accountDir := appDir + "/account"
	G_server_pubkey, err = ioutil.ReadFile(accountDir + "/public.pem")
	if err != nil {
		return err
	}

	G_henly_licensekey = "524faf3a-b6a0-42ce-9c49-9c07b66aa835"

	return nil
}

////////////////////////////////////////////////////////////////////////////////
type Web struct{
	srvNode *service.ServiceNode
	nodes []*data.SrvRegisterData

	users map[string]*user.AckUserLogin
}

func NewWeb() *Web {
	w := &Web{}

	w.users = make(map[string]*user.AckUserLogin)
	return w
}

func (self *Web)Init(srvNode *service.ServiceNode) error {
	self.srvNode = srvNode

	if err := loadRsaKeys(); err != nil {
		return err
	}

	if err := self.startHttpServer(); err != nil{
		return err
	}

	return nil
}

func (self *Web)GetApiGroup()(map[string]service.NodeApi){
	nam := make(map[string]service.NodeApi)

	return nam
}

// start http server
func (self *Web) startHttpServer() error {
	// http
	l4g.Debug("Start http server on 8077")

	http.Handle("/listsrv", http.HandlerFunc(self.handleListSrv))
	http.Handle("/getapi", http.HandlerFunc(self.handleGetApi))
	http.Handle("/runapi", http.HandlerFunc(self.handleRunApi))

	http.Handle("/doc/", http.FileServer(http.Dir("template")))
	http.Handle("/css/", http.FileServer(http.Dir("template")))
	http.Handle("/js/", http.FileServer(http.Dir("template")))

	http.Handle("/index",http.HandlerFunc(self.handleIndex))
	http.Handle("/login",http.HandlerFunc(self.handleLogin))
	http.Handle("/dologin",http.HandlerFunc(self.LoginAction))
	http.Handle("/testapi", http.HandlerFunc(self.handleTestApi))
	http.Handle("/wallet/", http.HandlerFunc(self.handleWallet))
	http.Handle("/",http.HandlerFunc(self.handle404))

	go func() {
		l4g.Info("Http server routine running... ")
		err := http.ListenAndServe(":8077", nil)
		if err != nil {
			l4g.Crash("%s", err.Error())
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
func (self *Web) handleListSrv(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()
	// 获取cookie
	cookie, err := req.Cookie("name")
	if err != nil || cookie.Value == ""{
		http.Redirect(w, req, "/login", http.StatusFound)
	}

	//fmt.Println("path=", req.URL.Path)
	//fmt.Println("query=", req.URL.RawQuery)

	self.nodes = self.nodes[:0]
	var ureq data.UserRequestData
	var ures data.UserResponseData
	if err := self.srvNode.ListSrv(&ureq, &ures); err == nil && ures.Err == data.NoErr{
		json.Unmarshal([]byte(ures.Value.Message), &self.nodes)
	}
	l4g.Debug(ures)

	// listsrv
	t, err := template.ParseFiles("template/html/listsrv.html")
	if err != nil {
		return
	}

	t.Execute(w, self.nodes)
	return
}

// http handler
func (self *Web) handleGetApi(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()
	cookie, err := req.Cookie("name")
	if err != nil || cookie.Value == ""{
		http.Redirect(w, req, "/login", http.StatusFound)
	}

	//fmt.Println("path=", req.URL.Path)
	//fmt.Println("query=", req.URL.RawQuery)

	if len(self.nodes) == 0 {
		var req data.UserRequestData
		var res data.UserResponseData
		if err := self.srvNode.ListSrv(&req, &res); err == nil && res.Err == data.NoErr{
			json.Unmarshal([]byte(res.Value.Message), &self.nodes)
		}

		fmt.Println(res)
	}

	// getapi?srv
	vv := req.URL.Query()
	srvname := vv.Get("srv")
	//fmt.Println("srv=", srvname)

	srvNode := data.SrvRegisterData{}
	for _, v := range self.nodes {
		if v.Srv == srvname {
			srvNode = *v
			break
		}
	}

	t, err := template.ParseFiles("template/html/getapi.html")
	if err != nil {
		return
	}

	t.Execute(w, srvNode)
	return
}

// http handler
func (self *Web) handleRunApi(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()
	cookie, err := req.Cookie("name")
	if err != nil || cookie.Value == ""{
		//http.Redirect(w, req, "/login", http.StatusFound)
		//return
	}

	//fmt.Println("path=", req.URL.Path)
	//fmt.Println("query=", req.URL.RawQuery)

	example := ""
	bb, err := ioutil.ReadAll(req.Body)
	u, err := url.Parse(string(bb))
	kvs := strings.Split(u.Path, "&")
	for _, v := range kvs{
		kvs2 := strings.Split(v, "=")
		if len(kvs2) == 2 && kvs2[0] == "argv" {
			example = kvs2[1]
			break
		}
	}
	//fmt.Println("argv", example)

	if len(self.nodes) == 0 {
		var req data.UserRequestData
		var res data.UserResponseData
		if err := self.srvNode.ListSrv(&req, &res); err == nil && res.Err == data.NoErr{
			json.Unmarshal([]byte(res.Value.Message), &self.nodes)
		}

		fmt.Println(res)
	}

	// getapi?srv
	vv := req.URL.Query()

	var ureq data.UserRequestData
	ureq.Method.Srv = vv.Get("srv")
	ureq.Method.Version = vv.Get("ver")
	ureq.Method.Function = vv.Get("func")
	//fmt.Println("name=", ureq.Method.Srv)
	//fmt.Println("version=", ureq.Method.Version)
	//fmt.Println("function=", ureq.Method.Function)

	var ures data.UserResponseData
	if example != ""{
		var ud data.UserData
		ud.Message = example
		ud.LicenseKey = G_henly_licensekey
		//encryptUserData(example, G_henly_prikey, &ud)
		ureq.Argv = ud


		if err := self.srvNode.Dispatch(&ureq, &ures); err == nil && ures.Err == data.NoErr{
			//decryptUserData(&ures, G_henly_prikey)
		}
	}else{
		ures.Err = 1
		ures.ErrMsg = "没有提供测试实例"
	}

	t, err := template.ParseFiles("template/html/runapi.html")
	if err != nil {
		return
	}

	t.Execute(w, ures)
	return
}

type LoginUser struct {
	UserName string
	Demo1 string
	Demo2 string
}
// http handler
func (self *Web) handleIndex(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

	//fmt.Println("path=", req.URL.Path)
	//fmt.Println("query=", req.URL.RawQuery)
	cookie, err := req.Cookie("name")
	if err != nil || cookie.Value == ""{
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}

	// listsrv
	t, err := template.ParseFiles("template/html/index.html")
	if err != nil {
		return
	}

	t.Execute(w, &LoginUser{UserName:cookie.Value, Demo1:req.Host+"/listsrv", Demo2:"/testapi"})
	return
}

// http handler
func (self *Web) handleLogin(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

	//fmt.Println("path=", req.URL.Path)
	//fmt.Println("query=", req.URL.RawQuery)

	// listsrv
	t, err := template.ParseFiles("template/html/login.html")
	if err != nil {
		l4g.Error("%s", err.Error())
		return
	}

	t.Execute(w, nil)
	return
}

func (this *Web)LoginAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	ures := data.UserResponseData{}

	func(){
		message := ""
		bb, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ures.Err = data.ErrDataCorrupted
			ures.ErrMsg = data.ErrDataCorruptedText
			return
		}
		message = string(bb)
		fmt.Println("argv=", message)

		ul := user.ReqUserLogin{}
		json.Unmarshal(bb, &ul)

		if ul.UserName == "" || ul.Password == ""{
			ures.Err = 1
			ures.ErrMsg = "参数错误"
			return
		}

		var ureq data.UserRequestData

		ureq.Method.Srv = "account"
		ureq.Method.Version = "v1"
		ureq.Method.Function = "login"
		ureq.Argv.Message = G_henly_licensekey
		ureq.Argv.Message = message

		if err := this.srvNode.Dispatch(&ureq, &ures); err != nil || ures.Err != data.NoErr{
			//decryptUserData(&ures, G_henly_prikey)
			return
		}

		// 存入cookie,使用cookie存储
		//expiration := time.Unix(5, 0)
		cookie := http.Cookie{Name: "name", Value: ul.UserName, Path: "/"}
		http.SetCookie(w, &cookie)

	}()

	b, _ := json.Marshal(ures)

	fmt.Println(string(b))
	w.Write(b)

	return
}

// http handler
func (self *Web) handleTestApi(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

	//fmt.Println("path=", req.URL.Path)
	//fmt.Println("query=", req.URL.RawQuery)
	cookie, err := req.Cookie("name")
	if err != nil || cookie.Value == ""{
		//http.Redirect(w, req, "/login", http.StatusFound)
		//return
	}

	// listsrv
	t, err := template.ParseFiles("template/html/testapi.html")
	if err != nil {
		return
	}

	t.Execute(w, nil)
	return
}

// http handler
func (self *Web) handleWallet(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型

	//cookie, err := req.Cookie("name")
	//if err != nil || cookie.Value == ""{
	//	http.Redirect(w, req, "/login", http.StatusFound)
	//	return
	//}

	var ures data.UserResponseData
	func(){

		reqData := data.UserRequestData{}

		fmt.Println("path=", req.URL.Path)

		path := req.URL.Path
		path = strings.Replace(path, "wallet", "", -1)
		path = strings.TrimLeft(path, "/")
		path = strings.TrimRight(path, "/")
		// get method
		paths := strings.Split(path, "/")
		for i := 0; i < len(paths); i++ {
			if i == 0 {
				reqData.Method.Version = paths[i]
			}else if i == 1{
				reqData.Method.Srv = paths[i]
			} else{
				if reqData.Method.Function != "" {
					reqData.Method.Function += "."
				}
				reqData.Method.Function += paths[i]
			}
		}

		message := ""
		bb, err := ioutil.ReadAll(req.Body)
		if err != nil {
			ures.Err = data.ErrDataCorrupted
			ures.ErrMsg = data.ErrDataCorruptedText
			return
		}
		message = string(bb)
		fmt.Println("argv=", message)

		// make data
		reqData.Argv.Message = message
		reqData.Argv.LicenseKey = G_henly_licensekey
		self.srvNode.Dispatch(&reqData, &ures)
	}()

	b, _ := json.Marshal(ures)

	w.Write(b)
	return
}
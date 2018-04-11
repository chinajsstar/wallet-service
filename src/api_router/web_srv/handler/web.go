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

	G_henly_licensekey = "719101fe-93a0-44e5-909b-84a6e7fcb132"

	return nil
}

////////////////////////////////////////////////////////////////////////////////
type Web struct{
	srvNode *service.ServiceNode
	nodes []*data.SrvRegisterData
}

func NewWeb() *Web {
	w := &Web{}
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

	//http.Handle("/", http.HandlerFunc(self.handleWeb))
	http.Handle("/listsrv", http.HandlerFunc(self.handleListSrv))
	http.Handle("/getapi", http.HandlerFunc(self.handleGetApi))
	http.Handle("/runapi", http.HandlerFunc(self.handleRunApi))

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

// http handler
func (self *Web) handleListSrv(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

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
	curDir, _ := utils.GetCurrentDir()
	t, err := template.ParseFiles(curDir + "/listsrv.html")
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

	curDir, _ := utils.GetCurrentDir()
	t, err := template.ParseFiles(curDir + "/getapi.html")
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

	curDir, _ := utils.GetCurrentDir()
	t, err := template.ParseFiles(curDir + "/runapi.html")
	if err != nil {
		return
	}

	t.Execute(w, ures)
	return
}
package consulhelper

import (
	l4g "github.com/alecthomas/log4go"
	consulapi "github.com/hashicorp/consul/api"
)

type ConsulClient struct{
	*consulapi.Client

	Config 	*consulapi.Config
	Check 	*consulapi.AgentServiceCheck
	Registration *consulapi.AgentServiceRegistration
}

func NewConsulClient() *ConsulClient {
	cc := &ConsulClient{}
	cc.Config = new(consulapi.Config)
	cc.Check = new(consulapi.AgentServiceCheck)
	cc.Registration = new(consulapi.AgentServiceRegistration)
	return cc
}

func (cc *ConsulClient) Start() error {
	var err error
	cc.Client, err = consulapi.NewClient(cc.Config)
	if err != nil {
		l4g.Error("%s", err.Error())
		return err
	}

	//cc.Registration.ID = self.id
	//cc.Registration.Name = name
	//cc.Registration.Port = port
	//cc.Registration.Tags = []string{tags}
	//cc.Registration.Address = addr

	//check.Args = []string{"sh", "-c", "sleep 1 && exit 0"}
	//check.HTTP = fmt.Sprintf("http://%s:%d%s", registration.Address, registration.Port, "/check")
	//设置超时 5s。
	//check.Timeout = "5s"
	//设置间隔 5s。
	//check.Interval = "5s"
	//注册check服务。
	//registration.Check = &self.Check

	cc.Registration.Check = cc.Check

	return cc.Agent().ServiceRegister(cc.Registration)
}

func (cc *ConsulClient)Stop() error {
	return cc.Agent().ServiceDeregister(cc.Registration.ID)
}

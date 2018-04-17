package config

import (
	"io/ioutil"
	"encoding/json"
	l4g "github.com/alecthomas/log4go"
	"errors"
)

func LoadJsonNode(absPath string, name string, value interface{}) error {
	var err error
	var data []byte
	data, err = ioutil.ReadFile(absPath)
	if err != nil {
		l4g.Error("", err)
		return err
	}

	var raw interface{}
	err = json.Unmarshal(data, &raw)
	if err != nil {
		l4g.Error("", err)
		return err
	}

	if v, ok := raw.(map[string]interface{}); ok {
		v2 := v[name]
		if v2 == nil {
			return errors.New("no a node:"+name)
		}

		b, err := json.Marshal(v2)
		if err != nil {
			l4g.Error("", err)
			return err
		}
		return json.Unmarshal(b, value)
	}else{
		return errors.New("not a map json")
	}
}

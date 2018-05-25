package L4G

import (
	"github.com/alecthomas/log4go"
	"strings"
	"blockchain_server/conf"
)

var ( l4g log4go.Logger )

func init () {
	l4g = make(log4go.Logger)
	l4g_config_file := strings.Trim(config.L4gConfigFile(), " ")
	l4g.LoadConfiguration(l4g_config_file)
	l4g.Trace("blockchain_api 'l4g' instance init ok!!!!")
}

func Trace(arg0 interface{}, args  ...interface{}) {
	l4g.Trace(arg0, args...)
}

func Error(arg0 interface{}, args ...interface{}) {
	l4g.Error(arg0, args...)
}



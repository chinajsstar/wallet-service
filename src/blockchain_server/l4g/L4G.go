package L4G

import (
	"github.com/alecthomas/log4go"
	"fmt"
)

var ( l4gs map[string]log4go.Logger )

func init () {
	l4gs = make(map[string]log4go.Logger)
	l := BuildL4g("defualt", "blockchain_api")
	l.Trace("blockchain_api 'l4g' instance init ok!!!!")
}

func GetL4g(name string) log4go.Logger {
	return l4gs[name]
}

func BuildL4g(name string, filename string) log4go.Logger {
	l := l4gs[name]
	if l!=nil { return l }

	if name=="" { return nil }
	if filename=="" { filename = name }

	l = make(log4go.Logger)

	maxsize := 20 * 1024 * 1024
	maxline := 100000
	formate := "[%D %T] [%L] (%S) %M"

	flw := log4go.NewFileLogWriter(fmt.Sprintf("%s.log", filename), true)
	flw.SetFormat(formate)
	flw.SetRotateSize(maxsize)
	flw.SetRotateLines(maxline)
	flw.SetRotateDaily(true)

	l.AddFilter("file", log4go.DEBUG, flw)
	l.AddFilter("stdout", log4go.DEBUG, log4go.NewConsoleLogWriter())
	l4gs[name] = l

	return l
}

func Close(name string) {
	if name=="all" {
		for _, l := range l4gs {
			l.Close()
		}
	} else {
		l := l4gs[name]
		if l!=nil {
			l.Close()
		}
	}
	delete(l4gs, name)
}

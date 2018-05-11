package apidoc

import (
	"bastionpay_api/api"
)

type (
	ApiDoc struct{
		Name 	string
		Comment string
		Path 	string
		Input 	interface{}
		Output 	interface{}
	}

	ApiProxy interface {
		Help()(*ApiDoc)
		Run(req interface{}, ack interface{}) (*api.Error)
		Output(message string, out *string) (*api.Error)
	}
)
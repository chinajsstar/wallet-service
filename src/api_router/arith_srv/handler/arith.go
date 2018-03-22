package handler

import "strconv"

type Args struct {
	A int `json:"a"`
	B int `json:"b"`
}
type Arith int

func (arith *Arith)Add(args *Args, res *string)  error{
	*res = strconv.Itoa(args.A + args.B)
	return nil
}
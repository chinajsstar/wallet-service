package function

import "strconv"

type Args struct {
	A int `json:"a"`
	B int `json:"b"`
}
type MyFunc1 int

func (myFunc1 *MyFunc1)Add(args *Args, res *string)  error{
	*res = strconv.Itoa(args.A + args.B)
	return nil
}

type Args2 struct {
	A string
	B string
}
type MyFunc2 int
func (myFunc2 *MyFunc2)Sub(args *Args2, res *string)  error{
	*res = args.A + args.B
	return nil
}

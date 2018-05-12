package v1

type Args struct {
	A int `json:"a" comment:"加数1"`
	B int `json:"b" comment:"加数2"`
}

type AckArgs struct {
	C int `json:"c" comment:"和"`
}

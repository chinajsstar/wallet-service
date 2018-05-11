package account

// register a user
type Register struct {
	UserClass int 	`json:"user_class" comment:"user_class"`
	Level int 		`json:"level" comment:"level"`
}

// return a user key
type AckRegister struct {
	UserKey string	`json:"user_key" comment:"user_key"`
}
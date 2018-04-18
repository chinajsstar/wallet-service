package user

// 账号生成-输入
type ReqUserCreate struct{
	UserName 		string `json:"user_name"`
	UserClass 		int `json:"user_class"`
	Phone 			string `json:"phone"`
	Email 			string `json:"email"`
	Password 		string `json:"password"`
	Level 			int `json:level`
	GoogleAuth 		string `json:"google_auth"`
	PublicKey 		string `json:"public_key"`
	CallbackUrl 	string `json:"callback_url"`
	TimeZone 		int `json:"timezone"`
	Country 		string `json:"country"`
	Language 		string `json:"language"`
}
// 账号生成-输出
type AckUserCreate struct{
	UserKey 		string  `json:"user_key"`
	ServerPublicKey string  `json:"server_public_key"`
}

// 账号登入-输入
type ReqUserLogin struct{
	UserName 		string `json:"user_name"`
	Phone 			string `json:"phone"`
	Email 			string `json:"email"`
	Password 		string `json:"password"`
}
// 账号登入-输出
type AckUserLogin struct{
	UserKey 		string `json:"user_key"`
	UserName 		string `json:"user_name"`
	Phone 			string `json:"phone"`
	Email 			string `json:"email"`
}

// 修改密码-输入
type ReqUserUpdatePassword struct{
	UserName 		string `json:"user_name"`
	Phone 			string `json:"phone"`
	Email 			string `json:"email"`
	OldPassword		string `json:"old_password"`
	NewPassword		string `json:"new_password"`
}
// 修改密码-输出
type AckUserUpdatePassword struct{
	Status 			string `json:"status"`
}

// 用户基本资料
type UserProfile struct{
	Id 				int    `json:"id"`
	UserKey 		string `json:"user_key"`
	UserName 		string `json:"user_name"`
	UserClass 		int `json:"user_class"`
	Phone 			string `json:"phone"`
	Email 			string `json:"email"`
	// TODO: others info
}

// 用户列表-输入
type ReqUserList struct{
	Id 				int    `json:"id"`
}
// 用户列表-输出
type AckUserList struct{
	Users []UserProfile `json:"users"`
}

// 权限信息
type UserLevel struct{
	Level 			int    `json:"level"`
	IsFrozen 		rune   `json:"isfrozen"`
	PublicKey 		string `json:"public_key"`
}
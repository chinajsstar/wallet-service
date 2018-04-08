package user

// 总结构数据
type User struct{
    Id 				int
    UserName 		string
    Phone 			string
    Email 			string
    Salt 			string
    Password 		string
    GoogleAuth 		string
    LicenseKey      string
    PublicKey 		string
    Level 			int
    IsFrozen 		rune
    LastLoginTime 	int64
    LastLoginIp 	string
    LastLoginMac 	string
    CreateTime 		int64
    UpdateTime 		int64
    TimeZone 		string
    Country 		string
    Language 		string
}

// 账号生成-输入
type ReqUserCreate struct{
	UserName 		string `json:"user_name"`
	Phone 			string `json:"phone"`
	Email 			string `json:"email"`
	Password 		string `json:"password"`
	Level 			int `json:level`
	GoogleAuth 		string `json:"google_auth"`
	PublicKey 		string `json:"public_key"`
	TimeZone 		string `json:"timezone"`
	Country 		string `json:"country"`
	Language 		string `json:"language"`
}
// 账号生成-输出
type AckUserCreate struct{
	LicenseKey      string  `json:"license_key"`
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
	Id 				int    `json:"id"`
	UserName 		string `json:"user_name"`
	Phone 			string `json:"phone"`
	Email 			string `json:"email"`
	LicenseKey      string `json:"license_key"`
}

// 修改密码-输入
type ReqUserUpdatePassword struct{
	Id 				int    `json:"id"`
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
	UserName 		string `json:"user_name"`
	Phone 			string `json:"phone"`
	Email 			string `json:"email"`
	LicenseKey      string `json:"license_key"`
}

// 权限信息
type UserLevel struct{
	Level 			int    `json:"level"`
	IsFrozen 		rune   `json:"isfrozen"`
	PublicKey 		string `json:"public_key"`
}

// 用户列表-输入
type ReqUserList struct{
	Id 				int    `json:"id"`
}
// 用户列表-输出
type AckUserList struct{
	Users []UserProfile `json:"users"`
}
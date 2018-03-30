package user

// 总结构
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

// 账号生成
type UserCreate struct{
	UserName 		string `json:"user_name"`
	Phone 			string `json:"phone"`
	Email 			string `json:"email"`
	Password 		string `json:"password"`
	GoogleAuth 		string `json:"google_auth"`
	PublicKey 		string `json:"public_key"`
	TimeZone 		string `json:"timezone"`
	Country 		string `json:"country"`
	Language 		string `json:"language"`
}
type UserCreateAck struct{
	LicenseKey      string  `json:"license_key"`
	ServerPublicKey string  `json:"server_public_key"`
}

// 账号登入
type UserLogin struct{
	UserName 		string `json:"user_name"`
	Phone 			string `json:"phone"`
	Email 			string `json:"email"`
	Password 		string `json:"password"`
}
type UserLoginAck struct{
	Id 				int    `json:"id"`
	UserName 		string `json:"user_name"`
	Phone 			string `json:"phone"`
	Email 			string `json:"email"`
	LicenseKey      string `json:"license_key"`
}

// 修改密码
type UserUpdatePassword struct{
	Id 				int    `json:"id"`
	UserName 		string `json:"user_name"`
	Phone 			string `json:"phone"`
	Email 			string `json:"email"`
	OldPassword		string `json:"old_password"`
	NewPassword		string `json:"new_password"`
}
type UserUpdatePasswordAck struct{
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

// 用户列表
type UserList struct{
	Id 				int    `json:"id"`
}
type UserListAck struct{
	Users []UserProfile `json:"users"`
}
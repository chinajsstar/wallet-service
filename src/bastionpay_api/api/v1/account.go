package v1

// 账号注册-输入--register
type ReqUserRegister struct{
	UserClass 		int `json:"user_class" comment:"用户类型，0:普通用户 1:热钱包; 2:管理员"`
	Level 			int `json:"level" comment:"级别，0：用户，100：普通管理员"`
}
// 账号注册-输出
type AckUserRegister struct{
	UserKey 		string  `json:"user_key" comment:"用户唯一标示"`
}

// 修改公钥和回调地址-输入--update profile
type ReqUserUpdateProfile struct{
	UserKey			string `json:"user_key" comment:"用户唯一标示"`
	PublicKey		string `json:"public_key" comment:"用户公钥"`
	SourceIP		string `json:"source_ip" comment:"用户源IP"`
	CallbackUrl		string `json:"callback_url" comment:"用户回调"`
}
// 修改公钥和回调地址-输出
type AckUserUpdateProfile struct{
	Status 			string `json:"status" comment:"状态"`
}

// 获取公钥和回调地址-输入--read profile
type ReqUserReadProfile struct{
	UserKey			string `json:"user_key" comment:"用户唯一标示"`
}
// 获取公钥和回调地址-输出
type AckUserReadProfile struct{
	UserKey			string `json:"user_key" comment:"用户唯一标示"`
	PublicKey		string `json:"public_key" comment:"用户公钥"`
	SourceIP		string `json:"source_ip" comment:"用户源IP"`
	CallbackUrl		string `json:"callback_url" comment:"用户回调"`
}

// 用户基本资料
type UserBasic struct{
	Id 				int    	`json:"id" comment:"用户ID"`
	UserKey 		string 	`json:"user_key" comment:"用户唯一标示"`
	UserClass 		int 	`json:"user_class" comment:"用户类型"`
	Level 			int 	`json:"level" comment:"级别"`
	IsFrozen 		rune 	`json:"is_frozen" comment:"用户是否冻结"`
}

// 用户列表-输入--list
type ReqUserList struct{
	Id 				int    `json:"id" comment:"用户当前最小ID"`
}
// 用户列表-输出
type AckUserList []UserBasic
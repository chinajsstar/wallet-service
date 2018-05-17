package v1

// 账号注册-输入--register
type ReqUserRegister struct{
	UserClass 		int `json:"user_class" doc:"用户类型，0:普通用户 1:热钱包; 2:管理员"`
	Level 			int `json:"level" doc:"级别，0：用户，100：普通管理员，200：创世管理员"`
}
// 账号注册-输出
type AckUserRegister struct{
	UserKey 		string  `json:"user_key" doc:"用户唯一标示"`
}

// 修改公钥和回调地址-输入--update profile
type ReqUserUpdateProfile struct{
	PublicKey		string `json:"public_key" doc:"用户公钥"`
	SourceIP		string `json:"source_ip" doc:"用户源IP，用逗号(,)隔开"`
	CallbackUrl		string `json:"callback_url" doc:"用户回调"`
}
// 修改公钥和回调地址-输出
type AckUserUpdateProfile struct{
	Status 			string `json:"status" doc:"状态"`
}

// 获取公钥和回调地址-输入--read profile
type ReqUserReadProfile struct{

}
// 获取公钥和回调地址-输出
type AckUserReadProfile struct{
	UserKey			string `json:"user_key" doc:"用户唯一标示"`
	PublicKey		string `json:"public_key" doc:"用户公钥"`
	SourceIP		string `json:"source_ip" doc:"用户源IP"`
	CallbackUrl		string `json:"callback_url" doc:"用户回调"`
}

// 用户基本资料
type UserBasic struct{
	Id 				int    	`json:"id" doc:"用户ID"`
	UserKey 		string 	`json:"user_key" doc:"用户唯一标示"`
	UserClass 		int 	`json:"user_class" doc:"用户类型"`
	Level 			int 	`json:"level" doc:"级别"`
	IsFrozen 		rune 	`json:"is_frozen" doc:"用户是否冻结"`
}

// 用户列表-输入--list
type ReqUserList struct{
	TotalLines 		int 		`json:"total_lines" doc:"总数,0：表示首次查询"`
	PageIndex 		int 		`json:"page_index" doc:"页索引,1开始"`
	MaxDispLines 	int 		`json:"max_disp_lines" doc:"页最大数，100以下"`
}
// 用户列表-输出
type AckUserList struct{
	Data 			[]UserBasic `json:"data" doc:"用户列表"`

	TotalLines 		int 		`json:"total_lines" doc:"总数"`
	PageIndex 		int 		`json:"page_index" doc:"页索引"`
	MaxDispLines 	int 		`json:"max_disp_lines" doc:"页最大数"`
}
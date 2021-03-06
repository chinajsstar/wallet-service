package backend

// 账号注册-输入--register
type ReqUserRegister struct{
	UserClass 		int `json:"user_class" doc:"用户类型，0:普通用户 1:热钱包; 2:管理员"`
	Level 			int `json:"level" doc:"级别，0：用户，100：普通管理员，200：创世管理员"`
	IsFrozen        int	`json:"is_frozen" doc:"用户冻结状态，0: 正常；1：冻结状态，默认是0"`
	UserName        string `json:"user_name" doc:"用户名称"`
	UserMobile      string `json:"user_mobile" doc:"用户电话"`
	UserEmail       string `json:"user_email" doc:"用户邮箱"`
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
	ServerPublicKey	string `json:"server_public_key" doc:"BastionPay公钥"`
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
	ServerPublicKey	string `json:"server_public_key" doc:"BastionPay公钥"`
}

// 用户列表-输入--list
// 用户基本资料查询
type UserCondition struct{
	Id 				int    `json:"id,omitempty" doc:"用户ID"`
	UserName        string `json:"user_name,omitempty" doc:"用户名称"`
	UserMobile      string `json:"user_mobile,omitempty" doc:"用户电话"`
	UserEmail       string `json:"user_email,omitempty" doc:"用户邮箱"`
	UserKey 		string 	`json:"user_key,omitempty" doc:"用户唯一标示"`
	UserClass 		int 	`json:"user_class,omitempty" doc:"用户类型"`
	Level 			int 	`json:"level,omitempty" doc:"级别"`
	IsFrozen        int		`json:"is_frozen,omitempty" doc:"用户冻结状态，0: 正常；1：冻结状态，默认是0"`
}
type ReqUserList struct{
	TotalLines 		int 		`json:"total_lines" doc:"总数,0：表示首次查询"`
	PageIndex 		int 		`json:"page_index" doc:"页索引,1开始"`
	MaxDispLines 	int 		`json:"max_disp_lines" doc:"页最大数，100以下"`

	Condition       UserCondition   `json:"condition" doc:"条件查询"`
}
// 用户列表-输出
// 用户基本资料
type UserBasic struct{
	Id 				int    	`json:"id" doc:"用户ID"`
	UserName        string `json:"user_name" doc:"用户名称"`
	UserMobile      string `json:"user_mobile" doc:"用户电话"`
	UserEmail       string `json:"user_email" doc:"用户邮箱"`
	UserKey 		string 	`json:"user_key" doc:"用户唯一标示"`
	UserClass 		int 	`json:"user_class" doc:"用户类型"`
	Level 			int 	`json:"level" doc:"级别"`
	IsFrozen        int		`json:"is_frozen" doc:"用户冻结状态，0: 正常；1：冻结状态，默认是0"`
	CreateTime      int64	`json:"create_time" doc:"用户注册时间"`
	UpdateTime      int64	`json:"update_time" doc:"用户更新时间"`
}
type AckUserList struct{
	Data 			[]UserBasic `json:"data" doc:"用户列表"`

	TotalLines 		int 		`json:"total_lines" doc:"总数"`
	PageIndex 		int 		`json:"page_index" doc:"页索引"`
	MaxDispLines 	int 		`json:"max_disp_lines" doc:"页最大数"`
}

// 设置冻结开关
type ReqFrozenUser struct{
	UserKey 		string 	`json:"user_key" doc:"用户唯一标示"`
	IsFrozen        int		`json:"is_frozen" doc:"用户冻结状态，0: 正常；1：冻结状态，默认是0"`
}
// 返回冻结开关
type AckFrozenUser struct {
	UserKey 		string 	`json:"user_key" doc:"用户唯一标示"`
	IsFrozen        int		`json:"is_frozen" doc:"用户冻结状态，0: 正常；1：冻结状态，默认是0"`
}
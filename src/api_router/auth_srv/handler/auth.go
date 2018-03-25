package handler

type Args struct {
	LicenseKey string `json:"license_key"`
	Signature string `json:"signature"`
	Encrypted string `json:"encrypted"`
}
type Auth struct{

}

// 验证数据
func (auth *Auth)AuthData(args *Args, res *string)  error{
	return nil
}

// 打包数据
func (auth *Auth)EncryptData(args *Args, res *string)  error{
	return nil
}
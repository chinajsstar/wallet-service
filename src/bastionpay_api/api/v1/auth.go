package v1

// 验证解密数据
type ReqAuth struct {
	DataType 	int 	`json:"origin" doc:"数据类型，0：原始数据，1：admin.UserMessage"`
	ChipperData string 	`json:"chipper_data" doc:"加密数据"`
}

//type AckAuth struct {
//	OriginData string 	`json:"origin_data" doc:"原始数据"`
//}
//
//// 加密签名数据
//type ReqEnCrypt struct {
//	OriginData string 	`json:"origin_data" doc:"原始数据"`
//}
//
//type AckEnCrypt struct {
//	ChipperData string 	`json:"chipper_data" doc:"加密数据"`
//}
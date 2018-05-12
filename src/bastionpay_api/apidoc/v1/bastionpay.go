package v1

import (
	"bastionpay_api/apidoc"
	"bastionpay_api/api/v1"
)

var ApiDocSupportAssets = apidoc.ApiDoc{
	Name:"获取支持币种",
	Comment:"获取支持币种",
	Path:"/api/v1/bastionpay/support_assets",
	Input:nil,
	Output:[]string{"btc"},
}

var ApiDocAssetAttribute = apidoc.ApiDoc{
	Name:"获取币种属性",
	Comment:"获取币种属性",
	Path:"/api/v1/bastionpay/asset_attribute",
	Input:[]string{"btc"},
	Output:v1.AckAssetsAttributes{v1.AckAssetsAttribute{}},
}

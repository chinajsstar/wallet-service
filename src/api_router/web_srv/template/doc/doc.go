// 超级钱包后台使用说明

// ~开发说明
// TODO:省略一万字先

// ～模块说明
// 	gateway: 		对外开发的API接入网关，所有外部请求经过API网关进行验证，分发，签名;
//  account_srv: 	账号服务，包括创建内部/外部用户，登陆系统，查看系统用户等；
//  auth_srv: 		验证和签名服务，gateway在分发请求之前，调用此服务验证，处理完之后，调用此服务签名；
//  xxxx_srv: 		钱包核心服务，。。。
//  web_srv: 		简易的测试和开发服务
//  arith_srv:      一个用于测试的服务

// ～安装服务
// TODO:省略一万字先

// ~接口接入
//  对外请求服务只提供http服务，同时推送业务支持Websocket服务

// ~接入示例
//  >规则，使用http post
// 		path : http://ip:port/wallet/version/srv/function
//		body : json data
//  >通过gateway接入，需要用户加密签名数据：
//    	curl -d '{"license_key":"719101fe-93a0-44e5-909b-84a6e7fcb132", "signature":"", "message":"{\"a\":2, \"b\":1}"}' http://ip:port/wallet/v1/arith/add
//  >通过web_srv接入，只需要明文数据：
// 		curl -d '{"a":2, "b":1}' http://ip:port/wallet/v1/arith/add
// 		curl -d '{"user_name":"henly", "password":"123456"}' http://ip:port/wallet/v1/account/login
// 		curl -d '{"id":-1}' http://ip:port/wallet/v1/account/listusers

package doc

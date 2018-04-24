BastionPay钱包后台使用说明

~接入示例
    >规则，使用http post
    	path : http://ip:port/wallet/version/srv/function
    	body : json data

    >通过gateway接入，需要用户加密签名数据：
      	curl -d '{"user_key":"719101fe-93a0-44e5-909b-84a6e7fcb132", "signature":"", "message":"{\"a\":2, \"b\":1}"}' http://ip:port/wallet/v1/arith/add

    >通过web接入，只需要明文数据：
    	curl -d '{}' http://ip:port/wallet/v1/center/listsrv
    	curl -d '{"a":2, "b":1}' http://ip:port/wallet/v1/arith/add
    	curl -d '{"user_name":"test", "password":"123456"}' http://ip:port/wallet/v1/account/login
    	curl -d '{"user_name":"test", "email":"123456", "password":"123456", ...}' http://ip:port/wallet/v1/account/create
    	curl -d '{"id":-1}' http://ip:port/wallet/v1/account/listusers
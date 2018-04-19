BastionPay钱包后台开发说明

~开发说明
    BastionPay钱包后台采用流行的微服务架构；
    对外提供统一的http接口服务；
    与用户使用非对称加密进行数据加密验证传输。

～模块说明
    gateway:
        >对外提供统一的http接口网关；
        >所有外部请求经过接口网关进行过滤，通过验证服务(auth_srv)进行验证解密；
        >将合法请求分发至相应服务处理；
        >将服务处理完的应答通过验证服务(auth_srv)加密签名，返回给用户

    account_srv:
        >账号服务，包括创建内部/外部用户，登陆系统，查看系统用户等；

    auth_srv:
        >验证和签名服务，gateway在分发请求之前，调用此服务验证，处理完之后，调用此服务签名；

    cobank_srv:
        >钱包核心服务；

    push_srv:
        >推送服务，将数据推送至相应的用户回调地址；

    arith_srv:
        >一个用于测试的加法服务

    web_test:
        >简易的测试和开发服务

～安装服务
    TODO:省略一万字先

~接入示例
    >规则，使用http post
    	path : http://ip:port/wallet/version/srv/function
    	body : json data

    >通过gateway接入，需要用户加密签名数据：
      	curl -d '{"user_key":"719101fe-93a0-44e5-909b-84a6e7fcb132", "signature":"", "message":"{\"a\":2, \"b\":1}"}' http://ip:port/wallet/v1/arith/add

    >通过web_test接入，只需要明文数据：
    	curl -d '{}' http://ip:port/wallet/v1/center/listsrv
    	curl -d '{"a":2, "b":1}' http://ip:port/wallet/v1/arith/add
    	curl -d '{"user_name":"test", "password":"123456"}' http://ip:port/wallet/v1/account/login
    	curl -d '{"user_name":"test", "email":"123456", "password":"123456", ...}' http://ip:port/wallet/v1/account/create
    	curl -d '{"id":-1}' http://ip:port/wallet/v1/account/listusers

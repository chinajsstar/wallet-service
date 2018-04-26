模拟用户测试流程：

～注册和设置流程，设置一次即可：
1. 在首页注册一个用户，记录自己的user_key；
2. 登陆系统，在"文档"页面下载main.tar.gz测试工具
3. 解压测试工具：>tar -xzvf main.tar.gz
4. 在终端中运行测试工具：>./main
5. 创建用户的RSA密钥对，输入：>rsagen yourname
6. 会在当前目录下生成public_yourname.pem和private_yourname.pem
7. 退出，输入：>q
7. 登陆系统，在"开发设置"页面上传你的public_yourname.pem的内容，还要通知回调的url(http://你的ip:6666/walletcb)

～使用
8. 在终端下，启动./main
9. 先设置远程服务地址，输入：>set http://192.168.21.109:8080
10. 加载数据，输入：>load 6666 yourname user_key
11. 如果没有错误，输入：>arith.add 9 2
12. 如果返回正确，说明测试已通

～加密模式
13. >api srv function message

～测试明文模式
14. >testapi srv function message
15. >curl -d '{"user_key":"719101fe-93a0-44e5-909b-84a6e7fcb132", "signature":"", "message":"{\"a\":2, \"b\":1}"}' http://192.168.21.109:8080/wallettest/v1/arith/add
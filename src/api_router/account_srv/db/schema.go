package db

var (
	UsersSchema = `CREATE TABLE IF NOT EXISTS users (
id int(11) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
user_name varchar(255) NOT NULL COMMENT '用户名',
phone varchar(255) DEFAULT NULL COMMENT '电话号码',
email varchar(255) DEFAULT NULL COMMENT '邮件地址',
salt varchar(16) NOT NULL COMMENT '密码算法加盐',
password text NOT NULL COMMENT '密码',
google_auth varchar(255) NOT NULL COMMENT 'google验证',
license_key varchar(255) NOT NULL COMMENT '唯一标示',
public_key varchar(2048) DEFAULT NULL COMMENT '公钥',
level int(11) NOT NULL DEFAULT '0' COMMENT '级别',
is_frozen char(1) NOT NULL DEFAULT '0' COMMENT '是否被冻结',
last_login_time datetime DEFAULT NULL COMMENT '最后登陆时间',
last_login_ip varchar(255) DEFAULT NULL COMMENT '最后登陆IP',
last_login_mac varchar(255) DEFAULT NULL COMMENT '最后登陆MAC',
create_time datetime NOT NULL COMMENT '创建时间',
update_time datetime NOT NULL COMMENT '信息更新时间',
time_zone varchar(255) DEFAULT NULL COMMENT '时区',
country varchar(255) DEFAULT NULL COMMENT '国家',
language varchar(255) DEFAULT NULL COMMENT '语言',
unique (user_name),
unique (phone),
unique (email),
unique (license_key),
PRIMARY KEY (id, license_key)
)ENGINE=InnoDB AUTO_INCREMENT=10052 DEFAULT CHARSET=utf8;`
)
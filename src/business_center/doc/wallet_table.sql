-- ----------------------------
-- Table structure for `user`
-- ----------------------------
DROP TABLE IF EXISTS `user_property`;
CREATE TABLE `user_property` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(255) NOT NULL DEFAULT '',
  `user_name` varchar(255) NOT NULL DEFAULT '',
  `user_class` int(11) NOT NULL DEFAULT 0 COMMENT '0:普通用户 1:热钱包; 100:管理员',
  `phone` varchar(255) NOT NULL DEFAULT '',
  `email` varchar(255) NOT NULL DEFAULT '',
  `salt` varchar(16) NOT NULL COMMENT '密码算法加盐',
  `password` text NOT NULL COMMENT '密码',
  `google_auth` varchar(255) NOT NULL DEFAULT '',
  `license_key` varchar(255) NOT NULL DEFAULT '',
  `public_key` text NOT NULL COMMENT '用户公钥',
  `callback_url` varchar(255) NOT NULL DEFAULT '',
  `level` int(11) NOT NULL DEFAULT 0 COMMENT '管理员级别，0：用户，100：普通管理员，200：创世管理员',
  `last_login_time` datetime DEFAULT NULL,
  `last_login_ip` varchar(255) NOT NULL DEFAULT '',
  `last_login_mac` varchar(255) NOT NULL DEFAULT '',
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  `is_frozen` int(11) NOT NULL DEFAULT 0,
  `time_zone` int(11) NOT NULL DEFAULT 0,
  `country` varchar(255) NOT NULL DEFAULT '',
  `language` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`user_id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of `user`
-- ----------------------------
--INSERT user_property (user_id, user_name, user_class, create_date, update_date) VALUES ('795b587d-2ee7-4979-832d-5d0ea64205d5', '超级钱包', 1, now(), now());
--INSERT user_property (user_id, user_name, user_class, create_date, update_date) VALUES ('737205c4-af3c-426d-973d-165a0bf46c71', '商户1', 0, now(), now());
--INSERT user_property (user_id, user_name, user_class, create_date, update_date) VALUES ('f223c88b-102a-485d-a5da-f96bb55f0bdf', '商户2', 0, now(), now());
--INSERT user_property (user_id, user_name, user_class, create_date, update_date) VALUES ('3adda5a7-ab90-453d-a18a-dc608ac22553', '商户3', 0, now(), now());

-- ----------------------------
-- Table structure for `user_account`
-- ----------------------------
DROP TABLE IF EXISTS `user_account`;
CREATE TABLE `user_account` (
  `id` int(11) NOT NULL AUTO_INCREMENT, 
  `user_id` varchar(255) NOT NULL DEFAULT '',
  `asset_id` int(11) NOT NULL DEFAULT 0,
  `available_amount` double NOT NULL DEFAULT 0,
  `frozen_amount` double NOT NULL DEFAULT 0,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`user_id`,`asset_id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  UNIQUE KEY `user_id_asset` (`user_id`,`asset_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for `assets_property`
-- ----------------------------
DROP TABLE IF EXISTS `asset_property`;
CREATE TABLE `asset_property` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '资产编号',
  `classid` int(11) DEFAULT NULL COMMENT '资产所属类型',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '资产名称',
  `full_name` varchar(255) NOT NULL DEFAULT '' COMMENT '资产全称',
  `logo` varchar(255) NOT NULL DEFAULT '' COMMENT '图标',
  `deposit_min` double NOT NULL DEFAULT 0 COMMENT '最小充值数量',
  `withdrawal_rate` double NOT NULL DEFAULT 0 COMMENT '提币手续费率(按交易百分比)',
  `withdrawal_value` double NOT NULL DEFAULT 0 COMMENT '提币手续费(按固定金额的手续费)',
  `withdrawal_reserve_rate` double NOT NULL DEFAULT 0 COMMENT '提币准备金比率',
  `withdrawal_alert_rate` double NOT NULL DEFAULT 0 COMMENT '提币警报比率',
  `withdrawal_stategy` double NOT NULL DEFAULT 0 COMMENT '提币策略预警值', 
  `confirmation_num` int(11) NOT NULL DEFAULT 0 COMMENT '确认数',
  `decaimal` int(11) NOT NULL DEFAULT 0 COMMENT '小数精度',
  `gas_factor` double NOT NULL DEFAULT 0 COMMENT '矿工费乘数因子',
  `debt` double NOT NULL DEFAULT 0 COMMENT '资产缺口',
  `park_amount` double NOT NULL DEFAULT 0 COMMENT '归集数',
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of `assets_property`
-- ----------------------------
INSERT asset_property (name, full_name) VALUES ('btc', 'Bitcoin');
INSERT asset_property (name, full_name) VALUES ('eth', 'Ethereum');

-- ----------------------------
-- Table structure for `user_address`
-- ----------------------------
DROP TABLE IF EXISTS `free_address`;
CREATE TABLE `free_address` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `asset_id` int(11) NOT NULL,
  `address` varchar(255) NOT NULL DEFAULT '',
  `private_key` varchar(400) NOT NULL DEFAULT '',
  `create_time` datetime NOT NULL,
  PRIMARY KEY (`asset_id`, `address`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for `user_address`
-- ----------------------------
DROP TABLE IF EXISTS `user_address`;
CREATE TABLE `user_address` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(255) NOT NULL,
  `asset_id` int(11) NOT NULL,
  `address` varchar(255) NOT NULL DEFAULT '',
  `private_key` varchar(400) NOT NULL DEFAULT '',
  `available_amount` double(255,0) NOT NULL DEFAULT '0',
  `frozen_amount` double(255,0) NOT NULL DEFAULT '0',
  `enabled` int(11) NOT NULL DEFAULT 1,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`asset_id`,`address`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for `user_address`
-- ----------------------------
DROP TABLE IF EXISTS `pay_address`;
CREATE TABLE `pay_address` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `asset_id` int(11) NOT NULL,
  `address` varchar(255) NOT NULL DEFAULT '',
  `private_key` varchar(400) NOT NULL DEFAULT '',
  PRIMARY KEY (`asset_id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


-- ----------------------------
-- Table structure for `transaction_status`
-- ----------------------------
DROP TABLE IF EXISTS `transaction_status`;
CREATE TABLE `transaction_status` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `asset_id` varchar(255) NOT NULL DEFAULT '', 
  `hash` varchar(255) NOT NULL DEFAULT '',
  `blockin_height` bigint(20) DEFAULT NULL,
  `confirm_height` bigint(20) DEFAULT NULL, 
  `blockin_time` datetime NOT NULL,
  `confirm_time` datetime NOT NULL,
  `miners_fee` double DEFAULT 0 COMMENT '矿工费',
  `status` int(11) NOT NULL DEFAULT 0 COMMENT '0初始状态1待人工审核2人工审核通过3提现请求待发送4请求已发送5钱包成功6钱包拒绝7人工审核拒绝',
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  `remark` varchar(255) NOT NULL DEFAULT '',
  `reviewer_id` int(11) DEFAULT NULL,
  `inspect_result` varchar(255) DEFAULT '' COMMENT '可能有多条检测结果',
  PRIMARY KEY (`asset_id`, `hash`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for `transaction_flow`
-- ----------------------------
DROP TABLE IF EXISTS `transaction_flow`;
CREATE TABLE `transaction_flow` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `asset_id` varchar(255) NOT NULL,
  `hash` varchar(255) DEFAULT NULL,
  `address` varchar(255) NOT NULL,
  `trans_type` varchar(255) NOT NULL COMMENT 'to or from',
  `amount` double NOT NULL,
  `wallet_fee` double DEFAULT NULL COMMENT '手续费',
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


-- ----------------------------
-- Table structure for `withdrawal_order`
-- ----------------------------
DROP TABLE IF EXISTS `withdrawal_order`;
CREATE TABLE `withdrawl_order` (
  `id` int(11) NOT NULL AUTO_INCREMENT,  
  `order_id` varchar(255) NOT NULL DEFAULT '',
  `user_order_id` varchar(255) NOT NULL DEFAULT '',
  `user_id` varchar(255) NOT NULL DEFAULT '',
  `asset_id` varchar(255) NOT NULL DEFAULT '',
  `address` varchar(255) NOT NULL DEFAULT '',
  `amount` double NOT NULL DEFAULT 0,
  `wallet_fee` double NOT NULL DEFAULT 0,
  `create_time` datetime NOT NULL, 
  `hash` varchar(255) DEFAULT NULL DEFAULT '',
  PRIMARY KEY (`order_id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

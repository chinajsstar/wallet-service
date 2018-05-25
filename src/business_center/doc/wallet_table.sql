-- ----------------------------
-- Table structure for `user_property`
-- ----------------------------
DROP TABLE IF EXISTS `user_property`;
CREATE TABLE `user_property` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_key` varchar(255) NOT NULL DEFAULT '',
  `user_class` int(11) NOT NULL DEFAULT 0 COMMENT '0:普通用户 1:热钱包; 2:管理员',
  `public_key` text DEFAULT NULL COMMENT '公钥',
  `source_ip` varchar(255) NOT NULL DEFAULT '',
  `callback_url` varchar(255) NOT NULL DEFAULT '',
  `level` int(11) NOT NULL DEFAULT 0 COMMENT '级别，0：用户，100：普通管理员，200：创世管理员',
  `is_frozen` int(11) NOT NULL DEFAULT 0,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`user_key`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of `user`
-- ----------------------------
--INSERT user_property (user_key, user_name, user_class, create_time, update_time) VALUES ('795b587d-2ee7-4979-832d-5d0ea64205d5', '超级钱包', 1, now(), now());
--INSERT user_property (user_key, user_name, user_class, create_time, update_time) VALUES ('737205c4-af3c-426d-973d-165a0bf46c71', '商户1', 0, now(), now());
--INSERT user_property (user_key, user_name, user_class, create_time, update_time) VALUES ('f223c88b-102a-485d-a5da-f96bb55f0bdf', '商户2', 0, now(), now());
--INSERT user_property (user_key, user_name, user_class, create_time, update_time) VALUES ('3adda5a7-ab90-453d-a18a-dc608ac22553', '商户3', 0, now(), now());

-- ----------------------------
-- Table structure for `user_account`
-- ----------------------------
DROP TABLE IF EXISTS `user_account`;
CREATE TABLE `user_account` (
  `id` int(11) NOT NULL AUTO_INCREMENT, 
  `user_key` varchar(255) NOT NULL DEFAULT '',
  `user_class` int(11) NOT NULL DEFAULT 0,
  `asset_name` varchar(255) NOT NULL DEFAULT '',
  `available_amount` decimal(32, 12) NOT NULL DEFAULT 0,
  `frozen_amount` decimal(32, 12) NOT NULL DEFAULT 0,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`user_key`,`asset_name`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  UNIQUE KEY `user_key_asset` (`user_key`,`asset_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for `assets_property`
-- ----------------------------
DROP TABLE IF EXISTS `asset_property`;
CREATE TABLE `asset_property` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '资产编号',
  `asset_name` varchar(255) NOT NULL DEFAULT '' COMMENT '资产名称',
  `full_name` varchar(255) NOT NULL DEFAULT '' COMMENT '资产全称',
  `is_token` int(11) NOT NULL DEFAULT '0' COMMENT '资产所属类型',
  `parent_name` varchar(255) NOT NULL DEFAULT '',
  `logo` varchar(255) NOT NULL DEFAULT '' COMMENT '图标',
  `deposit_min` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '最小充值数量',
  `withdrawal_rate` decimal(32, 12) NOT NULL DEFAULT '0' COMMENT '提币手续费率(按交易百分比)',
  `withdrawal_value` decimal(32, 12) NOT NULL DEFAULT '0' COMMENT '提币手续费(按固定金额的手续费)',
  `withdrawal_reserve_rate` decimal(32, 12) NOT NULL DEFAULT '0' COMMENT '提币准备金比率',
  `withdrawal_alert_rate` decimal(32, 12) NOT NULL DEFAULT '0' COMMENT '提币警报比率',
  `withdrawal_stategy` decimal(32, 12) NOT NULL DEFAULT '0' COMMENT '提币策略预警值',
  `confirmation_num` int(11) NOT NULL DEFAULT '0' COMMENT '确认数',
  `decimals` int(11) NOT NULL DEFAULT '0' COMMENT '小数精度',
  `gas_factor` decimal(32, 12) NOT NULL DEFAULT '0' COMMENT '矿工费乘数因子',
  `debt` decimal(32, 12) NOT NULL DEFAULT '0' COMMENT '资产缺口',
  `park_amount` decimal(32, 12) NOT NULL DEFAULT '0' COMMENT '归集数',
  PRIMARY KEY (`asset_name`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of `assets_property`
-- ----------------------------
INSERT asset_property (asset_name, full_name) VALUES ('btc', 'Bitcoin');
INSERT asset_property (asset_name, full_name) VALUES ('eth', 'Ethereum');
INSERT asset_property (asset_name, full_name, is_token, parent_name) VALUES ('ZToken', 'ZToken', 1, 'eth')

-- ----------------------------
-- Table structure for `user_address`
-- ----------------------------
DROP TABLE IF EXISTS `free_address`;
CREATE TABLE `free_address` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `asset_name` varchar(255) NOT NULL DEFAULT '',
  `address` varchar(255) NOT NULL DEFAULT '',
  `private_key` varchar(400) NOT NULL DEFAULT '',
  `create_time` datetime NOT NULL,
  PRIMARY KEY (`asset_name`, `address`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for `user_address`
-- ----------------------------
DROP TABLE IF EXISTS `user_address`;
CREATE TABLE `user_address` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_key` varchar(255) NOT NULL,
  `user_class` int(11) NOT NULL DEFAULT 0,
  `available_amount` decimal(32, 12) NOT NULL DEFAULT '0',
  `frozen_amount` decimal(32, 12) NOT NULL DEFAULT '0', 
  `asset_name` varchar(255) NOT NULL DEFAULT '',
  `address` varchar(255) NOT NULL DEFAULT '',
  `private_key` varchar(400) NOT NULL DEFAULT '',
  `enabled` int(11) NOT NULL DEFAULT 1,
  `create_time` datetime NOT NULL,
  `allocation_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`asset_name`,`address`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for `user_address`
-- ----------------------------
DROP TABLE IF EXISTS `pay_address`;
CREATE TABLE `pay_address` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `asset_name` varchar(255) NOT NULL DEFAULT '',
  `address` varchar(255) NOT NULL DEFAULT '',
  `private_key` varchar(400) NOT NULL DEFAULT '',
  PRIMARY KEY (`asset_name`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


-- ----------------------------
-- Table structure for `transaction_blockin`
-- ----------------------------
DROP TABLE IF EXISTS `transaction_blockin`;
CREATE TABLE `transaction_blockin` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `asset_name` varchar(255) NOT NULL DEFAULT '',
  `hash` varchar(255) NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '0入块,1已确认,>=2错误状态',
  `miner_fee` decimal(32, 12) NOT NULL DEFAULT '0',
  `blockin_height` bigint(20) NOT NULL,
  `blockin_time` datetime NOT NULL,
  `confirm_height` bigint(20) NOT NULL,
  `confirm_time` datetime NOT NULL,
  `order_id` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`asset_name`,`hash`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


-- ----------------------------
-- Table structure for `transaction_detail`
-- ----------------------------
DROP TABLE IF EXISTS `transaction_detail`;
CREATE TABLE `transaction_detail` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `asset_name` varchar(255) NOT NULL DEFAULT '',
  `address` varchar(255) NOT NULL,
  `trans_type` varchar(255) NOT NULL COMMENT '支出（from）, 收入(to), 矿工费(miner_fee), 找零(change)',
  `amount` decimal(32, 12) NOT NULL DEFAULT 0,
  `hash` varchar(255) DEFAULT NULL,
  `detail_id` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`detail_id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


-- ----------------------------
-- Table structure for `transaction_status`
-- ----------------------------
DROP TABLE IF EXISTS `transaction_status`;
CREATE TABLE `transaction_status` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `asset_name` varchar(255) NOT NULL DEFAULT '',
  `hash` varchar(255) NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT 0 COMMENT '0入块,1已确认,>=2错误状态',
  `confirm_height` bigint(20) DEFAULT NULL,
  `confirm_time` datetime DEFAULT NULL,
  `update_time` datetime NOT NULL,
  `order_id` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`asset_name`,`hash`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8;


-- ----------------------------
-- Table structure for `withdrawal_order`
-- ----------------------------
DROP TABLE IF EXISTS `withdrawal_order`;
CREATE TABLE `withdrawal_order` (
  `id` int(11) NOT NULL AUTO_INCREMENT,  
  `order_id` varchar(255) NOT NULL DEFAULT '',
  `user_order_id` varchar(255) NOT NULL DEFAULT '',
  `user_key` varchar(255) NOT NULL DEFAULT '',
  `asset_name` varchar(255) NOT NULL DEFAULT '',
  `address` varchar(255) NOT NULL DEFAULT '',
  `amount` decimal(32, 12) NOT NULL DEFAULT 0,
  `pay_fee` decimal(32, 12) NOT NULL DEFAULT 0,
  `miner_fee` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '矿工费',
  `create_time` datetime NOT NULL, 
  `hash` varchar(255) DEFAULT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT 0 COMMENT '0入块,1已确认,>=2错误状态',
  PRIMARY KEY (`order_id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for `transaction_bill`
-- ----------------------------
DROP TABLE IF EXISTS `transaction_bill`;
CREATE TABLE `transaction_bill` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_key` varchar(255) NOT NULL COMMENT '商户Key',
  `trans_type` int(11) NOT NULL COMMENT '0:充值, 1:提币',
  `status` int(11) NOT NULL COMMENT '0:入块, 1:成功, >1:失败',
  `blockin_height` bigint(20) NOT NULL,
  `asset_name` varchar(255) NOT NULL DEFAULT '',
  `address` varchar(255) NOT NULL DEFAULT '' COMMENT '地址',
  `amount` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '金额',
  `pay_fee` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '手续费',
  `miner_fee` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '矿工费',
  `balance` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '余额',
  `hash` varchar(255) NOT NULL DEFAULT '',
  `order_id` varchar(255) NOT NULL DEFAULT '',
  `time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for `transaction_notice`
-- ----------------------------
DROP TABLE IF EXISTS `transaction_notice`;
CREATE TABLE `transaction_notice` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_key` varchar(255) NOT NULL COMMENT '商户Key',
  `msg_id` bigint(20) NOT NULL DEFAULT 0 COMMENT '消息序号',
  `trans_type` int(11) NOT NULL COMMENT '0:充值, 1:提币',
  `status` int(11) NOT NULL COMMENT '0:入块, 1:成功, >1:失败',
  `blockin_height` bigint(20) NOT NULL,
  `asset_name` varchar(255) NOT NULL DEFAULT '',
  `address` varchar(255) NOT NULL DEFAULT '' COMMENT '地址',
  `amount` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '金额',
  `pay_fee` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '手续费',
  `miner_fee` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '矿工费',  
  `balance` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '余额',
  `hash` varchar(255) NOT NULL DEFAULT '',
  `order_id` varchar(255) NOT NULL DEFAULT '',
  `time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for `user_order`
-- ----------------------------
DROP TABLE IF EXISTS `user_order`;
CREATE TABLE `user_order` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_key` varchar(255) NOT NULL COMMENT '商户Key',
  `user_order_id` varchar(255) NOT NULL DEFAULT '',
  `order_id` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`user_key`,`user_order_id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


-- ----------------------------
-- Table structure for `transaction_bill_daily`
-- ----------------------------
DROP TABLE IF EXISTS `transaction_bill_daily`;
CREATE TABLE `transaction_bill_daily` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `period` int(11) NOT NULL DEFAULT 0 COMMENT '周期标识',
  `user_key` varchar(255) NOT NULL DEFAULT '商户Key',
  `asset_name` varchar(255) NOT NULL DEFAULT '' COMMENT '币种',
  `sum_dp_amount` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '总充币数量',
  `sum_wd_amount` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '总提币数据',
  `sum_pay_fee` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '总手续费',
  `sum_miner_fee` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '总矿工费',
  `pre_balance` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '上期余额',
  `last_balance` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '当期余额',
  `pre_time` datetime NOT NULL COMMENT '当期最早时间',
  `last_time` datetime NOT NULL COMMENT '最后更新时间',              
  PRIMARY KEY (`period`,`user_key`,`asset_name`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


-- ----------------------------
-- Table structure for `profit_bill`
-- ----------------------------
DROP TABLE IF EXISTS `profit_bill`;
CREATE TABLE `profit_bill` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `profit_user_key` varchar(255) NOT NULL DEFAULT '' COMMENT '利润归属对象',
  `user_key` varchar(255) NOT NULL DEFAULT '' COMMENT '商户Key',
  `trans_type` int(11) NOT NULL COMMENT '0:充值, 1:提币',
  `asset_name` varchar(255) NOT NULL DEFAULT '' COMMENT '币种', 
  `order_id` varchar(255) NOT NULL DEFAULT '' COMMENT '交易订单号',
  `hash` varchar(255) NOT NULL DEFAULT '',
  `amount` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '金额',
  `pay_fee` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '手续费',
  `miner_fee` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '矿工',
  `profit` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '利润',
  `time` bigint(20) NOT NULL DEFAULT 0 COMMENT '时间',             
  PRIMARY KEY (`profit_user_key`, `asset_name`, `order_id`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


-- ----------------------------
-- Table structure for `profit_bill_daily`
-- ----------------------------
DROP TABLE IF EXISTS `profit_bill_daily`;
CREATE TABLE `profit_bill_daily` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `period` int(11) NOT NULL DEFAULT 0 COMMENT '周期标识',
  `profit_user_key` varchar(255) NOT NULL DEFAULT '' COMMENT '利润归属对象',
  `asset_name` varchar(255) NOT NULL DEFAULT '' COMMENT '币种', 
  `sum_profit` decimal(32, 12) NOT NULL DEFAULT 0 COMMENT '利润',
  `time` bigint(20) NOT NULL DEFAULT 0 COMMENT '时间',             
  PRIMARY KEY (`profit_user_key`, `asset_name`),
  UNIQUE KEY `id_UNIQUE` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


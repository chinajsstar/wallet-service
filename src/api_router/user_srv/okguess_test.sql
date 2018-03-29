/*
Navicat MySQL Data Transfer

Source Server         : 119.37.198.140
Source Server Version : 50173
Source Host           : 119.37.198.140:3306
Source Database       : okguess_test

Target Server Type    : MYSQL
Target Server Version : 50173
File Encoding         : 65001

Date: 2018-03-23 14:21:29
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for `assets_property`
-- ----------------------------
DROP TABLE IF EXISTS `assets_property`;
CREATE TABLE `assets_property` (
  `asset_id` int(11) NOT NULL AUTO_INCREMENT,
  `asset_classid` int(11),
  `asset_name` varchar(255) NOT NULL COMMENT '资产名称',
  `asset_shortname` varchar(255) NOT NULL COMMENT '资产名称',
  `asset_logo` varchar(255) COMMENT '图标'
  `withdrawal_rate` double NOT NULL COMMENT '提币手续费率',
  `withdrawal_value` double NOT NULL COMMENT '提币手续费',
  `withdrawal_reserve_rate` NOT NULL COMMENT '提币准备金比率'
  `withdrawal_alert_rate` NOT NULL COMMENT '提币警报比率'
  `min_deposit` double NOT NULL COMMENT '最小充值数量',
  `confirmation_num` int(11) NOT NULL COMMENT '确认数',
  `stategy_withdrawal` double NOT NULL COMMENT '提币策略预警值',
  `decaimal` int(11) NOT NULL COMMENT '小数精度',
  `gas_factor` double NOT NULL COMMENT '矿工费乘数因子'
  `debt` double NOT NULL COMMENT '资产缺口'
  `park_amount` COMMENT '归集数'
  PRIMARY KEY (`asset_id`),
  KEY `asset` (`asset`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of assets_property
-- ----------------------------
INSERT INTO `assets_property` VALUES ('1', 'uBTC', '0', '0.001', '0.0001', '3', '10', '1', '2', '3', '1', '0', '0', '8', '0.001');

-- ----------------------------
-- Table structure for `back_oper_log`
-- ----------------------------
DROP TABLE IF EXISTS `back_oper_log`;
CREATE TABLE `back_oper_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `operAdminId` int(11) DEFAULT NULL,
  `operAdminNm` varchar(30) DEFAULT NULL,
  `operIp` varchar(15) DEFAULT NULL,
  `operedAdminId` int(11) DEFAULT NULL,
  `operedAdminNm` varchar(30) DEFAULT NULL,
  `operTime` datetime DEFAULT NULL,
  `operType` varchar(50) DEFAULT NULL,
  `operDescri` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=851 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of back_oper_log
-- ----------------------------
INSERT INTO `back_oper_log` VALUES ('684', '1', 'bocai_adm', '116.247.101.86', null, null, '2016-06-27 15:24:32', '刪除日志', '日志ID：680');
INSERT INTO `back_oper_log` VALUES ('691', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-07-25 10:31:46', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('689', '1', 'bocai_adm', '116.247.101.86', null, null, '2016-06-27 15:26:08', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('690', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-06-27 15:40:17', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('685', '1', 'bocai_adm', '116.247.101.86', null, null, '2016-06-27 15:24:32', '刪除日志', '日志ID：683');
INSERT INTO `back_oper_log` VALUES ('686', '1', 'bocai_adm', '116.247.101.86', null, null, '2016-06-27 15:24:32', '刪除日志', '日志ID：682');
INSERT INTO `back_oper_log` VALUES ('687', '1', 'bocai_adm', '116.247.101.86', null, null, '2016-06-27 15:24:32', '刪除日志', '日志ID：681');
INSERT INTO `back_oper_log` VALUES ('688', '1', 'bocai_adm', '116.247.101.86', null, null, '2016-06-27 15:24:32', '刪除日志', '日志ID：680');
INSERT INTO `back_oper_log` VALUES ('692', '1', 'bocai_adm', '116.246.22.146', null, null, '2016-07-25 10:37:17', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('693', '1', 'bocai_adm', '116.247.102.2', null, null, '2016-07-25 11:03:33', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('694', '1', 'bocai_adm', '116.246.22.146', null, null, '2016-07-25 11:23:33', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('695', '1', 'bocai_adm', '116.247.101.86', null, null, '2016-07-25 12:05:16', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('696', '1', 'bocai_adm', '116.246.22.146', null, null, '2016-07-25 13:03:29', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('697', '1', 'bocai_adm', '116.247.102.2', null, null, '2016-07-25 13:40:00', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('698', '1', 'bocai_adm', '116.246.22.146', null, null, '2016-07-25 14:31:52', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('699', '1', 'bocai_adm', '116.247.102.2', null, null, '2016-07-25 15:10:21', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('700', '1', 'bocai_adm', '116.247.102.2', null, null, '2016-07-25 15:18:47', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('701', '1', 'bocai_adm', '116.247.102.2', null, null, '2016-07-25 15:52:25', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('702', '1', 'bocai_adm', '116.246.22.146', null, null, '2016-07-25 15:56:02', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('703', '1', 'bocai_adm', '116.247.102.2', null, null, '2016-07-26 10:45:49', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('704', '1', 'bocai_adm', '116.247.102.2', null, null, '2016-07-26 13:24:56', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('705', '1', 'bocai_adm', '116.247.102.2', null, null, '2016-07-26 13:24:57', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('706', '1', 'bocai_adm', '116.247.102.2', null, null, '2016-07-26 17:16:01', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('707', '1', 'bocai_adm', '116.247.102.2', null, null, '2016-07-27 10:06:06', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('708', '1', 'bocai_adm', '116.247.102.2', null, null, '2016-07-27 10:40:33', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('709', '1', 'bocai_adm', '116.247.102.2', null, null, '2016-07-27 13:05:25', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('710', '1', 'bocai_adm', '116.247.101.86', null, null, '2016-07-28 09:40:08', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('711', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-08-07 16:47:04', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('712', '1', 'bocai_adm', '180.169.10.222', '2', 'lukuan', '2016-08-07 16:48:17', '添加用户', '添加用户： lukuan');
INSERT INTO `back_oper_log` VALUES ('713', '1', 'bocai_adm', '116.247.101.86', null, null, '2016-08-10 10:39:21', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('714', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-08-10 10:57:04', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('715', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-08-10 13:56:53', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('716', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-08-17 11:02:49', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('717', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-08-31 21:25:37', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('718', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-11-11 10:51:55', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('719', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-11-11 12:47:07', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('720', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-11-11 15:17:02', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('721', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-11-11 16:09:30', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('722', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-11-14 13:27:39', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('723', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-11-15 14:03:46', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('724', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-11-15 15:31:22', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('725', '1', 'bocai_adm', '180.169.10.222', null, null, '2016-11-16 13:51:57', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('726', '2', 'lukuan', '180.169.10.222', null, null, '2016-12-30 11:10:54', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('727', '2', 'lukuan', '180.169.10.222', null, null, '2016-12-30 13:38:12', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('728', '2', 'lukuan', '180.168.223.138', null, null, '2017-01-03 15:00:27', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('729', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-04 13:53:16', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('730', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-04 15:17:26', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('731', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-04 17:29:06', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('732', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-04 18:42:57', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('733', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-06 11:03:08', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('734', '2', 'lukuan', '180.168.223.138', null, null, '2017-01-06 11:12:26', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('735', '2', 'lukuan', '180.168.223.138', null, null, '2017-01-06 11:12:59', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('736', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-06 11:21:16', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('737', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-06 12:09:06', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('738', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-06 14:10:14', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('739', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-09 12:42:21', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('740', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-09 12:49:47', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('741', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-09 17:13:42', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('742', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-13 12:14:22', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('743', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-16 15:12:01', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('744', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-18 12:50:07', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('745', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-18 14:29:26', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('746', '2', 'lukuan', '180.168.223.138', null, null, '2017-01-18 15:17:11', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('747', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-18 17:14:43', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('748', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-18 18:30:59', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('749', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-20 15:11:10', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('750', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-20 16:01:11', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('751', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-20 17:01:33', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('752', '2', 'lukuan', '180.169.10.222', null, null, '2017-01-23 11:02:45', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('753', '2', 'lukuan', '180.164.249.130', null, null, '2017-02-04 16:52:17', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('754', '2', 'lukuan', '180.164.249.130', null, null, '2017-02-05 09:51:53', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('755', '2', 'lukuan', '180.164.249.130', null, null, '2017-02-05 13:35:21', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('756', '2', 'lukuan', '180.169.10.222', null, null, '2017-02-06 09:36:16', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('757', '2', 'lukuan', '180.169.10.222', null, null, '2017-02-06 12:40:42', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('758', '2', 'lukuan', '180.169.10.222', null, null, '2017-02-08 15:42:44', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('759', '2', 'lukuan', '180.164.249.130', null, null, '2017-02-11 13:10:32', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('760', '2', 'lukuan', '180.169.10.222', null, null, '2017-02-13 10:54:54', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('761', '2', 'lukuan', '180.169.10.222', null, null, '2017-02-13 13:26:58', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('762', '2', 'lukuan', '180.169.10.222', null, null, '2017-02-13 17:10:29', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('763', '2', 'lukuan', '180.169.10.222', null, null, '2017-03-09 10:16:44', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('764', '2', 'lukuan', '180.164.145.251', null, null, '2017-03-11 18:41:21', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('765', '2', 'lukuan', '101.81.162.202', null, null, '2017-05-13 14:14:47', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('766', '2', 'lukuan', '218.82.65.66', null, null, '2017-05-13 17:38:06', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('767', '2', 'lukuan', '218.82.65.66', null, null, '2017-05-13 17:38:06', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('768', '2', 'lukuan', '101.81.162.202', null, null, '2017-05-13 17:55:40', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('769', '2', 'lukuan', '180.173.44.0', null, null, '2017-06-26 14:44:23', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('770', '1', 'bocai_adm', '180.169.140.222', null, null, '2017-08-02 13:42:41', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('771', '1', 'bocai_adm', '180.169.140.222', null, null, '2017-08-02 14:15:47', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('772', '1', 'bocai_adm', '10.15.54.189', null, null, '2017-08-02 14:21:35', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('773', '1', 'bocai_adm', '180.169.140.222', null, null, '2017-08-02 14:54:24', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('774', '2', 'lukuan', '223.104.210.157', null, null, '2017-08-26 16:13:13', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('775', '2', 'lukuan', '223.104.212.114', null, null, '2017-08-27 11:44:37', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('776', '2', 'lukuan', '223.104.210.162', null, null, '2017-08-31 11:51:57', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('777', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-10 23:53:34', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('778', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:02:54', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('779', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:03:02', '删除菜单', '删除菜单，菜单id：27菜单名称：积分管理');
INSERT INTO `back_oper_log` VALUES ('780', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:04:07', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('781', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:08:16', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('782', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:09:01', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('783', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:14:25', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('784', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:23:47', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('785', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:26:20', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('786', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:32:13', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('787', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:40:36', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('788', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:42:56', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('789', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:47:00', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('790', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:52:12', '添加菜单', '添加菜单：充值订单');
INSERT INTO `back_oper_log` VALUES ('791', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:52:22', '编辑角色', '更新角色， 角色id:1');
INSERT INTO `back_oper_log` VALUES ('792', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 21:52:33', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('793', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 22:30:12', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('794', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 22:38:23', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('795', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 22:41:50', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('796', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-14 23:01:27', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('797', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-15 19:11:14', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('798', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-15 19:12:50', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('799', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-15 19:14:27', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('800', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-15 19:25:59', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('801', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-15 19:31:52', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('802', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-15 19:34:40', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('803', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-15 19:35:40', '添加菜单', '添加菜单：提现订单');
INSERT INTO `back_oper_log` VALUES ('804', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-15 22:03:22', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('805', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-15 22:03:32', '编辑角色', '更新角色， 角色id:1');
INSERT INTO `back_oper_log` VALUES ('806', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-15 22:03:38', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('807', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-15 22:04:49', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('808', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-15 22:12:46', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('809', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-15 22:20:10', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('810', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-16 18:32:07', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('811', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-17 11:33:01', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('812', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-17 11:42:12', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('813', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-17 17:09:02', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('814', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-17 17:09:02', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('815', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-17 17:26:06', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('816', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 00:37:54', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('817', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 00:42:47', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('818', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 00:44:11', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('819', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 00:46:23', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('820', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 01:20:04', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('821', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 01:28:44', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('822', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 01:43:15', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('823', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 01:45:02', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('824', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 01:46:59', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('825', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 01:49:17', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('826', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 01:59:34', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('827', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 02:01:42', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('828', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 02:03:32', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('829', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 02:03:35', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('830', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 02:04:28', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('831', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 02:08:48', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('832', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 02:09:41', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('833', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 02:13:52', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('834', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 02:15:33', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('835', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 02:20:02', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('836', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 20:44:40', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('837', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 20:49:18', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('838', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 21:16:39', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('839', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-18 21:18:00', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('840', '2', 'lukuan', '61.152.132.170', null, null, '2018-03-19 07:12:56', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('841', '2', 'lukuan', '167.99.60.108', null, null, '2018-03-19 08:56:45', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('842', '2', 'lukuan', '167.99.63.46', null, null, '2018-03-19 08:57:57', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('843', '2', 'lukuan', '117.136.8.235', null, null, '2018-03-19 17:35:30', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('844', '2', 'lukuan', '61.152.132.170', null, null, '2018-03-19 18:30:53', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('845', '2', 'lukuan', '61.152.132.170', null, null, '2018-03-19 19:26:28', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('846', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-19 20:33:38', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('847', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-19 20:33:45', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('848', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-19 20:37:29', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('849', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-19 20:39:12', '后台登陆', '后台登录系统');
INSERT INTO `back_oper_log` VALUES ('850', '1', 'bocai_adm', '0:0:0:0:0:0:0:1', null, null, '2018-03-19 20:39:14', '后台登陆', '后台登录系统');

-- ----------------------------
-- Table structure for `back_role`
-- ----------------------------
DROP TABLE IF EXISTS `back_role`;
CREATE TABLE `back_role` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `NAME` varchar(30) NOT NULL,
  `treeIds` varchar(100) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of back_role
-- ----------------------------
INSERT INTO `back_role` VALUES ('1', '系统管理员', '2, 3, 4, 5, 6, 26, 28, 29');

-- ----------------------------
-- Table structure for `back_tree`
-- ----------------------------
DROP TABLE IF EXISTS `back_tree`;
CREATE TABLE `back_tree` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `NAME` varchar(30) NOT NULL,
  `url` varchar(100) DEFAULT NULL,
  `parentId` int(11) NOT NULL,
  `LEVEL` int(11) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=30 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of back_tree
-- ----------------------------
INSERT INTO `back_tree` VALUES ('1', '系统设置', '', '0', '1');
INSERT INTO `back_tree` VALUES ('2', '角色管理', '../adm/searchRole.xhtml', '1', '2');
INSERT INTO `back_tree` VALUES ('3', '账号管理', '../adm/searchMan.xhtml', '1', '2');
INSERT INTO `back_tree` VALUES ('4', '系统日志', '../adm/searchLog.xhtml', '1', '2');
INSERT INTO `back_tree` VALUES ('5', '修改密码', '../adm/toEditPw.xhtml', '1', '2');
INSERT INTO `back_tree` VALUES ('6', '菜单管理', '../adm/searchTree.xhtml', '1', '2');
INSERT INTO `back_tree` VALUES ('25', '用户管理', '', '0', '1');
INSERT INTO `back_tree` VALUES ('26', '用户管理', '../bocai/searchUser.xhtml', '25', '2');
INSERT INTO `back_tree` VALUES ('28', '充值订单', '../bocai/searchDeposit.xhtml', '25', '2');
INSERT INTO `back_tree` VALUES ('29', '提现订单', '../bocai/searchWithdraw.xhtml', '25', '2');

-- ----------------------------
-- Table structure for `back_user`
-- ----------------------------
DROP TABLE IF EXISTS `back_user`;
CREATE TABLE `back_user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(30) NOT NULL,
  `realname` varchar(30) NOT NULL,
  `password` varchar(50) NOT NULL,
  `mobile`
  `google_auth`
  `email` varchar(30) NOT NULL,
  `roleId` int(11) NOT NULL,
  `is_frozen` char
  `createtime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_username` (`username`),
  KEY `uname` (`username`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of back_user
-- ----------------------------
INSERT INTO `back_user` VALUES ('1', 'bocai_adm', '管理员', 'e10adc3949ba59abbe56e057f20f883e', '', '1', '2015-05-06 17:38:20');
INSERT INTO `back_user` VALUES ('2', 'lukuan', '路宽', '040fb9f4b5f858b592b5fedaf3347c55', '366029485@qq.com', '1', '2016-08-07 16:48:17');

-- ----------------------------
-- Table structure for `deposit_order`
-- ----------------------------
DROP TABLE IF EXISTS `deposit_order`;
CREATE TABLE `deposit_order` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `asset` varchar(255) NOT NULL,
  `amount` bigint(20) NOT NULL,
  `address` varchar(255) NOT NULL,
  `blockin_time` datetime DEFAULT NULL COMMENT '入块时间',
  `wallet_confirm_time` datetime DEFAULT NULL COMMENT '确认时间（平台确认时间）',
  `status` int(11) NOT NULL COMMENT 'blockin,confirm,unconfirm',
  `blockin_height` int(11) DEFAULT NULL,
  `wallet_current_height` int(11) NOT NULL,
  `hash` varchar(255) DEFAULT NULL,
  `remark` varchar(255) DEFAULT NULL,
  `wallet_confirm_height` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of deposit_order
-- ----------------------------
INSERT INTO `deposit_order` VALUES ('1', '10031', 'BTC', '0.0012', '1KW9QenCzuoVqmFCgeDhxHeGKMg62KPQCr', null, '2018-03-18 19:15:57', '1', null, '514067', null, '2018-03-18 19:15:57', '335b79e64af8a7568e242f75d5166d7b8143f0ded0255bae811f014d9f61cb90', '2018-03-18 19:15:57', '', '0', '513993');
INSERT INTO `deposit_order` VALUES ('10', '10044', 'BTC', '0.00125', '15w3SLCbKWK4BX91kAscnn91MoVdWn8Sqs', null, '2018-03-19 16:52:23', '1', null, '514215', null, '2018-03-19 16:52:23', '044d7beefe160fbe3437c42fe0465e859b48d88d6194d0b40c8f0e9a3676ac2b', '2018-03-19 16:52:23', '', '0', '514210');
INSERT INTO `deposit_order` VALUES ('11', '10042', 'BTC', '0.00126', '1KdEXMCBwthqtM4N54gNBAkJS5DZ1zVBhG', null, '2018-03-19 17:56:07', '1', null, '514226', null, '2018-03-19 17:56:07', '484edd4df9fc1620bc17a963211fe5b9824245149739c2d946e6f5676cb1c728', '2018-03-19 17:56:07', '', '0', '514221');
INSERT INTO `deposit_order` VALUES ('12', '10041', 'BTC', '0.00122', '18ryaW6SXqbi8QqrgxGXjEjNoWrBqSa8QE', null, '2018-03-19 17:56:11', '1', null, '514226', null, '2018-03-19 17:56:11', 'c12df5bfffaf6dc6f112b30ae10bc377294484de18f3642bd88bba46a853a417', '2018-03-19 17:56:11', '', '0', '514221');
INSERT INTO `deposit_order` VALUES ('13', '10046', 'BTC', '0.00125', '1LtYvGDDBb1MjDAEcAYXTvvSerCQ3Fo2GD', null, '2018-03-19 19:27:28', '1', null, '514235', null, '2018-03-19 19:27:28', 'c1869fb9cf961393cd860d2d539af238f89df93be58a34902c15f79174d5622f', '2018-03-19 19:27:28', '', '0', '514230');
INSERT INTO `deposit_order` VALUES ('14', '10047', 'BTC', '0.00121', '1PZ8nLU2ckuyH2Ubbxg1xxVEap3wFt5zE3', null, '2018-03-19 19:27:33', '1', null, '514235', null, '2018-03-19 19:27:33', '54035f46532f242fb71857a6b2b0bbdcdf1bd78e10a89f0ce899beed0280cacf', '2018-03-19 19:27:33', '', '0', '514230');
INSERT INTO `deposit_order` VALUES ('15', '10048', 'BTC', '0.00118', '13vVw6Vz3Sv4ZYGWvSxiA6fmJnjFYnof9u', null, '2018-03-19 19:27:38', '1', null, '514235', null, '2018-03-19 19:27:38', 'db9cd62883a3dce2559ab4acbd57714871cf5d3ce3984d6de682cc6f656da9a5', '2018-03-19 19:27:38', '', '0', '514230');
INSERT INTO `deposit_order` VALUES ('2', '10031', 'BTC', '0.001', '1KW9QenCzuoVqmFCgeDhxHeGKMg62KPQCr', null, '2018-03-18 19:16:02', '1', null, '514067', null, '2018-03-18 19:16:02', '11f95fba2237170fa82b8de88c2bd44b90b0e6fb845eebe97a0ac3b052e803eb', '2018-03-18 19:16:02', '', '0', '513993');
INSERT INTO `deposit_order` VALUES ('3', '10028', 'BTC', '0.0011', '14VzPz3A6862G9yfA7exno7VVHMBXLxie4', null, '2018-03-18 21:07:12', '1', null, '514088', null, '2018-03-18 21:07:12', '3dc0f55de20cadceb81f1eb71c5e35f8aff1952992a7a11ae41db07ffbbe9ec8', '2018-03-18 21:07:12', '', '0', '514083');
INSERT INTO `deposit_order` VALUES ('4', '10033', 'BTC', '0.001', '1EGeEUdUtPXmGrKD3XkUxY7Ryfn24ao9vP', null, '2018-03-18 22:31:21', '1', null, '514098', null, '2018-03-18 22:31:21', '6e42683d0144bb20ce4512a0877ffac8389f6f89d6f9b17ae277ad45c5e0a45f', '2018-03-18 22:31:21', '', '0', '514093');
INSERT INTO `deposit_order` VALUES ('5', '10035', 'BTC', '0.001', '17AxYUv6ZRWm7ZNWoppbpixtnyrNi1wA7G', null, '2018-03-19 00:27:56', '1', null, '514105', null, '2018-03-19 00:27:56', '937b541a7253b4c2d04457a43c77a7ee4bdcca417e41ba948547357b5e1e31c1', '2018-03-19 00:27:56', '', '0', '514100');
INSERT INTO `deposit_order` VALUES ('6', '10032', 'BTC', '0.00123', '1ccPxK5BsTtmDc24nWiauDkV4aQt8bcgJ', null, '2018-03-19 14:35:49', '1', null, '514198', null, '2018-03-19 14:35:49', 'fb186c533aa6e3e4f368f04f11011bb05be6dfbf611002021b780a99f1b681f6', '2018-03-19 14:35:49', '', '0', '514191');
INSERT INTO `deposit_order` VALUES ('7', '10039', 'BTC', '0.00119', '18idR21QEdXfEsW54XxXpnHpgsKnWFBXcA', null, '2018-03-19 14:52:31', '1', null, '514199', null, '2018-03-19 14:52:31', 'ec4f23edd33b1c39cc55f34745c3d6df6114c3254b4071333eeb9894b08db6a2', '2018-03-19 14:52:31', '', '0', '514194');
INSERT INTO `deposit_order` VALUES ('8', '10040', 'BTC', '0.0012', '12w5irBGAR7BumWpkVnbzWNZVBvBnzvKyW', null, '2018-03-19 16:47:31', '1', null, '514214', null, '2018-03-19 16:47:31', '0df0f7704bfd8b43d9fc474fd0138324bf827c9441e1fc04429a74579307cd01', '2018-03-19 16:47:31', '', '0', '514209');
INSERT INTO `deposit_order` VALUES ('9', '10043', 'BTC', '0.00118', '19yefM2uAJSsLBWo173opHS5D4XdPr4p6P', null, '2018-03-19 16:47:36', '1', null, '514214', null, '2018-03-19 16:47:36', '76c714e3931ecb4a37ce48af4dd58c4833734174dd05192b0972af2181f352af', '2018-03-19 16:47:36', '', '0', '514209');

-- ----------------------------
-- Table structure for `user`
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_name` varchar(255) NOT NULL,
  `phone` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `password` varchar(255) NOT NULL,
  `google_auth` varchar(255) NOT NULL,
  `license_key` varchar(255) DEFAULT NULL,
  `public_key` varchar(255) DEFAULT NULL,
  `level` int(11) NOT NULL DEFAULT '0' COMMENT 'vip级别',
  `last_login_time` datetime DEFAULT NULL,
  `last_login_ip` varchar(255) DEFAULT NULL,
  `last_login_mac` varchar(255) DEFAULT NULL,
  `create_date` datetime NOT NULL,
  `update_date` datetime NOT NULL COMMENT '用户信息更改，对应日志待定',
  `is_frozen` char NOT NULL DEFAULT '0',
  `time_zone` varchar(255) DEFAULT NULL,
  `conutry` varchar(255) DEFAULT NULL,
  `language` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10052 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of user
-- ----------------------------
INSERT INTO `user` VALUES ('8888', '13818078927', '提币佣金专户', null, 'H3NJgOtmpnlmgWQ1sFYbRw==', null, '0', null, '2018-03-22 21:13:03', '0:0:0:0:0:0:0:1', '2018-03-14 21:05:04', '2018-03-22 21:30:44', '0', '0', null, null);
INSERT INTO `user` VALUES ('9999', '18816504095', '交易佣金专户', null, 'H3NJgOtmpnlmgWQ1sFYbRw==', null, '0', null, '2018-03-18 03:42:32', '61.152.132.170', '2018-03-14 21:05:07', '2018-03-18 04:19:16', '0', '0', null, null);
INSERT INTO `user` VALUES ('10028', '13370080511', 'å¤§é»å®¢', null, '5Urk3TFYmvF+PH01KdsG0Q==', null, '0', null, '2018-03-22 10:52:59', '117.136.8.70', '2018-03-18 01:37:36', '2018-03-22 11:10:43', '0', '0', null, null);
INSERT INTO `user` VALUES ('10031', '13918315909', 'è´å£³é»å®¢', null, 'H9CK7fzui7CU6tPcZqXy9A==', null, '0', null, '2018-03-19 18:47:11', '58.33.53.27', '2018-03-18 05:07:05', '2018-03-19 19:24:19', '0', '0', null, null);
INSERT INTO `user` VALUES ('10032', '13585596201', 'henly', null, 'myM08b2k2Ug7APvbpGGBFA==', null, '0', null, '2018-03-21 23:33:49', '42.196.37.153', '2018-03-18 10:55:16', '2018-03-21 23:51:37', '0', '0', null, null);
INSERT INTO `user` VALUES ('10033', '13916688387', 'hume', null, 'FTEMBHTeNRfkE33a/yWXTw==', null, '0', null, '2018-03-19 21:11:13', '223.104.212.48', '2018-03-18 21:40:57', '2018-03-19 21:48:23', '0', '0', null, null);
INSERT INTO `user` VALUES ('10034', '17610057589', 'ocean', null, 'lZre785DUjesmkuTjCiWsQ==', null, '0', null, '2018-03-18 14:05:08', '114.242.250.163', '2018-03-18 22:41:37', '2018-03-18 22:42:03', '0', '0', null, null);
INSERT INTO `user` VALUES ('10035', '15699881878', 'å®ç¾ä½å', null, 'lC3+69hBU+CGWoAHjw+ykg==', null, '0', null, '2018-03-19 01:47:56', '101.88.232.107', '2018-03-18 22:54:02', '2018-03-19 10:24:59', '0', '0', null, null);
INSERT INTO `user` VALUES ('10036', '18621198705', 'lee', null, 'p4QTg9Tkxgj4A9N3aDe/9g==', null, '0', null, '2018-03-18 14:27:41', '115.204.105.163', '2018-03-18 23:04:01', '2018-03-18 23:04:37', '0', '0', null, null);
INSERT INTO `user` VALUES ('10037', '15957503088', 'Jake', null, 'zfUSJzlxVpzGB8QM+pWDjA==', null, '0', null, '2018-03-18 14:39:36', '139.162.70.18', '2018-03-18 23:15:08', '2018-03-18 23:16:32', '0', '0', null, null);
INSERT INTO `user` VALUES ('10038', '18611539701', 'yy', null, 'o9lDadiTrir3bhdhME1tfw==', null, '0', null, '2018-03-18 15:00:22', '45.76.169.180', '2018-03-18 23:36:53', '2018-03-18 23:37:18', '0', '0', null, null);
INSERT INTO `user` VALUES ('10039', '18623103711', 'lamp.z', null, 'JaEeWGjnykB6+nTEruQGaA==', null, '0', null, '2018-03-19 04:29:44', '61.152.132.170', '2018-03-19 13:06:13', '2018-03-19 13:06:49', '0', '0', null, null);
INSERT INTO `user` VALUES ('10040', '18512175375', 'Richard', null, 'KIJP962M7zBhxAmFacHlBA==', null, '0', null, '2018-03-19 17:24:17', '103.192.224.118', '2018-03-19 13:25:54', '2018-03-19 18:01:24', '0', '0', null, null);
INSERT INTO `user` VALUES ('10041', '15221097340', 'Tom', null, '5TKiJZSGQkwA8BxSCnl0UA==', null, '0', null, '2018-03-19 16:41:15', '61.152.132.170', '2018-03-19 13:46:48', '2018-03-19 17:18:21', '0', '0', null, null);
INSERT INTO `user` VALUES ('10042', '13681602570', 'zyn', null, 't35EUieqhLaB8rLTPcZ81A==', null, '0', null, '2018-03-19 17:32:33', '61.152.132.170', '2018-03-19 13:49:05', '2018-03-19 18:09:40', '0', '0', null, null);
INSERT INTO `user` VALUES ('10043', '18621509572', 'funnay', null, 'o6HDEZjU+sKbQxeyAIlGzQ==', null, '0', null, '2018-03-19 16:11:07', '61.152.132.170', '2018-03-19 13:50:28', '2018-03-19 16:48:13', '0', '0', null, null);
INSERT INTO `user` VALUES ('10044', '18930350677', 'huhuh', null, 'iAbDwt7aJaqtbqbHg2oyAg==', null, '0', null, '2018-03-19 18:34:07', '101.90.252.128', '2018-03-19 16:05:38', '2018-03-19 19:11:15', '0', '0', null, null);
INSERT INTO `user` VALUES ('10046', '17301728029', 'ä¸è½å°ç±', null, 'YPlShfDWkig0kE231Hvf1w==', null, '0', null, '2018-03-19 17:24:29', '61.152.132.170', '2018-03-19 17:33:12', '2018-03-19 18:01:37', '0', '0', null, null);
INSERT INTO `user` VALUES ('10047', '13524233134', 'æ ç', null, 'zyKvBJNQHGx5YtUlxzMoFw==', null, '0', null, '2018-03-19 17:24:38', '61.152.132.170', '2018-03-19 17:39:02', '2018-03-19 18:01:45', '0', '0', null, null);
INSERT INTO `user` VALUES ('10048', '18018851831', 'Allan Liu', null, 'sUjHfx4VcPP93mZhccDkbw==', null, '0', null, '2018-03-19 17:38:30', '101.90.125.41', '2018-03-19 18:14:58', '2018-03-19 18:15:38', '0', '0', null, null);
INSERT INTO `user` VALUES ('10049', '13646610434', 'mochen999', null, 'ZAkEl/5EDTPIz06/CjFeKw==', null, '0', null, '2018-03-20 15:51:40', '117.136.41.120, 113.96.219.249', '2018-03-20 16:28:47', '2018-03-20 16:29:00', '0', '0', null, null);
INSERT INTO `user` VALUES ('10050', '15900692149', 'Fengzitang', null, 'vSZNIG5g26pIXx1C6FmhTQ==', null, '0', null, '2018-03-21 12:11:35', '223.104.5.168', '2018-03-21 11:32:22', '2018-03-21 12:29:25', '0', '0', null, null);
INSERT INTO `user` VALUES ('10051', '18702198187', '巴迪', null, 'H3NJgOtmpnlmgWQ1sFYbRw==', null, '0', null, '2018-03-22 10:39:28', '223.104.212.218', '2018-03-21 13:10:39', '2018-03-22 10:57:12', '0', '0', null, null);

-- ----------------------------
-- Table structure for `user_asset_address`
-- ----------------------------
DROP TABLE IF EXISTS `user_asset_address`;
CREATE TABLE `user_asset_address` (
  `user_id` int(11) NOT NULL,
  `asset` varchar(255) NOT NULL,
  `address` varchar(255) NOT NULL DEFAULT '0',
  `enabled` int(1) NOT NULL DEFAULT 1,
  `create_time` datetime NOT NULL,
  UNIQUE KEY `user_id_asset_address` (`user_id`,`asset`,`address`) USING BTREE,
  UNIQUE KEY `user_id_asset` (`user_id`,`asset`) USING BTREE,
  UNIQUE KEY `asset_adderss` (`asset`,`address`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT;

-- ----------------------------
-- Records of user_asset_address
-- ----------------------------
INSERT INTO `user_asset_address` VALUES ('8888', 'uBTC', '177nscaEjPEEPaV3w5vkdu6WAQo4ofTyYt', '2018-03-18 05:22:13');
INSERT INTO `user_asset_address` VALUES ('10028', 'uBTC', '14VzPz3A6862G9yfA7exno7VVHMBXLxie4', '2018-03-18 20:29:13');
INSERT INTO `user_asset_address` VALUES ('10030', 'uBTC', '1Fb2RDQ2c2FBdUCnhwHvuA8o25rmibh5q3', '2018-03-18 02:54:06');
INSERT INTO `user_asset_address` VALUES ('10031', 'uBTC', '1KW9QenCzuoVqmFCgeDhxHeGKMg62KPQCr', '2018-03-18 05:31:20');
INSERT INTO `user_asset_address` VALUES ('10032', 'uBTC', '1ccPxK5BsTtmDc24nWiauDkV4aQt8bcgJ', '2018-03-19 12:53:35');
INSERT INTO `user_asset_address` VALUES ('10033', 'uBTC', '1EGeEUdUtPXmGrKD3XkUxY7Ryfn24ao9vP', '2018-03-18 21:42:24');
INSERT INTO `user_asset_address` VALUES ('10034', 'uBTC', '1Ld4bq1PcS5ka6U52747c8wuqB5aQoQ37k', '2018-03-18 22:44:14');
INSERT INTO `user_asset_address` VALUES ('10035', 'uBTC', '17AxYUv6ZRWm7ZNWoppbpixtnyrNi1wA7G', '2018-03-18 23:14:48');
INSERT INTO `user_asset_address` VALUES ('10038', 'uBTC', '14smRb1RimoT428fVfP6qEHhsuAe3yoKi4', '2018-03-18 23:38:04');
INSERT INTO `user_asset_address` VALUES ('10039', 'uBTC', '18idR21QEdXfEsW54XxXpnHpgsKnWFBXcA', '2018-03-19 13:07:31');
INSERT INTO `user_asset_address` VALUES ('10040', 'uBTC', '12w5irBGAR7BumWpkVnbzWNZVBvBnzvKyW', '2018-03-19 13:26:51');
INSERT INTO `user_asset_address` VALUES ('10041', 'uBTC', '18ryaW6SXqbi8QqrgxGXjEjNoWrBqSa8QE', '2018-03-19 17:18:35');
INSERT INTO `user_asset_address` VALUES ('10042', 'uBTC', '1KdEXMCBwthqtM4N54gNBAkJS5DZ1zVBhG', '2018-03-19 17:18:07');
INSERT INTO `user_asset_address` VALUES ('10043', 'uBTC', '19yefM2uAJSsLBWo173opHS5D4XdPr4p6P', '2018-03-19 16:03:24');
INSERT INTO `user_asset_address` VALUES ('10044', 'uBTC', '15w3SLCbKWK4BX91kAscnn91MoVdWn8Sqs', '2018-03-19 16:10:01');
INSERT INTO `user_asset_address` VALUES ('10046', 'uBTC', '1LtYvGDDBb1MjDAEcAYXTvvSerCQ3Fo2GD', '2018-03-19 18:02:35');
INSERT INTO `user_asset_address` VALUES ('10047', 'uBTC', '1PZ8nLU2ckuyH2Ubbxg1xxVEap3wFt5zE3', '2018-03-19 18:01:51');
INSERT INTO `user_asset_address` VALUES ('10048', 'uBTC', '13vVw6Vz3Sv4ZYGWvSxiA6fmJnjFYnof9u', '2018-03-19 18:16:39');

-- ----------------------------
-- Table structure for `user_assets`
-- ----------------------------
DROP TABLE IF EXISTS `user_assets`;
CREATE TABLE `user_assets` (
  `user_id` int(11) NOT NULL,
  `asset` varchar(255) NOT NULL,
  `available_amount` double(255,0) NOT NULL DEFAULT '0',
  `frozen_amount` double(255,0) NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`user_id`,`asset`),
  UNIQUE KEY `user_id_asset` (`user_id`,`asset`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of user_assets
-- ----------------------------
INSERT INTO `user_assets` VALUES ('8888', 'uBTC', '918912', '0', '0000-00-00 00:00:00', '0000-00-00 00:00:00');
INSERT INTO `user_assets` VALUES ('9999', 'uBTC', '8002', '0', '0000-00-00 00:00:00', '0000-00-00 00:00:00');
INSERT INTO `user_assets` VALUES ('10028', 'uBTC', '100', '0', '2018-03-18 01:37:36', '2018-03-18 01:37:36');
INSERT INTO `user_assets` VALUES ('10030', 'uBTC', '0', '0', '2018-03-18 01:44:52', '2018-03-18 01:44:52');
INSERT INTO `user_assets` VALUES ('10031', 'uBTC', '2100', '0', '2018-03-18 05:07:06', '2018-03-18 05:07:06');
INSERT INTO `user_assets` VALUES ('10032', 'uBTC', '1230', '0', '2018-03-18 10:55:16', '2018-03-18 10:55:16');
INSERT INTO `user_assets` VALUES ('10033', 'uBTC', '600', '0', '2018-03-18 21:40:57', '2018-03-18 21:40:57');
INSERT INTO `user_assets` VALUES ('10034', 'uBTC', '0', '0', '2018-03-18 22:41:38', '2018-03-18 22:41:38');
INSERT INTO `user_assets` VALUES ('10035', 'uBTC', '752', '0', '2018-03-18 22:54:03', '2018-03-18 22:54:03');
INSERT INTO `user_assets` VALUES ('10036', 'uBTC', '0', '0', '2018-03-18 23:04:02', '2018-03-18 23:04:02');
INSERT INTO `user_assets` VALUES ('10037', 'uBTC', '0', '0', '2018-03-18 23:15:09', '2018-03-18 23:15:09');
INSERT INTO `user_assets` VALUES ('10038', 'uBTC', '0', '0', '2018-03-18 23:36:54', '2018-03-18 23:36:54');
INSERT INTO `user_assets` VALUES ('10039', 'uBTC', '1190', '0', '2018-03-19 13:06:14', '2018-03-19 13:06:14');
INSERT INTO `user_assets` VALUES ('10040', 'uBTC', '1200', '0', '2018-03-19 13:25:54', '2018-03-19 13:25:54');
INSERT INTO `user_assets` VALUES ('10041', 'uBTC', '1220', '0', '2018-03-19 13:46:49', '2018-03-19 13:46:49');
INSERT INTO `user_assets` VALUES ('10042', 'uBTC', '1260', '0', '2018-03-19 13:49:06', '2018-03-19 13:49:06');
INSERT INTO `user_assets` VALUES ('10043', 'uBTC', '1180', '0', '2018-03-19 13:50:28', '2018-03-19 13:50:28');
INSERT INTO `user_assets` VALUES ('10044', 'uBTC', '1699', '0', '2018-03-19 16:05:39', '2018-03-19 16:05:39');
INSERT INTO `user_assets` VALUES ('10045', 'uBTC', '0', '0', '2018-03-19 17:28:18', '2018-03-19 17:28:18');
INSERT INTO `user_assets` VALUES ('10046', 'uBTC', '1250', '0', '2018-03-19 17:33:13', '2018-03-19 17:33:13');
INSERT INTO `user_assets` VALUES ('10047', 'uBTC', '1210', '0', '2018-03-19 17:39:03', '2018-03-19 17:39:03');
INSERT INTO `user_assets` VALUES ('10048', 'uBTC', '1180', '0', '2018-03-19 18:14:59', '2018-03-19 18:14:59');
INSERT INTO `user_assets` VALUES ('10049', 'uBTC', '0', '0', '2018-03-20 16:28:47', '2018-03-20 16:28:47');
INSERT INTO `user_assets` VALUES ('10050', 'uBTC', '0', '0', '2018-03-21 11:32:22', '2018-03-21 11:32:22');
INSERT INTO `user_assets` VALUES ('10051', 'uBTC', '0', '0', '2018-03-21 13:10:39', '2018-03-21 13:10:39');

-- ----------------------------
-- Table structure for `user_ledger`
-- ----------------------------
DROP TABLE IF EXISTS `user_ledger`;
CREATE TABLE `user_ledger` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `order_id` int(11)
  `from_user_id` int(11) NOT NULL,
  `to_user_id` int(11) NOT NULL,
  `from_addr`
  `to_addr`
  `asset` varchar(255) NOT NULL,
  `amount` double(255,0) NOT NULL,
  `update_date` datetime NOT NULL,
  `transaction_type` varchar(255) NOT NULL COMMENT '充值，提币，提币手续费，提币矿工费，手续费利润，冷冷矿工费，冷热矿工费',
  --`balance` double NOT NULL COMMENT '变化后余额',
  `remark` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
  --UNIQUE KEY `user_id_asset` (`user_id`,`asset`)
) ENGINE=InnoDB AUTO_INCREMENT=22 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of user_ledger
-- ----------------------------
INSERT INTO `user_ledger` VALUES ('7', '10031', 'uBTC', '1200', '2018-03-18 19:15:59', 'IN', '1200', null);
INSERT INTO `user_ledger` VALUES ('8', '10031', 'uBTC', '1000', '2018-03-18 19:16:03', 'IN', '2200', null);
INSERT INTO `user_ledger` VALUES ('9', '10028', 'uBTC', '1100', '2018-03-18 21:07:14', 'IN', '1100', null);
INSERT INTO `user_ledger` VALUES ('10', '10033', 'uBTC', '1000', '2018-03-18 22:31:23', 'IN', '1000', null);
INSERT INTO `user_ledger` VALUES ('11', '10035', 'uBTC', '1000', '2018-03-19 00:27:58', 'IN', '1000', null);
INSERT INTO `user_ledger` VALUES ('12', '10032', 'uBTC', '1230', '2018-03-19 14:35:51', 'IN', '1230', null);
INSERT INTO `user_ledger` VALUES ('13', '10039', 'uBTC', '1190', '2018-03-19 14:52:32', 'IN', '1190', null);
INSERT INTO `user_ledger` VALUES ('14', '10040', 'uBTC', '1200', '2018-03-19 16:47:33', 'IN', '1200', null);
INSERT INTO `user_ledger` VALUES ('15', '10043', 'uBTC', '1180', '2018-03-19 16:47:38', 'IN', '1180', null);
INSERT INTO `user_ledger` VALUES ('16', '10044', 'uBTC', '1250', '2018-03-19 16:52:24', 'IN', '1250', null);
INSERT INTO `user_ledger` VALUES ('17', '10042', 'uBTC', '1260', '2018-03-19 17:56:08', 'IN', '1260', null);
INSERT INTO `user_ledger` VALUES ('18', '10041', 'uBTC', '1220', '2018-03-19 17:56:13', 'IN', '1220', null);
INSERT INTO `user_ledger` VALUES ('19', '10046', 'uBTC', '1250', '2018-03-19 19:27:30', 'IN', '1250', null);
INSERT INTO `user_ledger` VALUES ('20', '10047', 'uBTC', '1210', '2018-03-19 19:27:35', 'IN', '1210', null);
INSERT INTO `user_ledger` VALUES ('21', '10048', 'uBTC', '1180', '2018-03-19 19:27:40', 'IN', '1180', null);

-- ----------------------------
-- Table structure for `user_regis_channel`
-- ----------------------------
DROP TABLE IF EXISTS `user_regis_channel`;
CREATE TABLE `user_regis_channel` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) DEFAULT NULL,
  `fk_id` varchar(255) DEFAULT NULL,
  `fk_type` char(2) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid` (`user_id`,`fk_id`,`fk_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of user_regis_channel
-- ----------------------------

-- ----------------------------
-- Table structure for `withdraw_order`
-- ----------------------------
DROP TABLE IF EXISTS `withdrawal_order`;
CREATE TABLE `withdraw_order` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `user_order_id` varchar(255) NOT NULL,
  `withdrawal_amount` double NOT NULL,
  `asset` varchar(255) NOT NULL,
  `blockin_time` datetime DEFAULT NULL,
  `wallet_confirm_time` datetime DEFAULT NULL,
  `blockin_height` int(11) DEFAULT NULL,
  `wallet_current_height` int(11) DEFAULT NULL,
  `address` varchar(255) NOT NULL,
  `status` int(1) NOT NULL COMMENT '0初始状态1待人工审核2人工审核通过3提现请求待发送4请求已发送5钱包成功6钱包拒绝7人工审核拒绝',
  `hash` varchar(255) DEFAULT NULL,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  `remark` varchar(255) DEFAULT NULL,
  `reviewer_id` int(11) DEFAULT NULL,
  `inspect_result` varchar(255) DEFAULT NULL COMMENT '可能有多条检测结果',
  `withdrawal_wallet_fee` double DEFAULT NULL COMMENT '提币费',
  `withdrawal_miners_fee` double DEFAULT NULL COMMENT '矿工费' 
  `wallet_confirm_height` bigint(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of withdraw_order
-- ----------------------------

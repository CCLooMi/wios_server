/*
 Navicat Premium Data Transfer

 Source Server         : local3308
 Source Server Type    : MySQL
 Source Server Version : 50742
 Source Host           : localhost:3308
 Source Schema         : wios

 Target Server Type    : MySQL
 Target Server Version : 50742
 File Encoding         : 65001

 Date: 22/07/2023 14:39:30
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for sys_menu
-- ----------------------------
DROP TABLE IF EXISTS `sys_menu`;
CREATE TABLE `sys_menu`  (
  `id` binary(16) NOT NULL COMMENT '主键ID',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '名称',
  `url` varchar(256) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '地址',
  `pid` binary(16) NULL DEFAULT NULL COMMENT '上级权限ID',
  `icon` longtext CHARACTER SET utf8 COLLATE utf8_general_ci NULL COMMENT '图标',
  `type` varchar(16) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '菜单类型',
  `rootId` binary(16) NULL DEFAULT NULL COMMENT '根菜单ID',
  `idx` int(11) NULL DEFAULT NULL COMMENT '层级深度',
  `position` int(11) NULL DEFAULT NULL COMMENT '位置',
  `inserted_at` datetime NULL DEFAULT NULL COMMENT '创建日期',
  `updated_at` datetime NULL DEFAULT NULL COMMENT '更新日期',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_org
-- ----------------------------
DROP TABLE IF EXISTS `sys_org`;
CREATE TABLE `sys_org`  (
  `id` binary(16) NOT NULL COMMENT '主键ID',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '组织名称',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '组织描述',
  `inserted_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '组织信息表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_org_user
-- ----------------------------
DROP TABLE IF EXISTS `sys_org_user`;
CREATE TABLE `sys_org_user`  (
  `id` binary(16) NOT NULL COMMENT '主键ID',
  `user_id` binary(16) NULL DEFAULT NULL COMMENT '用户ID',
  `org_id` binary(16) NULL DEFAULT NULL COMMENT '组织ID',
  `inserted_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '用户组织关系表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_permission
-- ----------------------------
DROP TABLE IF EXISTS `sys_permission`;
CREATE TABLE `sys_permission`  (
  `id` binary(16) NOT NULL COMMENT '主键ID',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '权限名称',
  `type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '权限类型',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '权限描述',
  `inserted_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '权限信息表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_role
-- ----------------------------
DROP TABLE IF EXISTS `sys_role`;
CREATE TABLE `sys_role`  (
  `id` binary(16) NOT NULL COMMENT '主键ID',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '角色名称',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '角色描述',
  `inserted_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '角色表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_role_menu
-- ----------------------------
DROP TABLE IF EXISTS `sys_role_menu`;
CREATE TABLE `sys_role_menu`  (
  `id` binary(16) NOT NULL COMMENT 'ID',
  `role_id` binary(16) NULL DEFAULT NULL COMMENT '角色ID',
  `menu_id` binary(16) NULL DEFAULT NULL COMMENT '视图ID',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_role_permission
-- ----------------------------
DROP TABLE IF EXISTS `sys_role_permission`;
CREATE TABLE `sys_role_permission`  (
  `id` binary(16) NOT NULL COMMENT '主键ID',
  `role_id` binary(16) NULL DEFAULT NULL COMMENT '角色ID',
  `permission_id` binary(16) NULL DEFAULT NULL COMMENT '权限ID',
  `inserted_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '角色权限关系表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_role_user
-- ----------------------------
DROP TABLE IF EXISTS `sys_role_user`;
CREATE TABLE `sys_role_user`  (
  `id` binary(16) NOT NULL COMMENT '主键ID',
  `user_id` binary(16) NULL DEFAULT NULL COMMENT '用户ID',
  `role_id` binary(16) NULL DEFAULT NULL COMMENT '角色ID',
  `inserted_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '用户角色关系表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_upload
-- ----------------------------
DROP TABLE IF EXISTS `sys_upload`;
CREATE TABLE `sys_upload`  (
  `id` binary(16) NOT NULL COMMENT '主键ID',
  `file_id` varbinary(32) NULL DEFAULT NULL COMMENT '文件ID',
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '文件名称',
  `file_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '文件类型',
  `file_size` bigint(20) NULL DEFAULT NULL COMMENT '文件大小',
  `biz_id` binary(16) NULL DEFAULT NULL COMMENT '业务ID',
  `biz_type` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '业务类型',
  `inserted_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '上传文件表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_user
-- ----------------------------
DROP TABLE IF EXISTS `sys_user`;
CREATE TABLE `sys_user`  (
  `id` binary(16) NOT NULL COMMENT '主键ID',
  `username` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '用户名',
  `password` varbinary(32) NOT NULL COMMENT '用户密码',
  `seed` binary(8) NOT NULL COMMENT '密码种子',
  `inserted_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '用户表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for t_account
-- ----------------------------
DROP TABLE IF EXISTS `t_account`;
CREATE TABLE `t_account`  (
  `id` binary(16) NOT NULL COMMENT '主键ID',
  `user_id` binary(16) NULL DEFAULT NULL COMMENT '用户ID',
  `balance` decimal(19, 2) NULL DEFAULT NULL COMMENT '资金',
  `inserted_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '账户表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for t_category
-- ----------------------------
DROP TABLE IF EXISTS `t_category`;
CREATE TABLE `t_category`  (
  `id` binary(16) NOT NULL COMMENT '主键ID',
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '分类名称',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '分类描述',
  `order` int(11) NULL DEFAULT 0 COMMENT '分类排序',
  `inserted_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '分类表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for t_comment
-- ----------------------------
DROP TABLE IF EXISTS `t_comment`;
CREATE TABLE `t_comment`  (
  `id` binary(16) NOT NULL COMMENT '主键ID',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '评论内容',
  `rating` int(11) NULL DEFAULT NULL COMMENT '评分',
  `user_id` binary(16) NULL DEFAULT NULL COMMENT '用户ID',
  `target_id` binary(16) NULL DEFAULT NULL COMMENT '目标ID',
  `root_id` binary(16) NULL DEFAULT NULL COMMENT '根ID',
  `inserted_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '评论表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for t_purchased_wpp
-- ----------------------------
DROP TABLE IF EXISTS `t_purchased_wpp`;
CREATE TABLE `t_purchased_wpp`  (
  `id` binary(16) NOT NULL COMMENT '主键ID',
  `user_id` binary(16) NULL DEFAULT NULL COMMENT '用户ID',
  `wpp_id` binary(16) NULL DEFAULT NULL COMMENT '应用ID',
  `price` decimal(10, 0) NULL DEFAULT NULL COMMENT '购买价格',
  `purchase_time` datetime(6) NULL DEFAULT NULL COMMENT '购买时间',
  `inserted_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '已购买应用表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for t_wpps
-- ----------------------------
DROP TABLE IF EXISTS `t_wpps`;
CREATE TABLE `t_wpps`  (
  `id` binary(16) NOT NULL COMMENT '主键ID',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '应用名称',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '描述',
  `version` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '版本号',
  `developer_id` binary(16) NULL DEFAULT NULL COMMENT '开发者ID',
  `file_id` varbinary(32) NULL DEFAULT NULL COMMENT '文件ID',
  `inserted_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '应用表' ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;

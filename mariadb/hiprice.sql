/*
 Navicat Premium Data Transfer

 Source Server Type    : MariaDB
 Source Server Version : 100309
 Source Schema         : hiprice

 Target Server Type    : MariaDB
 Target Server Version : 100309
 File Encoding         : 65001

 Date: 17/09/2018 10:40:42
*/

CREATE DATABASE IF NOT EXISTS `hiprice` DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_general_ci;

USE `hiprice`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for msg
-- ----------------------------
DROP TABLE IF EXISTS `msg`;
CREATE TABLE `msg`  (
  `_id` int(11) NOT NULL AUTO_INCREMENT,
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `from_user_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `to_user_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `type` int(11) NOT NULL DEFAULT 0,
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '\'\'',
  `url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_time` datetime(0) NOT NULL DEFAULT '0001-01-01 00:00:00',
  `raw` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '\'\'',
  PRIMARY KEY (`_id`) USING BTREE,
  UNIQUE INDEX `msg_id`(`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for product
-- ----------------------------
DROP TABLE IF EXISTS `product`;
CREATE TABLE `product`  (
  `_id` int(11) NOT NULL AUTO_INCREMENT,
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `source` int(11) NOT NULL DEFAULT 0,
  `url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `short_url` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `title` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `currency` int(11) NOT NULL DEFAULT 0,
  `price` float NOT NULL DEFAULT -1,
  `price_low` float NOT NULL DEFAULT 0,
  `price_high` float NOT NULL DEFAULT 0,
  `stock` int(11) NOT NULL DEFAULT -1,
  `sales` int(11) NOT NULL DEFAULT -1,
  `category` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `comments` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '\'\'',
  `update_time` datetime(0) NOT NULL DEFAULT '0001-01-01 00:00:00',
  `last_dispatch_time` datetime(0) NOT NULL DEFAULT '0001-01-01 00:00:00',
  PRIMARY KEY (`_id`) USING BTREE,
  UNIQUE INDEX `product_id`(`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for product_update
-- ----------------------------
DROP TABLE IF EXISTS `product_update`;
CREATE TABLE `product_update`  (
  `_id` int(11) NOT NULL AUTO_INCREMENT,
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `source` int(11) NOT NULL DEFAULT 0,
  `url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `short_url` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `title` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `currency` int(11) NOT NULL DEFAULT 0,
  `price` float NOT NULL DEFAULT -1,
  `price_low` float NOT NULL DEFAULT 0,
  `price_high` float NOT NULL DEFAULT 0,
  `stock` int(11) NOT NULL DEFAULT -1,
  `sales` int(11) NOT NULL DEFAULT -1,
  `category` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `comments` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '\'\'',
  `update_time` datetime(0) NOT NULL DEFAULT '0001-01-01 00:00:00',
  PRIMARY KEY (`_id`) USING BTREE,
  INDEX `product_update_id`(`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for product_watch
-- ----------------------------
DROP TABLE IF EXISTS `product_watch`;
CREATE TABLE `product_watch`  (
  `_id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `product_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `currency` int(11) NOT NULL DEFAULT 0,
  `price` float NOT NULL DEFAULT -1,
  `price_low` float NOT NULL DEFAULT 0,
  `price_high` float NOT NULL DEFAULT 0,
  `stock` int(11) NOT NULL DEFAULT -1,
  `watch_time` datetime(0) NOT NULL DEFAULT '0001-01-01 00:00:00',
  `unwatch_time` datetime(0) NOT NULL DEFAULT '0001-01-01 00:00:00',
  `state` int(11) NOT NULL DEFAULT 0,
  `remind_decrease_option` int(11) NOT NULL DEFAULT 2,
  `remind_decrease_value` float NOT NULL DEFAULT 5,
  `remind_increase_option` int(11) NOT NULL DEFAULT 2,
  `remind_increase_value` float NOT NULL DEFAULT 5,
  PRIMARY KEY (`_id`) USING BTREE,
  INDEX `product_watch_user_id`(`user_id`) USING BTREE,
  INDEX `product_watch_product_id`(`product_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for suggestion
-- ----------------------------
DROP TABLE IF EXISTS `suggestion`;
CREATE TABLE `suggestion`  (
  `_id` int(11) NOT NULL AUTO_INCREMENT,
  `msg_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `from_user_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '\'\'',
  PRIMARY KEY (`_id`) USING BTREE,
  UNIQUE INDEX `suggestion_msg_id`(`msg_id`) USING BTREE,
  INDEX `suggestion_from_user_id`(`from_user_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user`  (
  `_id` int(11) NOT NULL AUTO_INCREMENT,
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `nickname` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_time` datetime(0) NOT NULL DEFAULT '0001-01-01 00:00:00',
  `uin` int(11) NOT NULL DEFAULT 0,
  `disturb` int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`_id`) USING BTREE,
  UNIQUE INDEX `user_id`(`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;

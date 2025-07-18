-- 设置字符集
SET NAMES utf8mb4;

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS haven_camp_server 
DEFAULT CHARACTER SET utf8mb4 
DEFAULT COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE haven_camp_server;

-- 创建用户（如果不存在）
CREATE USER IF NOT EXISTS 'havencamp'@'%' IDENTIFIED BY 'havencamp123';

-- 授予权限
GRANT ALL PRIVILEGES ON haven_camp_server.* TO 'havencamp'@'%';

-- 刷新权限
FLUSH PRIVILEGES;

-- 注意：实际的表结构将由GORM自动迁移创建
-- 这里只是初始化数据库和用户权限 
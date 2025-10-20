-- ============================================
-- HavenCamp AI 虚拟用户初始化脚本
-- ============================================
-- 说明: 
--   此脚本用于在数据库中创建 AI 虚拟用户记录
--   这是可选操作,不创建也不影响 AI 对话功能
--   但创建后前端可以更好地展示 AI 的昵称和头像
-- ============================================

USE haven_camp_server;

-- 检查 AI 用户是否已存在
SELECT 
    CASE 
        WHEN EXISTS (SELECT 1 FROM user_info WHERE uuid = 'UAI000000000')
        THEN 'AI 用户已存在,无需重复创建'
        ELSE 'AI 用户不存在,即将创建'
    END AS status;

-- 插入 AI 虚拟用户(如果不存在)
INSERT INTO user_info (
    uuid, 
    nickname, 
    telephone, 
    email, 
    avatar, 
    gender, 
    signature, 
    password, 
    birthday, 
    created_at, 
    is_admin, 
    status
)
SELECT 
    'UAI000000000' AS uuid,
    'AI助手' AS nickname,
    '00000000000' AS telephone,
    'ai@havencamp.com' AS email,
    'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png' AS avatar,
    0 AS gender,
    '我是 AI 助手,随时为您服务' AS signature,
    '' AS password,
    '' AS birthday,
    NOW() AS created_at,
    0 AS is_admin,
    0 AS status
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM user_info WHERE uuid = 'UAI000000000'
);

-- 验证插入结果
SELECT 
    uuid,
    nickname,
    avatar,
    signature,
    created_at,
    status
FROM user_info 
WHERE uuid = 'UAI000000000';

-- ============================================
-- 可选: 自定义 AI 用户信息
-- ============================================
-- 如果你想修改 AI 的昵称、头像或签名,可以执行以下 SQL:
/*
UPDATE user_info 
SET 
    nickname = '你的 AI 名称',
    avatar = 'https://你的头像URL',
    signature = '你的个性签名'
WHERE uuid = 'UAI000000000';
*/

-- ============================================
-- 可选: 查看与 AI 的对话记录
-- ============================================
-- 查看最近 10 条与 AI 的对话
/*
SELECT 
    m.uuid AS message_id,
    m.send_id,
    m.send_name,
    m.receive_id,
    m.content,
    m.created_at
FROM message m
WHERE m.send_id = 'UAI000000000' 
   OR m.receive_id = 'UAI000000000'
ORDER BY m.created_at DESC
LIMIT 10;
*/

-- 查看与 AI 的会话
/*
SELECT 
    s.uuid AS session_id,
    s.send_id,
    s.receive_id,
    s.created_at
FROM session s
WHERE s.send_id = 'UAI000000000' 
   OR s.receive_id = 'UAI000000000'
ORDER BY s.created_at DESC;
*/

-- ============================================
-- 可选: 删除 AI 用户(慎用!)
-- ============================================
-- 如果需要删除 AI 用户及相关数据,执行以下 SQL:
/*
-- 注意: 这将删除所有与 AI 的对话记录和会话!

-- 1. 删除消息
DELETE FROM message 
WHERE send_id = 'UAI000000000' 
   OR receive_id = 'UAI000000000';

-- 2. 删除会话
DELETE FROM session 
WHERE send_id = 'UAI000000000' 
   OR receive_id = 'UAI000000000';

-- 3. 删除用户
DELETE FROM user_info 
WHERE uuid = 'UAI000000000';

-- 验证删除
SELECT COUNT(*) AS remaining_records
FROM user_info 
WHERE uuid = 'UAI000000000';
*/

-- ============================================
-- 完成!
-- ============================================
-- 现在可以:
-- 1. 启动 HavenCamp 服务
-- 2. 调用 POST /ai/chat 接口
-- 3. 享受 AI 对话功能
-- 
-- 详细文档:
-- - docs/AI对话快速开始.md
-- - docs/AI对话接口文档.md
-- - docs/AI对话技术实现文档.md
-- ============================================


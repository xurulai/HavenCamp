-- 测试用户数据
USE haven_camp_server;

-- 插入测试用户账号
INSERT INTO `user_info` (
    `uuid`, `nickname`, `telephone`, `email`, `avatar`, `gender`, 
    `signature`, `password`, `birthday`, `created_at`, `deleted_at`, 
    `last_online_at`, `last_offline_at`, `is_admin`, `status`
) VALUES 
(
    'U12345678901', 
    '测试用户1', 
    '18888888888', 
    'test1@example.com', 
    'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png', 
    0, 
    '这是一个测试用户', 
    '123456', 
    '19900101', 
    NOW(), 
    NULL, 
    NULL, 
    NULL, 
    0, 
    0
),
(
    'U12345678902', 
    '测试用户2', 
    '18888888889', 
    'test2@example.com', 
    'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png', 
    1, 
    '这是第二个测试用户', 
    '123456', 
    '19950601', 
    NOW(), 
    NULL, 
    NULL, 
    NULL, 
    0, 
    0
),
(
    'U12345678903', 
    '管理员', 
    '18888888880', 
    'admin@example.com', 
    'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png', 
    0, 
    '系统管理员账号', 
    '123456', 
    '19880101', 
    NOW(), 
    NULL, 
    NULL, 
    NULL, 
    1, 
    0
); 
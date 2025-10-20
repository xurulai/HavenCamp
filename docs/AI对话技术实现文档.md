# AI 对话功能技术实现文档

## 架构设计

### 整体架构

AI 对话功能在现有 HavenCamp 聊天系统架构上扩展,遵循以下设计原则:

1. **无侵入性**: 不修改现有消息流程,通过复用 Message 模型实现
2. **可扩展性**: 服务层独立,易于切换或扩展其他 AI 服务
3. **配置化**: 所有 AI 相关参数通过配置文件管理
4. **兼容性**: 前端无需改造,AI 回复作为普通文本消息处理

### 技术栈

- **后端框架**: Gin
- **数据库**: MySQL (GORM)
- **缓存**: Redis
- **消息推送**: WebSocket / Kafka
- **AI 服务**: Dify API

## 代码结构

```
HavenCamp/
├── api/v1/
│   └── ai_controller.go              # AI 控制器
├── internal/
│   ├── config/
│   │   └── config.go                 # 配置结构(新增 DifyConfig)
│   ├── dto/
│   │   ├── request/
│   │   │   └── ai_chat_request.go    # AI 请求 DTO
│   │   └── respond/
│   │       └── ai_chat_respond.go    # AI 响应 DTO
│   ├── service/
│   │   └── ai/
│   │       └── dify_service.go       # Dify 服务层
│   └── https_server/
│       └── https_server.go           # 路由注册
└── configs/
    ├── config.toml                   # 本地配置
    └── config.docker.toml            # Docker 配置
```

## 核心实现

### 1. 配置层 (internal/config/config.go)

新增 `DifyConfig` 结构体:

```go
type DifyConfig struct {
    BaseUrl   string        `toml:"baseUrl"`   // Dify API 基础地址
    ApiKey    string        `toml:"apiKey"`    // API 密钥
    AppId     string        `toml:"appId"`     // 应用 ID
    AgentId   string        `toml:"agentId"`   // Agent ID
    Timeout   time.Duration `toml:"timeout"`   // 超时时间
    AiUserId  string        `toml:"aiUserId"`  // AI 虚拟用户 UUID
    AiName    string        `toml:"aiName"`    // AI 显示名称
    AiAvatar  string        `toml:"aiAvatar"`  // AI 头像
}
```

并将其嵌入到 `Config` 结构体中。

### 2. 服务层 (internal/service/ai/dify_service.go)

#### 主要功能

- **Ask()**: 调用 Dify API 获取 AI 回复(阻塞模式)
- **AskStream()**: 预留流式调用接口(暂未实现)

#### Dify API 调用流程

```go
func (s *difyService) Ask(question, userId, sessionId string, meta map[string]interface{}) (string, error) {
    // 1. 检查配置
    // 2. 构造请求体
    difyReq := DifyRequest{
        Inputs:         meta,
        Query:          question,
        ResponseMode:   "blocking",
        ConversationId: sessionId,
        User:           userId,
    }
    
    // 3. 发送 HTTP POST 请求
    // 4. 解析响应
    // 5. 返回 AI 回复文本
}
```

#### 错误处理

- API Key 未配置: 返回 "Dify API 未启用"
- 请求失败: 记录详细日志并返回错误
- 响应解析失败: 返回解析错误

### 3. 控制器层 (api/v1/ai_controller.go)

#### AiChat 函数流程

```
1. 解析请求参数 (AiChatRequest)
   ↓
2. 检查 session_id
   - 为空: 调用 SessionService.OpenSession 创建会话
   - 不为空: 直接使用
   ↓
3. 调用 DifyService.Ask 获取 AI 回复
   ↓
4. 构造 Message 对象并存入数据库
   - Type: Text (0)
   - SendId: aiUserId
   - ReceiveId: owner_id
   - Content: AI 回复
   ↓
5. 通过 WebSocket 推送消息
   - Channel 模式: ChatServer.Clients
   - Kafka 模式: KafkaChatServer.Clients
   ↓
6. 更新 Redis 缓存
   ↓
7. 返回 HTTP 响应 (AiChatRespond)
```

#### 消息持久化与推送

复用现有消息处理逻辑:

```go
// 1. 创建消息对象
aiMessage := model.Message{
    Uuid:       fmt.Sprintf("M%s", random.GetNowAndLenRandomString(11)),
    SessionId:  sessionId,
    Type:       message_type_enum.Text,
    Content:    answer,
    SendId:     aiUserId,
    ReceiveId:  req.OwnerId,
    // ...
}

// 2. 存入数据库
dao.GormDB.Create(&aiMessage)

// 3. 构造响应对象
messageRsp := respond.GetMessageListRespond{
    SendId:    aiMessage.SendId,
    Content:   aiMessage.Content,
    // ...
}

// 4. WebSocket 推送
messageBack := &chat.MessageBack{
    Message: jsonMessage,
    Uuid:    aiMessage.Uuid,
}
receiveClient.SendBack <- messageBack
```

### 4. 路由注册 (internal/https_server/https_server.go)

```go
GE.POST("/ai/chat", v1.AiChat)
```

## 数据流图

```
用户发起请求
    ↓
POST /ai/chat
    ↓
ai_controller.AiChat()
    ├─→ SessionService.OpenSession() (如需创建会话)
    ├─→ DifyService.Ask()
    │       ↓
    │   HTTP POST → Dify API
    │       ↓
    │   返回 AI 回复
    ├─→ dao.GormDB.Create() (存入 message 表)
    ├─→ WebSocket 推送 (ChatServer/KafkaChatServer)
    ├─→ Redis 缓存更新
    └─→ HTTP 响应 (AiChatRespond)
```

## 关键设计决策

### 1. 为什么复用 Message 模型?

**优点**:
- 前端无需改造,现有消息渲染逻辑可直接使用
- 统一的消息查询和管理
- 利用现有的 WebSocket 推送机制

**实现方式**:
- 引入 AI 虚拟用户 (`aiUserId`)
- AI 回复视为从 AI 用户发给真实用户的文本消息

### 2. 为什么选择 HTTP + WebSocket 双通道?

- **HTTP**: 提供同步反馈,前端可立即显示 AI 回复
- **WebSocket**: 保持与现有消息流程一致,支持离线用户后续查看

### 3. 为什么不直接实现流式传输?

- 现有消息系统按"完整消息"为单位处理
- 流式传输需要前端改造
- 预留 `AskStream()` 接口,便于后续扩展

### 4. Redis 缓存策略

复用现有消息缓存逻辑:
- Key: `message_list_{sendId}_{receiveId}`
- 双向缓存: `sendId → receiveId` 和 `receiveId → sendId`
- 过期时间: `REDIS_TIMEOUT` 分钟

## 部署说明

### 1. 本地开发环境

1. 修改 `configs/config.toml`:
   ```toml
   [difyConfig]
   baseUrl = "https://api.dify.ai/v1"
   apiKey = "your-actual-api-key"
   appId = "your-app-id"
   timeout = 30
   aiUserId = "UAI000000000"
   aiName = "AI助手"
   aiAvatar = "https://..."
   ```

2. (可选) 在数据库中插入 AI 虚拟用户:
   ```sql
   INSERT INTO user_info (uuid, nickname, avatar, telephone, password, created_at, is_admin, status)
   VALUES ('UAI000000000', 'AI助手', 'https://...', '00000000000', '', NOW(), 0, 0);
   ```

3. 启动服务:
   ```bash
   cd cmd/haven_camp_server
   go run main.go
   ```

### 2. Docker 环境

1. 修改 `configs/config.docker.toml` 中的 `[difyConfig]`
2. 重新构建镜像:
   ```bash
   docker-compose build backend
   ```
3. 启动服务:
   ```bash
   docker-compose up -d
   ```

## 测试验证

### 1. 功能测试

```bash
# 测试 AI 对话
curl -X POST http://localhost:8000/ai/chat \
  -H "Content-Type: application/json" \
  -d '{
    "owner_id": "U12345678901",
    "question": "你好"
  }'
```

期望响应:
```json
{
  "code": 200,
  "message": "AI对话成功",
  "data": {
    "session_id": "S...",
    "answer": "你好!我是..."
  }
}
```

### 2. 数据库验证

```sql
-- 查看生成的消息
SELECT * FROM message 
WHERE send_id = 'UAI000000000' 
ORDER BY created_at DESC 
LIMIT 1;

-- 查看创建的会话
SELECT * FROM session 
WHERE receive_id = 'UAI000000000' 
ORDER BY created_at DESC 
LIMIT 1;
```

### 3. WebSocket 验证

1. 用户登录并建立 WebSocket 连接
2. 调用 `/ai/chat` 接口
3. 观察 WebSocket 是否收到消息推送

## 性能优化建议

### 1. 并发控制

为 AI 请求添加限流:
```go
// 使用 golang.org/x/time/rate
limiter := rate.NewLimiter(rate.Limit(10), 20) // 每秒10个请求,突发20个

if !limiter.Allow() {
    return "请求过于频繁,请稍后重试", -1
}
```

### 2. 缓存优化

对常见问题的 AI 回复进行缓存:
```go
cacheKey := "ai_answer_" + md5(question)
if cached, err := redis.Get(cacheKey); err == nil {
    return cached, nil
}
// 调用 Dify 后缓存结果
redis.Set(cacheKey, answer, 1*time.Hour)
```

### 3. 异步处理

对于非紧急场景,可将 Dify 调用改为异步:
```go
go func() {
    answer, _ := DifyService.Ask(...)
    // 存库并推送
}()
// 立即返回 "AI 正在思考..."
```

## 扩展性设计

### 1. 支持多个 AI 模型

```go
type AIProvider interface {
    Ask(question, userId, sessionId string, meta map[string]interface{}) (string, error)
}

// 实现不同的 Provider
type DifyProvider struct {}
type OpenAIProvider struct {}
type LocalLLMProvider struct {}
```

### 2. 支持流式传输

实现 `AskStream()` 并配合 SSE (Server-Sent Events):
```go
GE.GET("/ai/chat/stream", v1.AiChatStream)

func AiChatStream(c *gin.Context) {
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    
    answerChan, errChan := ai.DifyService.AskStream(...)
    for {
        select {
        case chunk := <-answerChan:
            fmt.Fprintf(c.Writer, "data: %s\n\n", chunk)
            c.Writer.Flush()
        case err := <-errChan:
            // 处理错误
        }
    }
}
```

### 3. 支持插件系统

为 AI 添加工具调用能力(Function Calling):
```go
type AIPlugin interface {
    Name() string
    Execute(params map[string]interface{}) (string, error)
}

// 注册插件
pluginRegistry.Register("search", SearchPlugin{})
pluginRegistry.Register("weather", WeatherPlugin{})
```

## 故障排查

### 问题 1: AI 不返回回复

**排查步骤**:
1. 检查 Dify API Key 是否正确
2. 查看日志中的 HTTP 请求/响应
3. 检查 Dify 服务是否可访问
4. 验证 AppId 是否正确

### 问题 2: 消息未推送到前端

**排查步骤**:
1. 检查用户是否在线 (WebSocket 连接)
2. 查看 `ChatServer.Clients` 或 `KafkaChatServer.Clients` 中是否有该用户
3. 检查消息是否已存入数据库
4. 查看 WebSocket 错误日志

### 问题 3: 会话创建失败

**排查步骤**:
1. 检查 `owner_id` 是否有效
2. 验证 `aiUserId` 配置是否正确
3. 查看数据库连接状态

## 安全考虑

1. **API Key 保护**: 
   - 不要将 API Key 提交到版本控制
   - 使用环境变量或密钥管理系统

2. **输入验证**:
   - 限制问题长度
   - 过滤敏感词汇
   - 防止注入攻击

3. **访问控制**:
   - 验证用户身份
   - 添加频率限制
   - 记录审计日志

4. **数据隐私**:
   - 不将敏感信息传给 Dify
   - 定期清理历史对话

## 监控与日志

建议添加以下监控指标:

1. **业务指标**:
   - AI 对话请求量
   - 成功率/失败率
   - 平均响应时间

2. **性能指标**:
   - Dify API 调用延迟
   - 数据库写入延迟
   - WebSocket 推送延迟

3. **错误日志**:
   - Dify API 调用失败
   - 数据库操作失败
   - WebSocket 推送失败

## 成本优化

1. **缓存常见问题**: 减少 Dify API 调用次数
2. **批量处理**: 合并相似问题
3. **智能路由**: 简单问题使用规则引擎,复杂问题调用 AI
4. **使用限制**: 设置每用户每日最大请求次数

## 总结

本实现方案通过以下方式实现了 AI 对话功能与现有系统的无缝集成:

1. **复用现有模型**: 利用 Message 和 Session 模型
2. **保持架构一致**: 遵循现有的 Controller → Service → DAO 分层
3. **配置化管理**: 所有 AI 参数可配置
4. **扩展性强**: 易于添加新的 AI 提供商或功能

该方案已在代码中完整实现,可直接部署使用。


package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"haven_camp_server/internal/config"
	"haven_camp_server/pkg/zlog"
	"io"
	"net/http"
	"time"
)

type difyService struct {
	httpClient *http.Client
}

var DifyService = &difyService{
	httpClient: &http.Client{},
}

// DifyRequest 请求结构
type DifyRequest struct {
	Inputs         map[string]interface{} `json:"inputs"`
	Query          string                 `json:"query"`
	ResponseMode   string                 `json:"response_mode"`
	ConversationId string                 `json:"conversation_id,omitempty"`
	User           string                 `json:"user"`
}

// DifyResponse 响应结构
type DifyResponse struct {
	Event          string `json:"event"`
	MessageId      string `json:"message_id"`
	ConversationId string `json:"conversation_id"`
	Answer         string `json:"answer"`
	CreatedAt      int64  `json:"created_at"`
}

// Ask 调用 Dify API 获取回答
func (s *difyService) Ask(question, userId, sessionId string, meta map[string]interface{}) (string, error) {
	conf := config.GetConfig().DifyConfig

	// 检查配置
	if conf.ApiKey == "" || conf.ApiKey == "your-dify-api-key" {
		zlog.Error("Dify API Key 未配置")
		return "", errors.New("Dify API 未启用")
	}

	// 构造请求
	difyReq := DifyRequest{
		Inputs:         meta,
		Query:          question,
		ResponseMode:   "blocking",
		ConversationId: sessionId,
		User:           userId,
	}

	reqBody, err := json.Marshal(difyReq)
	if err != nil {
		zlog.Error("序列化 Dify 请求失败: " + err.Error())
		return "", err
	}

	// 构造 URL
	apiUrl := fmt.Sprintf("%s/chat-messages", conf.BaseUrl)
	if conf.AppId != "" {
		apiUrl = fmt.Sprintf("%s/chat-messages", conf.BaseUrl)
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		zlog.Error("创建 Dify 请求失败: " + err.Error())
		return "", err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+conf.ApiKey)

	// 设置超时
	s.httpClient.Timeout = time.Duration(conf.Timeout) * time.Second

	// 发送请求
	zlog.Info("调用 Dify API: " + apiUrl)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		zlog.Error("调用 Dify API 失败: " + err.Error())
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		zlog.Error("读取 Dify 响应失败: " + err.Error())
		return "", err
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		zlog.Error(fmt.Sprintf("Dify API 返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body)))
		return "", fmt.Errorf("Dify API 返回错误: %d", resp.StatusCode)
	}

	// 解析响应
	var difyResp DifyResponse
	if err := json.Unmarshal(body, &difyResp); err != nil {
		zlog.Error("解析 Dify 响应失败: " + err.Error())
		return "", err
	}

	zlog.Info("Dify API 调用成功，返回答案")
	return difyResp.Answer, nil
}

// AskStream 流式调用 Dify API（预留，暂不实现）
func (s *difyService) AskStream(question, userId, sessionId string, meta map[string]interface{}) (<-chan string, <-chan error) {
	answerChan := make(chan string)
	errorChan := make(chan error)

	go func() {
		defer close(answerChan)
		defer close(errorChan)

		// 暂时使用阻塞模式
		answer, err := s.Ask(question, userId, sessionId, meta)
		if err != nil {
			errorChan <- err
			return
		}
		answerChan <- answer
	}()

	return answerChan, errorChan
}

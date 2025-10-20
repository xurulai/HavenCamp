package request

type AiChatRequest struct {
	OwnerId   string                 `json:"owner_id" binding:"required"`
	SessionId string                 `json:"session_id"`
	Question  string                 `json:"question" binding:"required"`
	Meta      map[string]interface{} `json:"meta"`
}

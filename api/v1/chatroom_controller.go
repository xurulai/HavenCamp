package v1

import (
	"github.com/gin-gonic/gin"
	"haven_camp_server/internal/dto/request"
	"haven_camp_server/internal/service/gorm"
	"haven_camp_server/pkg/constants"
	"haven_camp_server/pkg/zlog"
	"net/http"
)

// GetCurContactListInChatRoom 获取当前聊天室联系人列表
func GetCurContactListInChatRoom(c *gin.Context) {
	var req request.GetCurContactListInChatRoomRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Error(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, rspList, ret := gorm.ChatRoomService.GetCurContactListInChatRoom(req.OwnerId, req.ContactId)
	JsonBack(c, message, ret, rspList)
}

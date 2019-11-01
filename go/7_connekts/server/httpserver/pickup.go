package httpserver

import (
	"connekts/server/grpcserver"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type pickupIn struct {
	MID    string `form:"mid" binding:"required"`
	Pickup int    `form:"pickup" binding:"required"`
}

func pickup(c *gin.Context) {
	var pi pickupIn
	err := c.ShouldBindWith(&pi, binding.FormPost)
	if err != nil {
		c.JSON(400, gin.H{"msg": "参数错误:" + err.Error()})
		return
	}

	err = grpcserver.ChangePickup(pi.MID, pi.Pickup)
	if err != nil {
		c.JSON(500, gin.H{"msg": "Update失败:" + err.Error()})
		return
	}

	//c.JSON(200, gin.H{"msg": "更新pickup成功:"})
	c.Data(200, "text/html", []byte(LIST_HOSTS))
}

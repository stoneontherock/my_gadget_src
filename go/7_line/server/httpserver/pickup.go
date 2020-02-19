package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"line/server/grpcserver"
)

type pickupIn struct {
	MID    string `form:"mid" binding:"required"`
	Pickup int    `form:"pickup" binding:"required"`
}

func pickup(c *gin.Context) {
	var pi pickupIn
	err := c.ShouldBindWith(&pi, binding.Form)
	if err != nil {
		c.JSON(400, gin.H{"msg": "参数错误:" + err.Error()})
		return
	}

	err = grpcserver.ChangePickup(pi.MID, pi.Pickup)
	if err != nil {
		c.JSON(500, gin.H{"msg": "Update失败:" + err.Error()})
		return
	}

	c.Redirect(303, "./list_hosts")
}

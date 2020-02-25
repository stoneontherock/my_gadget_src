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
		respJSAlert(c, 500, err.Error())
		return
	}

	err = grpcserver.ChangePickup(pi.MID, pi.Pickup)
	if err != nil {
		respJSAlert(c, 500, err.Error())
		return
	}

	c.Redirect(303, "./list_hosts")
}

package httpserver

import (
	"connekts/server/panicerr"
	"github.com/gin-gonic/gin"
)

const prefix = "/connekt"

func newEngine() *gin.Engine {
	gin.SetMode(gin.DebugMode) //todo: release
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Any("/", func(c *gin.Context) {})

	c := router.Group(prefix)
	{
		c.POST("/cmd", command)
		c.GET("/list_hosts", listHosts)
		c.POST("/change_pickup", pickup)
		c.POST("/del_host", delHost)
		c.POST("/rpxy", rProxy)
		//c.GET("/list_file", listFile)
		//c.GET("/file_up", fileUpload)
		c.GET("/filesystem", filesystem)
	}

	return router
}

func Serve(addr string) {
	r := newEngine()
	err := r.Run(addr)
	panicerr.Handle(err, "启动http服务失败:")
}

package httpserver

import (
	"connekts/server/panicerr"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net"
	"time"
)

const prefix = "/connekt"

func newEngine() *gin.Engine {
	gin.SetMode(gin.DebugMode) //todo: release
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Any("/", func(c *gin.Context) {})
	router.POST("/login", login)
	router.GET("/login", login)

	router.Use(auth)
	c := router.Group(prefix)
	{
		c.GET("/list_hosts", listHosts)
		c.POST("/del_host", delHost)

		c.POST("/cmd", command)

		c.POST("/change_pickup", pickup)

		c.POST("/rpxy", rProxy)
		c.GET("/list_rproxied", list_rproxied)
		c.POST("/del_rproxied", del_rproxied)

		c.GET("/filesystem", filesystem)
	}

	return router
}

func Serve(addr string) {
	r := newEngine()
	err := r.RunTLS(addr, "server.crt", "server.key")
	panicerr.Handle(err, "启动http服务失败:")
}

func auth(c *gin.Context) {
	login := false
	defer func() {
		if !login {
			c.Redirect(307, "/login")
		}
	}()

	ses, err := c.Cookie("S")
	if err != nil {
		fmt.Printf("cookie获取失败\n")
		return
	}

	name, expire, err := unmarshalCookieValue(ses)
	if err != nil || expire < time.Now().Unix() || name != "管理员" {
		fmt.Printf("过期,用户名不对或错误,err=%v\n", err)
		return
	}

	login = true
	c.Next()
}

func login(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.Header("Content-Type", "	text/html")
		c.String(200, LOGIN_HTML)
		return
	}

	in := struct {
		User string `form:"user" binding:"required"`
		Pstr string `form:"pv" binding:"required"`
	}{}

	err := c.ShouldBindWith(&in, binding.Form)
	if err != nil {
		c.String(401, "login:参数错误"+err.Error())
		return
	}

	if in.User != "管理员" || in.Pstr != "zh@85058" {
		time.Sleep(time.Second * 1)
		c.String(401, "login:用户名或密码错误")
	}

	dm, _, _ := net.SplitHostPort(c.Request.Host)

	c.SetCookie("S", marshalCookieValue("管理员"), SesDur, "/", dm, false, true)
	c.Redirect(303, prefix+"/list_hosts")
}

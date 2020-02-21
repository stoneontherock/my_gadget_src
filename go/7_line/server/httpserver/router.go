package httpserver

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"line/server/panicerr"
	"net"
	"time"
)

const prefix = "/line"

var AdminName = "管理员"
var AdminPv = "zh@85058"

func newEngine() *gin.Engine {
	gin.SetMode(gin.DebugMode) //todo: release
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Any("/", func(c *gin.Context) {})
	router.POST("/line/login", login)
	router.GET("/line/login", login)

	router.Use(auth)
	c := router.Group(prefix)
	{
		c.GET("", listHosts)
		c.GET("/list_hosts", listHosts)
		c.GET("/del_host", delHost)

		c.GET("/cmd", command)

		c.GET("/change_pickup", pickup)

		c.GET("/rpxy", rProxy)
		//c.GET("/list_rproxied", list_rproxied)    //功能冗余
		c.GET("/del_rproxied", del_rproxied)

		c.GET("/filesystem", filesystem)

		c.GET("/logout", func(c *gin.Context) {
			dm, _, _ := net.SplitHostPort(c.Request.Host)
			c.SetCookie("S", "", 0, "/", dm, false, true)
			c.Redirect(307, "/line/login")
		})
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
			c.Redirect(307, "/line/login")
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
		c.Header("Content-Type", "text/html")
		c.String(200, LOGIN_HTML)
		return
	}

	in := struct {
		User string `form:"user" binding:"required"`
		Pstr string `form:"pv" binding:"required"`
	}{}

	err := c.ShouldBindWith(&in, binding.FormPost)
	if err != nil {
		c.String(401, "login:参数错误"+err.Error())
		return
	}

	if in.User != AdminName || in.Pstr != AdminPv {
		time.Sleep(time.Second * 1)
		c.String(401, "login:用户名或密码错误")
	}

	dm, _, _ := net.SplitHostPort(c.Request.Host)

	c.SetCookie("S", marshalCookieValue("管理员"), SesDur, "/", dm, false, true)
	c.Redirect(303, prefix+"/list_hosts")
}
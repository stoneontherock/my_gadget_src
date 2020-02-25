package server

import (
	"fmt"
	"line/server/panicerr"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	BinDir string
	Debug  = "off"

	GRPCListenAddr  = ":65000"
	GRPCPongTimeout = time.Duration(1200)

	HTTPListenAddr = ":65080"
	AdminName      = "管理员"
	AdminPv        = "zh@85058"
)

func init() {
	var err error
	BinDir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	panicerr.Handle(err, "获取可执行文件所在路径的绝对路径失败")

	getEnv(&Debug, "LINE_DEBUG")

	getEnv(&GRPCListenAddr, "LINE_GRPC_LISTEN_ADDR")
	var tmout = "1200"
	getEnv(&tmout, "LINE_GRPC_PONG_TIMEOUT")
	i, _ := strconv.Atoi(tmout)
	if i >= 60 {
		GRPCPongTimeout = time.Duration(i)
	}

	getEnv(&AdminName, "LINE_HTTP_ADMIN")
	getEnv(&AdminPv, "LINE_HTTP_PASSWD")

	getEnv(&HTTPListenAddr, "LINE_HTTP_LISTEN_ADDR")

	fmt.Printf("Debug=%s  GRPCListenAddr=%s  GRPCPongTimeout=%d  HTTPListenAddr=%s  AdminName=%s Pv=%s",
		Debug,
		GRPCListenAddr,
		GRPCPongTimeout,
		HTTPListenAddr,
		AdminName,
		AdminPv)
}

func getEnv(value *string, envKey string) {
	tmp := os.Getenv(envKey)
	fmt.Printf("获取环境变量的值:%s=%s, 对应默认值:%s\n", envKey, tmp, *value)
	if tmp == "" {
		return
	}

	*value = tmp
}

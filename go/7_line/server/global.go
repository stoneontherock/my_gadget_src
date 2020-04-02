package server

import (
	"fmt"
	"line/server/panicerr"
	"os"
	"path/filepath"
	"strconv"
)

var (
	BinDir string

	CheckAliveInterval int64

	Debug = "off"

	GRPCListenAddr = ":65000"
	HTTPListenAddr = ":65080"

	AdminName = "管理员"
	AdminPv   = "zh@85058"
)

func init() {
	var err error
	BinDir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	panicerr.Handle(err, "获取可执行文件所在路径的绝对路径失败")

	tmp := "600"
	getEnv(&tmp, "LINE_CHECK_ALIVE_INTERVAL")
	cai, err := strconv.ParseInt(tmp,10,0)
	if err == nil {
		CheckAliveInterval = cai
	}

	getEnv(&Debug, "LINE_DEBUG")

	getEnv(&GRPCListenAddr, "LINE_GRPC_LISTEN_ADDR")

	getEnv(&AdminName, "LINE_HTTP_ADMIN")
	getEnv(&AdminPv, "LINE_HTTP_PASSWD")

	getEnv(&HTTPListenAddr, "LINE_HTTP_LISTEN_ADDR")

	fmt.Printf("CheckAliveInterval=%d Debug=%s  GRPCListenAddr=%s  HTTPListenAddr=%s  AdminName=%s Pv=%s\n",
		CheckAliveInterval,
		Debug,
		GRPCListenAddr,
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

package model

import (
	"fmt"
	"line/common/panicerr"
	"os"
	"strconv"
)

var (
	CheckAliveInterval int64

	GRPCListenAddr = ":65000"
	HTTPListenAddr = ":65080"

	AdminName = ""
	AdminPv   = ""

	LogLevel = "info"
)

func init() {
	var err error

	tmp := "600"
	getEnv(&tmp, "LINE_CHECK_ALIVE_INTERVAL")
	cai, err := strconv.ParseInt(tmp, 10, 0)
	if err == nil {
		CheckAliveInterval = cai
	}

	getEnv(&LogLevel, "LINE_LOG_LEVEL")
	getEnv(&GRPCListenAddr, "LINE_GRPC_LISTEN_ADDR")
	getEnv(&HTTPListenAddr, "LINE_HTTP_LISTEN_ADDR")

	AdminName = os.Getenv("LINE_HTTP_ADMIN")
	AdminPv = os.Getenv("LINE_HTTP_PASSWD")
	if AdminName == "" || AdminPv == "" {
		panicerr.Handle(fmt.Errorf("环境变量LINE_HTTP_ADMIN或LINE_HTTP_PASSWD没有赋值"))
	}

	fmt.Printf("CheckAliveInterval=%d Debug=%s  GRPCListenAddr=%s  HTTPListenAddr=%s  AdminName=%s Pv=%s\n",
		CheckAliveInterval,
		LogLevel,
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

package httpserver

import (
	"connekts/grpcchannel"
	"connekts/server/model"
	"encoding/binary"
	"errors"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"time"
)

const SesDur = 7 * 24 * 3600

//过期时间:用户名([]uint64,如果用户名超过8字节,则会用冒号分隔各个uint64):用户名%8
func marshalCookieValue(name string) string {
	now := time.Now()
	u64 := uint64(now.Add(time.Duration(7*24*3600)*time.Second).Unix()) - uint64(now.Second()+now.Minute()*60)
	n := strconv.FormatUint(u64*13, 16)

	nameBytes := []byte(name)
	trimRight := 8 - len(nameBytes)%8
	for i := 0; i < trimRight; i++ {
		nameBytes = append(nameBytes, 0)
	}

	for i := 0; i < len(nameBytes); i += 8 {
		u64str := strconv.FormatUint(binary.LittleEndian.Uint64(nameBytes[i:i+8]), 16)
		n = n + ":" + u64str
	}
	return n + ":" + strconv.Itoa(trimRight)
}

func unmarshalCookieValue(value string) (string, int64, error) {
	strs := strings.Split(value, ":")
	if len(strs) < 3 {
		return "", 0, errors.New("cookie长度错误")
	}

	lastField := len(strs) - 1
	trimRight, err := strconv.Atoi(strs[lastField])
	if err != nil {
		return "", 0, errors.New("计算trimRight数失败")
	}

	unixTime, err := strconv.ParseUint(strs[0], 16, 64)
	if err != nil {
		return "", 0, errors.New("cookie解析日期失败1")
	}

	strs = strs[1:lastField]
	buf := make([]byte, len(strs)*8)

	for i := 0; i < len(strs); i++ {
		u64, err := strconv.ParseUint(strs[i], 16, 64)
		if err != nil {
			return "", 0, errors.New("cookie解析日期失败2")
		}
		binary.LittleEndian.PutUint64(buf[i*8:i*8+8], u64)
	}

	return string(buf[:len(buf)-trimRight]), int64(unixTime / 13), nil
}

//label传空字符串表示关闭对应mid的所有连接
func closeConnection(label, mid string) {
	logrus.Debugf("***** label=%s RPxyLisAndConnM=%v", label, model.RPxyConnResM[mid][label])
	for plab, ifaces := range model.RPxyConnResM[mid] {
		if plab != "" && plab != label {
			continue
		}

		for _, ifc := range ifaces {
			switch v := ifc.(type) {
			case *net.TCPConn:
				logrus.Debugf("closeConnection:关闭TCPconn, 内存地址:%p, 远端:%s, 近端:%s", v, v.RemoteAddr(), v.LocalAddr())
				v.Close()
			case *net.TCPListener:
				logrus.Debugf("closeConnection:关闭TCPListener, 内存地址:%p, 监听地址:%s", v, v.Addr())
				v.Close()
			case string:
				go func() {
					time.Sleep(time.Second*10)
					logrus.Debugf("closeConnection:发送关闭连接命令到客户端: addr2=%s ", v)
					model.PongM[mid] <- grpcchannel.Pong{Action: "closeConnections", Data: []byte(v)}
					//logrus.Debugf("closeConnection:发送关闭连接命令到客户端: addr2=%s  [done]", v)
				}()
			default:
				logrus.Errorf("closeConnection:不支持的类型：%v", v)
			}
		}
		delete(model.RPxyConnResM[mid], plab)
		if len(model.RPxyConnResM[mid]) == 0 {
			delete(model.RPxyConnResM, mid)
		}
	}
}

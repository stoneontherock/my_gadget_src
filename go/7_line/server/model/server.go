package model

import (
	"github.com/sirupsen/logrus"
	"line/grpcchannel"
	"net"
)

//
//type PongC struct {
//	RW sync.RWMutex
//	M  map[string]chan grpcchannel.Pong
//}

var PongM = make(map[string]chan grpcchannel.Pong)

var CmdOutM = make(map[string]chan grpcchannel.CmdOutput)

var RPxyConn2M = make(map[string]chan *net.TCPConn)
var RPxyConn1M = make(map[string]chan *net.TCPConn)

var RPxyConnResM = make(map[string]map[string][]interface{}) //key0: mid, key1:label, interface{}对应*net.TCPListener或*net.TCPConn

//var ListFileM = make(map[string]chan *grpcchannel.FileList)
//
//var FileUpDataM = make(map[string]chan *grpcchannel.FileDataUp)

//label传空字符串表示关闭对应mid的所有连接
func CloseConnection(label, mid string) {
	//logrus.Debugf("***** label=%s RPxyLisAndConnM[%s]=%v", label, mid, model.RPxyConnResM[mid])
	for key, ifaces := range RPxyConnResM[mid] {
		if label != "" && label != key {
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
					//time.Sleep(time.Second * 1) //todo 延迟多久？
					logrus.Debugf("closeConnection:发送关闭连接命令到客户端: addr2=%s ", v)
					PongM[mid] <- grpcchannel.Pong{Action: "closeConnections", Data: []byte(v)}
					//logrus.Debugf("closeConnection:发送关闭连接命令到客户端: addr2=%s  [done]", v)
				}()
			default:
				logrus.Errorf("closeConnection:不支持的类型：%v", v)
			}
		}

		delete(RPxyConnResM[mid], key)
		if len(RPxyConnResM[mid]) == 0 {
			delete(RPxyConnResM, mid)
		}
	}
}

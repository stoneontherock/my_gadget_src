package common

import (
	"github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"net"
	"strconv"
	"time"
)

func CopyData(src, dst net.Conn, dir string, serverCloseSocket bool) {
	n, err := io.Copy(dst, src)
	if err != nil {
		logrus.Errorf("copyData:io.Copy:dir=%s, err=%v  n=%d B, src=%p,dst=%p\n", dir, err, n, src, dst)
		err = src.(*net.TCPConn).Close()
		if err != nil {
			logrus.Errorf("copyData: src.Close() dir=%s, err=%v  src=%p,dst=%p\n", dir, err, src, dst)
		}
		err = dst.(*net.TCPConn).Close()
		if err != nil {
			logrus.Errorf("copyData: dst.Close() dir=%s, err=%v  src=%p,dst=%p\n", dir, err, src, dst)
		}
	}

	logrus.Infof("copy数据,dir=%s 成功, %d Bytes, src=%p, dst=%p", dir, n, src, dst)

	err = src.(*net.TCPConn).CloseWrite()
	if err != nil {
		logrus.Errorf("copyData: src.CloseWrite() dir=%s, err=%v  src=%p,dst=%p\n", dir, err, src, dst)
	}
	err = dst.(*net.TCPConn).CloseRead()
	if err != nil {
		logrus.Errorf("copyData: dst.CloseRead() dir=%s, err=%v  src=%p,dst=%p\n", dir, err, src, dst)
	}

	if serverCloseSocket {
		err = src.(*net.TCPConn).Close()
		if err != nil {
			logrus.Errorf("copyData: src.Close() dir=%s, err=%v  src=%p,dst=%p\n", dir, err, src, dst)
		}
		err = dst.(*net.TCPConn).Close()
		if err != nil {
			logrus.Errorf("copyData: dst.Close() dir=%s, err=%v  src=%p,dst=%p\n", dir, err, src, dst)
		}
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func IsPortAvalible(port string) bool {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return false
	}

	lis.Close()
	return true
}

func RandomAvaliblePort() int32 {
	var port int32
	for {
		port = rand.Int31n(1001) + 45535
		if IsPortAvalible(":"+strconv.Itoa(int(port))) {
			break
		}
	}

	return port
}

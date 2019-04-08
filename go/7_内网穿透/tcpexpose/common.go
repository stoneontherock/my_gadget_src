//  A-->B--C--D
//  A端是用户
//  B端是中转服务器
//  C端是内网客户端
//  D端是内网服务
package tcpexpose

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/sirupsen/logrus"
)

var Magic = []byte{'^', '_', '`'} //魔数 ^_`

const (
	HB  = iota //HeartBeat
	NDC        //New Data Channel,新建一个数据通道
)

const (
	CtrlConnTTL = 3
	HeaderLen   = 5
)

type Header struct {
	CMD     byte
	BodyLen int
	Conn    net.Conn
}

func GetHeader(c net.Conn) (*Header, error) {
	var h Header
	buf := make([]byte, HeaderLen)
	_, err := io.ReadFull(c, buf)
	if err != nil {
		return nil, err
	}

	if !EqualMagic(buf) {
		return nil, fmt.Errorf("invalid magic, % x", buf[:3])
	}

	h.CMD = buf[3]
	h.BodyLen = int(buf[4])
	return &h, nil
}

func (h *Header) ReadAll(c net.Conn) ([]byte, error) {
	if h.BodyLen == 0 {
		return nil, nil
	}
	buf := make([]byte, h.BodyLen)
	n, err := io.ReadFull(c, buf)
	if err != nil {
		return nil, err
	}

	if n < h.BodyLen {
		return nil, fmt.Errorf("Header.ReadAll: read body too short %d/%d", n, h.BodyLen)
	}

	return buf, nil
}

func ReadProto(c net.Conn) (*Header, []byte, error) {
	hdr, err := GetHeader(c)
	if err != nil {
		return nil, nil, err
	}

	body, err := hdr.ReadAll(c)
	if err != nil {
		return nil, nil, err
	}

	return hdr, body, err
}

//Todo: Buffer池
func WriteProto(wr net.Conn, body []byte, cmd byte) error {
	hdr := make([]byte, HeaderLen)
	copy(hdr, Magic)
	hdr[3] = cmd
	if len(body) > 255 {
		return errors.New("body to long")
	}
	hdr[4] = byte(len(body))
	_, err := wr.Write(hdr)
	if err != nil {
		return err
	}

	_, err = wr.Write(body)
	if err != nil {
		return err
	}

	return nil
}

const LifeLong = 10

func EqualMagic(b []byte) bool {
	if len(b) < len(Magic) {
		return false
	}

	for i := 0; i < len(Magic); i++ {
		if Magic[i] != b[i] {
			return false
		}
	}

	return true
}

func NewDataChannel(c net.Conn, bPort, dPort int) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint16(buf[:2], uint16(bPort))
	binary.BigEndian.PutUint16(buf[2:], uint16(dPort))

	err := WriteProto(c, buf, NDC)
	return err
}

func NetCopy(src, dst net.Conn, dir string) {
	logrus.Debugf("%s...", dir)
	n, err := io.Copy(dst, src)
	if err != nil {
		logrus.Errorf("%s, %dByte,err:%v", dir, n, err)
		return
	}
	logrus.Debugf("%s..Done, %dBytes", dir, n)
}

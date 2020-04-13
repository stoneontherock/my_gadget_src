package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sirupsen/logrus"
	"gopkg.in/square/go-jose.v2/json"
	"line/common/connection"
	"line/common/connection/pb"
	"line/server/model"
	"net"
	"strconv"
	"strings"
)

type rproxyIn struct {
	Mid        string `form:"mid" binding:"required"` // binding:"hexadecimal"`
	NumOfConn2 int32  `form:"num_of_conn2"`           // binding:"gte=1"`
	Port1      string `form:"port1"`                  //binding:"numeric"`
	Addr3      string `form:"addr3"`                  // binding:"tcp_addr"`
	Label      string `form:"label"`                  // binding:"required"`
}

func rProxy(c *gin.Context) {
	var ri rproxyIn
	err := c.ShouldBindWith(&ri, binding.Form)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	if !isHostPickedUp(ri.Mid) {
		respJSAlert(c, 500, "主机未勾住")
		return
	}

	//如果port1为空，则列出当前mid的rproxy
	if ri.Port1 == "" {
		data := struct {
			Mid    string
			Labels []string
		}{
			Mid: ri.Mid,
		}

		if rp, ok := model.RPxyConnResM[ri.Mid]; ok {
			for label, _ := range rp {
				//logrus.Debugf("**** label:%s", label)
				data.Labels = append(data.Labels, label)
			}
		}

		err = AddRPxyTmpl.Execute(c.Writer, data)
		if err != nil {
			respJSAlert(c, 500, "模板渲染出错"+err.Error())
			return
		}
		return
	}

	port1 := ":" + ri.Port1
	if !connection.IsPortAvalible(port1) {
		respJSAlert(c, 400, "port1不可用:")
		return
	}

	port2 := ":" + strconv.Itoa(int(connection.RandomAvaliblePort()))

	pongC, ok := model.PongM[ri.Mid]
	if !ok {
		respJSAlert(c, 400, "主机不在活动状态,id:"+ri.Mid)
		return
	}

	err = listen2Side(ri.Mid, ri.Label, port1, port2, int(ri.NumOfConn2))
	if err != nil {
		respJSAlert(c, 500, "创建bridge listener 失败:"+err.Error())
		return
	}

	rpr := pb.RPxyResp{Port2: port2, Addr3: ri.Addr3, NumOfConn2: ri.NumOfConn2}
	bs, err := json.Marshal(&rpr)
	if err != nil {
		respJSAlert(c, 500, "序列化到pong data失败:"+err.Error())
		return
	}

	pongC <- pb.Pong{Action: "rpxy", Data: bs}

	c.Redirect(303, "./rpxy?mid="+ri.Mid)
}

func listen2Side(mid, label, port1, port2 string, numOfConn2 int) error {
	conn1Ch := make(chan *net.TCPConn) //1端连接的chan
	conn2Pool := make(chan *net.TCPConn, numOfConn2)

	model.RPxyConn1M[mid] = conn1Ch
	model.RPxyConn2M[mid] = conn2Pool

	taddr1, err := net.ResolveTCPAddr("tcp", port1)
	if err != nil {
		return err
	}

	taddr2, err := net.ResolveTCPAddr("tcp", port2)
	if err != nil {
		return err
	}

	if _, ok := model.RPxyConnResM[mid]; !ok {
		model.RPxyConnResM[mid] = make(map[string][]interface{})
	}

	pLabel := label + port1
	model.RPxyConnResM[mid][pLabel] = append(model.RPxyConnResM[mid][pLabel], taddr2.String())
	go listen(taddr1, conn1Ch, mid, pLabel)
	go listen(taddr2, conn2Pool, mid, pLabel)

	return nil
}

func listen(addr *net.TCPAddr, connCh chan<- *net.TCPConn, mid, label string) {
	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		logrus.Errorf("listen:监听失败:%s. err:%v", addr.String(), err)
		return
	}
	logrus.Debugf("listen:监听成功,%s, listener引用的内存地址:%p", addr, lis)
	model.RPxyConnResM[mid][label] = append(model.RPxyConnResM[mid][label], lis)

	for {
		conn, err := lis.AcceptTCP()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			}
			logrus.Warnf("lis.Accept:addr=%s err:%v", addr, err)
			continue
		}

		logrus.Debugf("listen:连接到来,%s->%s 连接引用的内存地址%p", conn.RemoteAddr(), addr, conn)
		model.RPxyConnResM[mid][label] = append(model.RPxyConnResM[mid][label], conn)
		connCh <- conn
	}
}

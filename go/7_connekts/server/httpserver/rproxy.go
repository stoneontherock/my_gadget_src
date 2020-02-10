package httpserver

import (
	"connekts/common"
	gc "connekts/grpcchannel"
	"connekts/server/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sirupsen/logrus"
	"gopkg.in/square/go-jose.v2/json"
	"net"
	"strconv"
	"strings"
)

type rproxyIn struct {
	MID        string `form:"mid"`          // binding:"hexadecimal"`
	NumOfConn2 int32  `form:"num_of_conn2"` // binding:"gte=1"`
	Port1      string `form:"port1"`        //binding:"numeric"`
	Addr3      string `form:"addr3"`        // binding:"tcp_addr"`
	Label      string `form:"label"`        // binding:"required"`
}

func rProxy(c *gin.Context) {
	var ri rproxyIn
	err := c.ShouldBindWith(&ri, binding.FormPost)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	if ri.Port1 == "" {
		data := struct {
			MID   string
			Labels []string
		}{
			MID: ri.MID,
		}

		if rp, ok := model.RPxyListenerM[ri.MID]; ok {
			for label, _ := range rp {
				logrus.Debugf("**** label:%s",label)
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
	if !common.IsPortAvalible(port1) {
		respJSAlert(c, 400, "port1被占用:")
		return
	}

	port2 := ":" + strconv.Itoa(int(common.RandomAvaliblePort()))

	pongC, ok := model.PongM[ri.MID]
	if !ok {
		respJSAlert(c, 400, "主机不在活动状态,id:"+ri.MID)
		return
	}

	err = listen2Side(ri.MID, ri.Label, port1, port2, int(ri.NumOfConn2))
	if err != nil {
		respJSAlert(c, 500, "创建bridge listener 失败:"+err.Error())
		return
	}

	rpr := gc.RPxyResp{Port2: port2, Addr3: ri.Addr3, NumOfConn2: ri.NumOfConn2}
	data, err := json.Marshal(&rpr)
	if err != nil {
		respJSAlert(c, 500, "序列化到pong data失败:"+err.Error())
		return
	}

	pongC <- gc.Pong{Action: "rpxy", Data: data}

	c.Redirect(303,"./list_rproxied")
}

func listen2Side(mid, label, port1, port2 string, numOfConn2 int) error {
	connC1 := make(chan *net.TCPConn)
	conn2Pool := make(chan *net.TCPConn, numOfConn2)

	model.RPxyConn1M[mid] = connC1
	model.RPxyConn2M[mid] = conn2Pool

	taddr1, err := net.ResolveTCPAddr("tcp", port1)
	if err != nil {
		return err
	}

	taddr2, err := net.ResolveTCPAddr("tcp", port2)
	if err != nil {
		return err
	}

	lisC := make(chan *net.TCPListener)
	go listen(taddr1, connC1, lisC)
	go listen(taddr2, conn2Pool, lisC)

	lis1, ok := <-lisC
	if !ok {
		return fmt.Errorf("监听port1(%s)或port2(%s)失败", port1, port2)
	}

	lis2, ok := <-lisC
	if !ok {
		return fmt.Errorf("监听port1(%s)或port2(%s)失败", port1, port2)
	}

	if _, ok := model.RPxyListenerM[mid]; !ok {
		model.RPxyListenerM[mid] = make(map[string][]*net.TCPListener)
	}
	model.RPxyListenerM[mid][label+"【"+port1+"】"] = []*net.TCPListener{lis1, lis2}
	close(lisC)

	return nil
}

func listen(addr *net.TCPAddr, connC chan<- *net.TCPConn, lisC chan<- *net.TCPListener) {
	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		logrus.Errorf("监听1侧失败:%s. err:%v", addr.String(), err)
		close(lisC)
		return
	}
	lisC <- lis

	for {
		conn, err := lis.AcceptTCP()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			}
			logrus.Warnf("lis.Accept:addr=%s err:%v", addr, err)
			continue
		}

		connC <- conn
	}
}

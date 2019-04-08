package httpserver

import "context"

type aCtx struct {
	Cancel     context.CancelFunc `json:"-"`
	CHostDPort string
}

var BindMap = make(map[string]*aCtx) //key=A的IP地址+A端监听端口号 ， value=C端IP地址+D端监听端口号

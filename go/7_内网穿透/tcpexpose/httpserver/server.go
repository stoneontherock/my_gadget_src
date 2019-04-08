package httpserver

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	ts "tcpexpose/tserver"
)

func HTTPServe(addr string) {
	http.HandleFunc("/", func(http.ResponseWriter, *http.Request) {})
	http.HandleFunc("/nodes", listNodes)
	http.HandleFunc("/bind", bindSideA)          //todo
	http.HandleFunc("/list_bind", listBindSideA) //todo
	err := http.ListenAndServe(addr, nil)
	panic(err)
}

func listNodes(wr http.ResponseWriter, req *http.Request) {
	ec := json.NewEncoder(wr)
	err := ec.Encode(&ts.CtrlMap)
	if err != nil {
		http.Error(wr, err.Error(), 500)
	}
}

func listBindSideA(wr http.ResponseWriter, req *http.Request) {
	ec := json.NewEncoder(wr)
	err := ec.Encode(&BindMap)
	if err != nil {
		http.Error(wr, err.Error(), 500)
	}
}

func bindSideA(wr http.ResponseWriter, req *http.Request) {
	da := req.FormValue("dstAddr")    //C的IP和D端的端口号
	po := req.FormValue("portOffset") //在D端port的基础上加portOffset就得到了A端的监听端口

	cHost, dPort, e1 := net.SplitHostPort(da)
	portOffset, e2 := strconv.Atoi(po)

	if e1 != nil || cHost == "" || e2 != nil {
		http.Error(wr, "invalid value of `dstAddr` or `portOffset`", 400)
		return
	}

	aHost, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		http.Error(wr, "bindSideA:net.SplitHostPort:"+err.Error(), 400)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	dp, _ := strconv.Atoi(dPort)
	go ts.SideAServe(ctx, aHost, cHost, dp, portOffset)

	var actx = aCtx{cancel, da}
	BindMap[aHost+"/"+strconv.Itoa((portOffset+dp)%65536)] = &actx

	wr.Write([]byte("OK\n"))
}

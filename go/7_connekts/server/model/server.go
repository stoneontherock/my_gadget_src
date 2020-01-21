package model

import (
	gc "connekts/grpcchannel"
	"net"
)

//
//type PongC struct {
//	RW sync.RWMutex
//	M  map[string]chan gc.Pong
//}

var PongM = make(map[string]chan gc.Pong)

var CmdOutM = make(map[string]chan gc.CmdOutput)

var RPxyConn2M = make(map[string]chan *net.TCPConn)
var RPxyConn1M = make(map[string]chan *net.TCPConn)

var RPxyListenerM = make(map[string]map[string][]*net.TCPListener)

//var ListFileM = make(map[string]chan *gc.FileList)
//
//var FileUpDataM = make(map[string]chan *gc.FileDataUp)

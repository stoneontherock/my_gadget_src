package model

import (
	"connekts/grpcchannel"
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

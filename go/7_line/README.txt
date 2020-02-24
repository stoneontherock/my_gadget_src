1-客户端的环境变量: 
	LINE_GRPC_SERVER=grpc服务端的地址 
	LINE_REPORT_INTERVAL=上报信息的间隔，单位秒,默认30s
2-服务端的环境变量: 
	LINE_STDOUT_DEBUG     开启DEBUG日志，并打印到stdout,取值on/off 
	LINE_GRPC_LISTEN_ADDR
	LINE_GRPC_PONG_TIMEOUT //响应客户端的上报动作的超时时间，pong一般携带动作，比如命令、反代等
	LINE_HTTP_LISTEN_ADDR
	LINE_HTTP_ADMIN
	LINE_HTTP_PASSWD
3-windows文件浏览服务是按照硬盘顺序来的，第一次是C盘，第二次是D盘...
4-服务端可执行文件所在目录下需要放tls证书(server.crt)和秘钥(server.key)

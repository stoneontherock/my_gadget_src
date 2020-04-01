0-windows下的客户端可以通过nssm来注册服务， nssm install <ServiceName> <binPath>
1-客户端的环境变量: 
	LINE_GRPC_SERVER=grpc服务端的地址 
	LINE_REPORT_INTERVAL=上报信息的间隔，单位秒,默认60s
2-服务端的环境变量: 
	LINE_STDOUT_DEBUG     开启DEBUG日志，并打印到stdout,取值on/off 
	LINE_GRPC_LISTEN_ADDR
	LINE_HTTP_LISTEN_ADDR
	LINE_HTTP_ADMIN
	LINE_HTTP_PASSWD
        LINE_CHECK_ALIVE_INTERVAL  单位秒，服务端会按这个周期检查客户端，如果客户端超过这个这个周期没上报信息，则服务端将此客户端从db中删除,默认600秒
3-windows文件浏览服务是按照硬盘顺序来的，第一次是C盘，第二次是D盘...
4-服务端可执行文件所在目录下需要放tls证书(server.crt)和秘钥(server.key)

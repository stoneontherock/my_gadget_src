#!/bin/bash
# [关于]
#本脚本是客户端懒人脚本， 应用于： https://github.com/{我}/suansuan.git
#记得给该脚本链接到/usr/local/bin
#
# [服务端/客户端安装]
# https://github.com/{我}/suansuan-go
#
# [privoxy]
#需要安装privoxy: apt install privoxy
# 修改/etc/privoxy/config:(搜关键字)
#      listen-address  127.0.0.1:8118        #privoxy的监听地址
#      forward-socks5   /  127.0.0.1:1080 . #指定socks5的服务端地址。 forward-socks5和forward-socks5t是有区别的，前者dns在服务端解析，后者dns在本地端解析
# #脚本中用到的pvox脚本文件的内容是：/bin/systemctl "$1" privoxy 
# 记得给pvox开sudo:  {用户}  ALL=(ALL)  NOPASSWD: /usr/local/bin/pvox



function fn_main(){
    #gnome下开关代理的命令,cinnamon下测试通过
    proxy_switch='gsettings set org.gnome.system.proxy mode'
    app_dir=$(dirname $(readlink -m $0))

    case "$1" in 
    "on")
        [ -n "$(ps -eo cmd |grep ssr_go_client |grep -v grep)"  ] && exit 0
        sudo -u root /usr/local/bin/pvox restart
        (	
        cd $app_dir
	nohup ./ssr_go_client -c config.json >/dev/null 2>&1 &
	)
	
        if [ -z "$(ps -eo cmd |grep ssr_go_client |grep -v grep)"  ] 
        then
            printf "\e[031mfailed\e[0m\n"
            exit 1
        fi
        $proxy_switch  manual
        ;;
    "off")
        sudo -u root /usr/local/bin/pvox stop
        pid=$(ps -eo pid,cmd |awk '/ssr_go_client/ && !/awk/{print $1}')
        [ -n "$pid" ] && kill -15 $pid
        $proxy_switch none
       ;;
    *)
       echo "Usage: $0 <on|off>"
       exit 1
       ;;
    esac
}


if [ $(id -u) -eq 0 ]
then
    su -l zh -c "$(readlink -m $0) $@" 
else
    fn_main "$@"
fi




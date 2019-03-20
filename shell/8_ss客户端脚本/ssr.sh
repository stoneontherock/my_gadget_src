#!/bin/bash
# [关于]
#本脚本是ssr客户端懒人脚本。为了方便命令执行，记得给该脚本链接到/usr/local/bin或/usr/bin之类的PATH路径
#
# [服务端/客户端安装]
# 服务端和客户端源码都在里面 https://github.com/{我}/suansuan-go 
# 本脚本同目录的config.json文件是给用的ssr客户端使用的
#
# [privoxy]
# 安装privoxy的作用是将socks5代理转成http/https代理,方便shell命令红杏出墙
# 安装privoxy: apt install privoxy
# 修改/etc/privoxy/config:(搜关键字)
#      listen-address  127.0.0.1:8118        # privoxy的监听地址,这个地址即是http/https代理的监听地址
#      forward-socks5   /  127.0.0.1:1080 .  # 指定要转发到哪个socks5的服务端, socks5和socks5t区别: 前者dns在服务端解析，后者dns在本地端解析
#
# [privoxy的sudo配置]
# 假设操作系统的普通用户是zh
# /etc/sudoers文件追加一行: zh ALL=(root)  NOPASSWD: /bin/systemctl stop privoxy, NOPASSWD: /bin/systemctl restart privoxy


# [注意]
# 本脚本的普通用户是zh, 请根据自己操作系统实际用户修改
# 本脚本对应的ssr客户端二进制文件名为go_ssr_client, 请根据自己操作系统实际用户修改
COMMON_USER=zh      
CLIENT_BIN=go_ssr_client

function fn_main(){
    #gnome下开关代理的命令,cinnamon下测试通过
    proxy_switch='gsettings set org.gnome.system.proxy mode'
    app_dir=$(dirname $(readlink -m $0))

    case "$1" in 
    "on")
        [ -n "$(ps -eo cmd |grep "${CLIENT_BIN}" |grep -v grep)"  ] && exit 0
        sudo -u root systemctl restart privoxy
        (
           cd $app_dir
           nohup ./"${CLIENT_BIN}" -c config.json >/dev/null 2>&1 &
        )

        if [ -z "$(ps -eo cmd |grep "${CLIENT_BIN}"  |grep -v grep)"  ] 
        then
            printf "\e[031mfailed\e[0m\n"
            exit 1
        fi
        $proxy_switch  manual
        ;;
    "off")
        sudo -u root systemctl stop privoxy
        pid=$(ps -eo pid,cmd |awk '/'"${CLIENT_BIN}"'/ && !/awk/{print $1}')
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
    su -l "$COMMON_USER" -c "$(readlink -m $0) $@" 
else
    fn_main "$@"
fi




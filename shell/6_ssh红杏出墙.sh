#!/bin/bash
function fn_main(){
    case "$1" in 
        "on")
            gsettings set org.gnome.system.proxy mode manual 
            ps=$(ps -eo cmd |grep -v grep |grep 1080)
            if [ -z "$ps" ]
            then
                ssh -oStrictHostKeyChecking=no  -Nf -D 127.0.0.1:1080  -p3389 zh@ze.ddns.net
            fi
            ;;
        "off") 
            gsettings set org.gnome.system.proxy mode none
	    pid=$(ps -eo pid,cmd |awk '/:1080/ && !/awk/{print $1}' |xargs)
	    if [ -n "$pid" ]
	    then
		    kill $pid
	    fi
            ;;
        *) echo "Usage: pxy <on|off>"
    esac
}

if [ $(id -u) -eq 0 ]
then
    su -l zh -c "$(readlink -m $0) $@" 
else
    fn_main "$@"
fi

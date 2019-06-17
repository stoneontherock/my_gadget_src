#!/bin/bash
function fn_do(){
    case "$1" in 
        "on")
	    if [ $# -ne 5 ]
	    then
		    echo "$0 on <host> <port> <user> <str>"
		    return 100
	    fi
            gsettings set org.gnome.system.proxy mode manual 
            ps=$(ps -eo cmd |grep -v grep |grep ":1080")
            if [ -z "$ps" ]
            then
		shift 1
		local ok=false
		for ((i=0;i<10;i++))
		do
		    ssh_tunel "$@" 
		    ps -eo cmd  |grep -v grep|grep -q 1:1080 && { ok=true;break; }
		    echo -n "."
		    sleep 0.2
	        done
		if [ "$ok" == true ]
		then
			echo "connected!"
		fi
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
        *) echo "Usage: $0 <on|off>"
    esac
}

function ssh_tunel(){
	local host="$1"
	local port="$2"
	local user="$3"
	local pstr="$4"

	/usr/bin/expect <<< "set timeout 10
	spawn ssh -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -Nf -D127.0.0.1:1080 -p$port $user@$host
	set ssh_id \$spawn_id
        #exp_internal 1
	expect {
	   -nocase \"yes/no\" { exp_send \"yes\r\"; exp_continue; }
	   \"password: \" { exp_send \"$pstr\r\"; send_user \"send pstr to \$ssh_id\n\" } 
	   timeout { send_error \"ERROR:timeout\"; close \$ssh_id; exit 11; }
        }
	wait \$ssh_id
	" >/dev/null 2>&1
	#return $?
}

function fn_main(){
    if [ $(id -u) -eq 0 ]
    then
        su -l zh -c "$(readlink -m $0) $@" 
    else
        fn_do "$@"
	local ret=$?
	if [ $ret -ne 0 ]
	then
		printf "\e[31mfailed\e[0m\n"
		exit $ret 
	fi
    fi
}

fn_main "$@"

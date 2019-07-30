#!/bin/bash
function fn_do(){
	case "$1" in 
	"on")
		if [ $# -ne 5 ]
		then
			echo "$0 on <host> <port> <user> <pstr>"
			return 100
		fi
		gsettings set org.gnome.system.proxy mode manual 
		ps=$(ps -ef |grep -v grep |grep "1:1080")
		if [ -n "$ps" ]
		then
			echo "$ps"
			return 0
		fi

		shift 1
		for ((i=0;i<10;i++))
		do
			ssh_tunel "$@" 
			if [ $? -eq 33 ] 
			then
				printf "\e[31mincorrect password!\n\e[0m"
				return 33
			fi
			ps -ef |grep -v grep|grep 1:1080 && { return 0; break; }
			echo -n "."
			sleep 0.3
		done
		return 99
		;;
	"off") 
		gsettings set org.gnome.system.proxy mode none
		pid=$(ps -eo pid,cmd |awk '/:1080/ && !/awk/{print $1}' |xargs)
		if [ -n "$pid" ]
		then
			kill $pid
		fi
		;;
	*) echo "Usage: $0 {on <host> <ssh_port> <ssh_user> <ssh_pwd>|off}"
	esac
}

function ssh_tunel(){
	local host="$1"
	local port="$2"
	local user="$3"
	local pstr="$4"

	/usr/bin/expect <<< "set timeout 10
	spawn bash -c \"ssh -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -Nf -D127.0.0.1:1080 -p$port $user@$host; echo CMD_END \"
	#exp_internal 1
	expect {
		-nocase \"yes/no\" { exp_send \"yes\r\"; exp_continue; }
		timeout { send_error \"ERROR:timeout1\"; exit 11; }
		\"password: \" { exp_send {$pstr}; exp_send \"\r\"; send_user \"passwd has been sent\n\" } 
	}
	expect {
		timeout { send_error \"ERROR:timeout2\"; exit 12; }
		\"CMD_END\" { send_user \"SSH_TUNEL_OK!\n\"; exit 0; } 
		\"password: \" { send_error \"password incorrect!\n\"; exit 33; } 
	}
	exit 34
	" >/dev/null 2>&1
	return $?
}

function fn_main(){
	if [ $(id -u) -eq 0 ]
	then
		su -l zh -c "$(readlink -m $0) $*" 
	else
		fn_do "$@" 
		if [ $? -eq 0 ]
		then
			printf "\e[1;32mDONE!\n\e[0m"
		else
			printf "\e[31mfailed\n\e[0m"
		fi
	fi
}

fn_main "$@"
exit $?

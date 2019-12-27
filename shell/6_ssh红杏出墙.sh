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
		for ((i=0;i<2;i++))
		do
			ssh_tunel "$@" 
			if [ $? -eq 30 ] 
			then
				printf "\e[31mincorrect password!\n\e[0m"
				return 30 
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

	echo '
#!/usr/bin/expect
if { $argc < 4 } {
    send_user " ERROR : Invalid arguments.\n"
    send_user " Usage : $argv0 host port user pwd\n"
    exit 10 
}

lassign $argv host port user pstr
set timeout 15

spawn bash -c "ssh -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -fq -NTD127.0.0.1:1080 -p$port $user@$host && echo TUNNEL_OK || echo TUNNEL_FAIL" 
#exp_internal 1
set ask_pw 0
expect {
	-nocase "yes/no" { 
		exp_send "yes\r";
		exp_continue;
	 }

	"TUNNEL_OK" { send_user "TUNNEL_OK!\n"; exit 0; } 
	"TUNNEL_FAIL" { send_error "TUNNEL_FAIL!\n"; exit 20; } 

	"password: " {
		 incr ask_pw
		 if {$ask_pw == 1} {
		 	exp_send "$pstr\r"
			exp_continue
		 }
		 send_error "password incorrect!\n";
		 exit 30;
	 } 

	default { send_user "ERROR: eof or timeout,1\n"; exit 40; } 
}' | /usr/bin/expect -f - "$host" "$port" "$user" "$pstr" >/dev/null 2>&1
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

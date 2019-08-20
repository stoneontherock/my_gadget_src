#!/bin/bash
#
#在/etc/rc.local中加入本脚本路径，或者创建systemd服务开机启动

trap '' SIGHUP

interval=$1 #20s
retry=$2 #3次
if [ $# -ne 2 -o -n "${interval//[0-9]}" -o -n "${retry//[0-9]}" ]
then
	echo "Usage: $0 <sleep interval> <retry times>"
	exit 1
fi


logf=/root/wlan0_addr_change.log
invalid_cnt=0

lines=50
if [ $(wc -l $logf |awk '{print $1}') -gt $lines ]
then
	last=$(tail -n $lines $logf)
	echo "$last" > $logf
fi

addr=$(ip a s dev wlan0 |awk '/inet / && !/169\.254\./{print $2}')
while :
do
	tmp=$(ip a s dev wlan0 |awk '/inet / && !/169\.254\./{print $2}')
	if [ "$tmp" != "$addr" ]
	then
		addr=$tmp
		echo "$(date +%F_%T) IP:${addr} GW:$(ip r s |awk '/via/{print $3}')" >>$logf
	fi

	if [ -z "$tmp" ]
	then
		let invalid_cnt++
	else
		invalid_cnt=0
	fi	

	if [ $invalid_cnt -gt $retry ]
	then
		echo "$(date +%F_%T)  wlan0 has no ip addr, create AP" >>$logf
		systemctl stop dnsmasq
		/usr/bin/create_ap -n wlan0 Raspi zh@85058  >/dev/null 2>&1 &   # 脚本： https://github.com/oblique/create_ap
		sleep 50
		if [ -z "$(ps -ef |fgrep hostapd |grep -v grep)" ]
		then
			echo "$(date +%F_%T) create AP on wlan0 failed, reboot" >>$logf
			/sbin/reboot
		fi
		invalid_cnt=0
	fi

	sleep $interval
done

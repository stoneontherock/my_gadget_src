#!/bin/bash

function rand(){
	ln=$1
	nano=$(date +%s%N)
	r=$((nano % ln))
	if [ $r  -eq 0 ] 
	then
		echo 1
	else
		echo $r
	fi
}

first_song="钢琴魅音24kbps.mp3"

function fn_main(){
	mdir=$(dirname $(readlink -m $0))
	plist="$mdir"/playlist.txt
	[ ! -f "$plist" ] && ls "$mdir" >"$plist" 

	l=$(wc -l "$plist")
	l=${l% *}

	s1=$(sed -n "$(rand $l)p" "$plist")
	s2=$(sed -n "$(rand $l)p" "$plist")
	s3=$(sed -n "$(rand $l)p" "$plist")
	s4=$(sed -n "$(rand $l)p" "$plist")
	s5=$(sed -n "$(rand $l)p" "$plist")
	s6=$(sed -n "$(rand $l)p" "$plist")
	s7=$(sed -n "$(rand $l)p" "$plist")
	s8=$(sed -n "$(rand $l)p" "$plist")
	echo -e "* $s1\n* $s2\n* $s3\n* $s4\n* $s5\n"

	cd "$mdir" >/dev/null 
	amixer cset numid=1,iface=MIXER,name='PCM Playback Volume' 10%  
	mpg123 $first_song "$s1" "$s2" "$s3" "$s4" "$s5" "$s6" "$s7" "$s8" >/dev/null 2>&1 &
	pid=$!

	cd - >/dev/null
        # volume up
	for i in {80..100}  #要根据音响箱大小来确定范围。
	do
		#这里的numid和name等参数要根据命令"amixer contents"确定(找到volume那项) 	
		amixer cset numid=1,iface=MIXER,name='PCM Playback Volume' ${i}%  
		sleep 10
	done

	sleep 900
	kill $pid
	amixer cset numid=1,iface=MIXER,name='PCM Playback Volume' 10%  
	/sbin/poweroff
}


fn_main

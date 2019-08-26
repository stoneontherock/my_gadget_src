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
	echo -e "* $s1\n* $s2\n* $s3\n* $s4\n* $s5\n"

	cd "$mdir" >/dev/null 
	mpg123 "$s1" "$s2" "$s3" "$s4" "$s5">/dev/null 2>&1 &
	pid=$!

	cd - >/dev/null
	sleep 1140 #19分钟
	kill $pid
	/sbin/poweroff
}


fn_main

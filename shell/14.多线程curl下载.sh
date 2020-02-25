#!/bin/bash

function fn_down(){
	index=$(awk -v ind="$1" 'BEGIN{printf("%03d",ind)}')
	range="$2"
	fname="$3"
	url="$4"
	
	echo "curl -L -o tmp_${fname}_$index -r $range $url"
	ua="Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.97 Safari/537.36"
	curl -A "$ua" -L -o "tmp_${fname}_$index" -r $range "$url" >/dev/null 2>&1
}

function fn_main(){
	if [ -z "$(which curl)" ]
	then
		echo "找不到curl"	
		return 1
	fi
	
	url="$1"
	if [ -z "$url" ]
	then
		echo "Usage:$0 <url>"	
		return 2
	fi
	
	size=($(curl -I "$url" |awk -F'[: ]+' '/Content-Length:/{printf("%d %.2f",$2,$2/1024/1024)}'))
	
	fname=$(basename $url)
	if [ "$fname" == "/" -o "$fname" == "/" -o -z "$fname" ] 
	then
		echo "获取文件名失败"
		return 2	
	fi	

	read -p "大小=${size[0]} B(${size[1]} MB), 要分成几份下载?: " num
	
	part=$((${size[0]}/num))
	last=${size[0]} #初始化为总大小
	
        [ -d "$fname" ] && { rm -fv "$fname"/*; } || { mkdir "$fname" || { echo "创建目录\"$fname\"失败"; return 3; }; }

	for ((i=0;i<num;i++))
	do
		fn_down $i "$((part*i))-$(((part*(i+1))-1))" "$fname" "$url" &
		pids="${pids} $!"
		last=$((last-part))
		sleep 0.1
	done
	
	if [ $last -gt 0 ]
	then
		fn_down $i  "$((${size[0]}-$last))-$((${size[0]}-1))" "$fname" "$url" &
		pids="${pids} $!"
	fi
	wait $pids
	if [ $? -ne 0 ]
	then
		echo "部分失败"
		return 4
	fi

	cat "tmp_$fname"_* >"$fname"
	rm "tmp_$fname"_*
}

fn_main "$1"
exit $?

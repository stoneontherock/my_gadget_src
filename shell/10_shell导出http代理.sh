#!/bin/bash

pxySrv=127.0.0.1:8118

function shpxy (){
	case "$1" in 
	   "on") export http_proxy=$pxySrv https_proxy=$pxySrv ;;
	   "off") unset http_proxy https_proxy ;;
	   "*") echo "Usage: funcname <on|off>";;
	esac
}

export -f shpxy

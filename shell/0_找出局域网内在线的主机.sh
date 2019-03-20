#!/bin/bash
# Description: 找出局域网中的在线主机
# Author: Zhou Hui
# Release: 2014-12-02
# Usage:$0 <IP WITH RANGE>"
# Example: $0 192.168.1.100-254"

neigh_file="/tmp/find_neigh.txt"
[ -f "${neigh_file}" ] && rm -f "${neigh_file}" &>/dev/null

function fn_check_conn()
{
    local ip_list=$1
    local range=${ip_list##*.}
    local space=$[${#ip_list}-${#range}]

    for r in `seq ${range%-*} ${range#*-}`
    do
       {
           ping -n -c2 -W2 ${ip_list%.*}.${r} &>/dev/null 
           if [ $? -eq 0 ]
           then
               printf "%${space}s%s up\n" '' "${r}" >>${neigh_file}
           else
               printf "%${space}s%s down\n" '' "${r}" >>${neigh_file}
           fi
       } &
    done
    wait
}

function fn_main()
{
   echo "$@" |egrep '^([0-9]{1,3}\.){3}[0-9]{1,3}-[0-9]{1,3}$' &>/dev/null
   if [ $? -ne 0 ]
   then
       echo "usage:$0 <IP WITH RANGE>"
       echo "example: $0 192.168.1.100-254"
       return 1
   fi

   fn_check_conn $@ 
   
   echo -e "$@"
   echo -e "`cat "${neigh_file}" 2>/dev/null |sort -k1 -n |sed -re 's:(up):\\\e[1;32m\1\\\e[0m:' -e 's:([0-9]+):\\\e[1;34m\1\\\e[0m:'`"
}

fn_main $@

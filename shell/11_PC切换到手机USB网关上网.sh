#!/bin/bash

USB_INTERFACE=""
USB_GW="192.168.42.129"

function fn_wait_usb_iface(){
    while :
    do
        USB_INTERFACE=$(ip l  |awk -F'[ :]+' '/enp[0-9]+s[0-9]+u[0-9]+/{print $2}')
        if [ -n "$USB_INTERFACE" ]
        then
            echo -e "usb net card pluged in"
            break
        fi
        sleep 1
        echo -n "."
    done
}

function fn_set_gw_to_usb(){
    fn_wait_usb_iface
    ip route del default via ${USB_GW} dev ${USB_INTERFACE} >/dev/null 2>&1
    ip route add default via ${USB_GW} dev ${USB_INTERFACE} proto static metric 10
    ip route |grep 'default via'
    sed -i 's/^/#/' /etc/resolv.conf
    echo "nameserver $USB_GW" >>/etc/resolv.conf
}

function fn_del_usb_gw(){
    USB_INTERFACE=$(ip route |awk '/default via.*enp[0-9]+s[0-9]+u[0-9]+/{print $5}')
    if [ -n "$USB_INTERFACE" ]
    then
        ip route del default via ${USB_GW} dev ${USB_INTERFACE}
        sed -i 's/^#//;/nameserver '$USB_GW'/d' /etc/resolv.conf
        echo "usb gw has been deleted."
    fi
}

function fn_main(){
    local action="$1"
    case "$action" in
        "on") fn_set_gw_to_usb ; return $? ;;
        "off") fn_del_usb_gw; return $? ;;
        *) echo "Usage: $0 <on|off>"; return 1;;
    esac
}

fn_main "$@"


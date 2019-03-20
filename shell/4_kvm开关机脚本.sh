#!/bin/bash
# gitzh @ 2019-01-23

function fn_vm_off(){
	virsh list --state-running |awk 'BEGIN{ind=1} NR>2 && $2!=""{print ind" - "$2; ind++}' 
	echo -n "select vm domain(split by space):"
	read ch
	onlist=($(virsh list --state-running |awk 'NR>2{print $2}' |xargs))
	
	echo ""
	for c in $ch
	do
		virsh shutdown --domain "${onlist[$[c-1]]}" --mode acpi
	done
}


function fn_vm_on(){
	virsh list --state-shutoff |awk 'BEGIN{ind=1} NR>2 && $2!=""{print ind" - "$2; ind++}' 
	echo -n "select vm domain(split by space):"
	read ch
	offlist=($(virsh list --state-shutoff |awk 'NR>2{print $2}' |xargs))
	
	echo ""
	for c in $ch
	do
		virsh start --domain "${offlist[$[c-1]]}" 
	done
}

function fn_main(){
	action="$1"
	case "$action" in 
		"on") fn_vm_on ;;
		"off") fn_vm_off ;;
		"list") virsh list --all ;;
		*) echo "Usage: vm <on|off|list>" ;;
	esac
}

fn_main "$@"
exit $?

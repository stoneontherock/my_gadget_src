#!/bin/bash
# Release 2018-08
USAGE="conf.sh <file path> [seprator]"
echo "$USAGE"
echo "Type CTRL+D or read the EOF of file for finish input"
echo "-------------------------"
echo "Get conf text from STDIN:"
if [ $# -gt 2 -o $# -eq 0 ]
then
	echo "$USAGE"
	exit 1
fi

# Global Variable
CONF="$1"
SEP="${2}" ; [ -z "$SEP" ] && SEP=' '
declare -a INPUT_KEY_VALUE


function fn_main() {
	if [ ! -f "$CONF" ]
	then
		echo "conf file($CONF) does not exist, Usage: $0 <conf_file>"
		return 10
	fi
	cp $CONF $CONF.bak.$(date +%F_%H%M%S)
	
	get_key_value_from_stdin
	return $?
}

# 修改/新增配置文件的key-value
function uniq_key_value() {
	if [ $# -ne 2 ]
	then
		echo "number of Arguments must 2"
		return 20
	fi
	local key="$1"
	local value="$2"

	
	local kv
	[ -z "$SEP" ] && local reg_sep='[[:space:]]' || reg_sep="$SEP"
	kv=$(egrep "^[[:blank:]]*$key$reg_sep+" $CONF)
	if [ $? -eq 0 ]             # *** Key已经存在了 ***
	then
		echo "Current is: $kv"
		echo "$value" |grep -q '/'
		if [ ${PIPESTATUS[1]} -eq 0 ]        # *** sed特殊处理/  ***
		then
			sed -ri "s:^[[:blank:]]*$key$reg_sep.*:$key$SEP$value:" $CONF
		else
			sed -ri "s/^[[:blank:]]*$key$reg_sep.*/$key$SEP$value/" $CONF
		fi
        else
		echo "Add new key-value"
		echo "$key$SEP$value" >> $CONF
        fi

	return 0
}


# 从STDIN 获取key-value对
function get_key_value_from_stdin() {
	local key value
	while read line
        do
		[ -z "$line" ] && continue
		key=$(echo "${line}" |sed "s/[$SEP].*//")
		value=$(echo "${line}" |sed -r "s/^$key[$SEP]+//")
		uniq_key_value "$key" "$value"
		if [ $? -ne 0 ]
		then
		    echo "ERROR: replace or add key($key) value($value) failed."
		    return 30
		fi
        done < /dev/stdin

	return 0 
}

fn_main "$@"
exit $?


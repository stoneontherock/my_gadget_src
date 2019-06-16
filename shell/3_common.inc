export PS4='+{${BASH_SOURCE[0]##*/}:${FUNCNAME[0]}}:$LINENO :'
SSH='ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no'


# DEBUG CODE at LAST LINE
# deps environment variable "SH_DBG=ON|OFF", if need put debug output to file, you should set environment "DBG_FILE",
if [ "$SH_DBG" == 'ON' ] 
then
	set -x
else
	set +x
fi

function Err() {
	printf "\033[31mE:[$(date +%F_%T)][${BASH_SOURCE[1]##*/}:$BASH_LINENO] $1\033[0m\n" >&2
}

function Warn() {
	printf "\033[33mW:[$(date +%F_%T)] $1\033[0m\n"
}

function Info() {
	printf "[$(date +%F_%T)] $1\n"
}


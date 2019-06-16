
function mcmd(){
	cmd="$1"
	shift
	for h in $*
	do
		ssh root@$h "$cmd"
	done
}

function mscp(){
	file="$1"
	dst="$2"
	shift; shift
	for h in $*
	do
		scp $file root@$h:"$dst" 
	done
}

export -f mcmd
export -f mscp

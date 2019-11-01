package log

import "fmt"

var silent bool

func Infof(fmtStr string, args ...interface{}) {
	if silent {
		return
	}
	fmt.Printf(fmtStr, args...)
}

func Errorf(fmtStr string, args ...interface{}) {
	if silent {
		return
	}
	fmt.Printf("[Error]:"+fmtStr, args...)
}

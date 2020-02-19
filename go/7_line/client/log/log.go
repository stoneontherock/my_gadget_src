package log

import (
	"fmt"
	"time"
)

var silent bool

func Infof(fmtStr string, args ...interface{}) {
	if silent {
		return
	}
	fmt.Printf(logTime()+fmtStr, args...)
}

func Errorf(fmtStr string, args ...interface{}) {
	if silent {
		return
	}
	fmt.Printf(logTime()+"[Error]:"+fmtStr, args...)
}

func logTime() string {
	return time.Now().Format("[15:04:05.0000] ")
}

package panicerr

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

func Handle(err error, strs ...string) {
	if err != nil {
		fmt.Printf("%s:%v\n", strings.Join(strs, "# "), err)
		logrus.Panicf("%s:%v", strings.Join(strs, "# "), err)
	}
}

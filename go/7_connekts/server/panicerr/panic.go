package panicerr

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

func Handle(err error, strs ...string) {
	if err != nil {
		logrus.Errorf("%s:%v", strings.Join(strs, "# "), err)
		panic(fmt.Sprintf("%s:%v\n", strings.Join(strs, "# "), err))
	}
}

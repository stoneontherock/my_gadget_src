package cc

import (
	"cc/core/module/anti_cc/basic_rule"
	"github.com/sirupsen/logrus"
	"net/http"
)

type PathRule struct {
	basic_rule.Rule
	CCconf []interface{}
}

func NewPathRule(rule basic_rule.Rule, conf []interface{}) *PathRule {
	var pr PathRule
	pr.Rule = rule
	if conf == nil {
		pr.CCconf = make([]interface{}, 0)
	} else {
		pr.CCconf = conf
	}

	return &pr
}

func (pr *PathRule) antiCC(wr http.ResponseWriter, req *http.Request) bool {
	for i := 0; i < len(pr.CCconf); i++ {
		switch c := pr.CCconf[i].(type) {
		case *reqFreqLimitDB:
			logrus.Debugf("进入pr 限速")
			if !reqLimit(wr, req, c, req.Host+req.URL.Path) {
				return false
			}
		case *Redirect_conf:
			logrus.Debugf("进入pr 重定向%d, %s", c.Code, c.URL)
			http.Redirect(wr, req, c.URL, c.Code)
			return false
		case *JS_conf:
			logrus.Debugf("进入pr JS")
			if !c.handle1stReqJS(wr, req) {
				return false
			}
		default:
		}
	}

	return true
}

package cc

import (
	"cc/core/module/anti_cc/basic_rule"
	"github.com/sirupsen/logrus"
	"net/http"
)

type DomainSelector struct {
	basic_rule.Rule
	CCconf   []interface{}
	PathList []PathRule
}

func NewDomainSelector(dmRule basic_rule.Rule, conf []interface{}, pathlist []PathRule) *DomainSelector {
	var dr DomainSelector
	dr.Rule = dmRule
	if conf == nil {
		dr.CCconf = make([]interface{}, 0)
	} else {
		dr.CCconf = conf
	}

	if pathlist == nil {
		dr.PathList = make([]PathRule, 0)
	} else {
		dr.PathList = pathlist
	}

	return &dr
}

func (dr *DomainSelector) antiCC(wr http.ResponseWriter, req *http.Request) bool {
	//logrus.Debugf("进入domain AntiCC,len(ds.conf)=%d", len(dr.CCconf))

	for i := 0; i < len(dr.CCconf); i++ {
		switch c := dr.CCconf[i].(type) {
		case *reqFreqLimitDB:
			logrus.Debugf("进入dr 限速")
			if !reqLimit(wr, req, c, req.Host) {
				return false
			}
		case *Redirect_conf:
			logrus.Debugf("进入dr 重定向%d,%s", c.Code, c.URL)
			http.Redirect(wr, req, c.URL, c.Code)
			return false
		case *JS_conf:
			logrus.Debugf("进入dr js")
			if !c.handle1stReqJS(wr, req) {
				return false
			}
		default:
			logrus.Error("unknown DomainSelector CCconf type")
		}
	}

	return true
}

package cc

import (
	"errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

type DomainGroup struct {
	DomainList []*DomainSelector
	CCconf     *reqFreqLimitDB
}

func NewDomainGroup(ds []*DomainSelector) *DomainGroup {
	var dg DomainGroup
	if ds == nil {
		dg.DomainList = []*DomainSelector{}
	} else {
		dg.DomainList = ds
	}

	return &dg
}

func (dg *DomainGroup) HandleHttpRequest(wr http.ResponseWriter, req *http.Request) error {
	logrus.Debugf("**** Host:%s Path=%s ReqURI:%s****", req.Host, req.URL.Path, req.RequestURI)

	runLifeCheck.Do(allLifeCheck)
	if !l1BlacklistCheck(wr, req) {
		return errors.New("L1Blacklist blocked")
	}

	var dr *DomainSelector
	for i := 0; i < len(dg.DomainList); i++ {
		if dg.DomainList[i].Rule.Match(req.Host) {
			dr = dg.DomainList[i]
			break
		}
	}

	//域名匹配到了
	var pr *PathRule
	if dr != nil {
		//logrus.Debugf("dr非nil, 继续")
		if ok := dr.antiCC(wr, req); !ok {
			return errors.New("CC domain rule has handled the request, REGEX:" + dr.Regexp.String())
		}

		for i := 0; i < len(dr.PathList); i++ {
			if dr.PathList[i].Rule.Match(req.URL.Path) {
				pr = &dr.PathList[i]
				break
			}
		}

	} else {
		if dg.CCconf != nil {
			if !reqLimit(wr, req, dg.CCconf, "") {
				return errors.New("default CC domain rule has handled the request")
			}
		}
	}

	if pr != nil {
		if ok := pr.antiCC(wr, req); !ok {
			return errors.New("CC path rule has handled the request, REGEX:" + pr.Regexp.String())
		}
	}

	logrus.Debug("直接转发")
	return nil
}

var runLifeCheck sync.Once

func allLifeCheck() {
	go jsFailMapLifeCheck()
	go l1BlacklistLifeCheck()
	go l2BlacklistLifeCheck()
}

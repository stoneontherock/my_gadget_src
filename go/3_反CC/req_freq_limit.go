package cc

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"

	"cc/core/module/anti_cc/circle_link_node"
	"cc/core/module/anti_cc/common"
)

//请求频率数据结构：
type Req_freq_conf struct {
	AmongTime   time.Duration `yaml:"among_time"`
	Threshold   int           `yaml:"threshold"`
	L1BlistTerm time.Duration `yaml:"prison1_term"`
	CaptchaTTL  int           `yaml:"captcha_fail"`
	L2BlistTerm time.Duration `yaml:"prison2_term"`
}

type reqFreqLimitDB struct {
	Req_freq_conf
	data map[string]*circle_link_node.Circle //key：hash(ip+域名+路径) ,value: 结构体指针，包含环形链表
	sync.RWMutex
}

func NewReqFreqLimitConf(among time.Duration, threshold int, L1BlistTerm time.Duration, captchaTTL int, L2BlistTerm time.Duration) *reqFreqLimitDB {
	var rfl reqFreqLimitDB
	rfl.AmongTime = among
	rfl.Threshold = threshold
	rfl.L1BlistTerm = L1BlistTerm
	rfl.CaptchaTTL = captchaTTL
	rfl.L2BlistTerm = L2BlistTerm
	rfl.data = make(map[string]*circle_link_node.Circle, 500)

	return &rfl
}

func (rfl *reqFreqLimitDB) delRFLimitor(key string) {
	//logrus.Debugf("delRFLimitor() 删除限速,key;%s", key)
	rfl.Lock()
	delete(rfl.data, key)
	logrus.Debugf("delRFLimitor() 删除限速,key=%s，[Done]", key)
	rfl.Unlock()
}

func reqLimit(wr http.ResponseWriter, req *http.Request, rfl *reqFreqLimitDB, domainPath string) bool {
	cip := common.SplitHostPort(req.RemoteAddr, 0)
	k := common.Md5sum(cip + domainPath)

	rfl.Lock()
	c := rfl.data[k]
	if c == nil {
		logrus.Debugf("ReqLimit() 新增环")
		rfl.data[k] = circle_link_node.NewCircle(int64(rfl.AmongTime)) //这里不能赋值给c，只能赋值给map[k]
		rfl.Unlock()                                                   //重要！解锁
		return true
	}

	fail := c.UpdateCircle()
	rfl.Unlock() //重要！解锁

	//req太频繁,加一级黑名单
	if fail >= rfl.Threshold {
		logrus.Debug("ReqLimit() 发验证码" + cip)
		rfl.delRFLimitor(k)
		addL1Blacklist(cip, rfl.L1BlistTerm, rfl.CaptchaTTL, rfl.L2BlistTerm)
		sendCaptcha(wr, req)

		return false
	}

	return true
}

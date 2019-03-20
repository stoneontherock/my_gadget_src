package cc

import (
	"github.com/mojocn/base64Captcha"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"

	"cc/core/module/anti_cc/common"
)

//软件黑名单： 多个策略共享一个黑名单，触犯任意一条策略都会导致该IP请求受特殊照顾
//软件黑名单唯一性(key)：客户端ip地址
//软件黑名单value：生命周期、未释放前验证失败次数
//释放方式：到期释放/验证码释放
type l1BlacklistUnit struct {
	idKey      string        //captcha
	term       time.Duration //待在一级黑名单多久？(刑满释放)
	captchaTTL int           //刑期内如果TTL值变成了0，就会进入iptables黑名单
	l2Term     time.Duration
}
type l1BlacklistMap struct {
	data map[string]*l1BlacklistUnit
	sync.RWMutex
}

var httpBlacklist = l1BlacklistMap{data: make(map[string]*l1BlacklistUnit, 500)}

var L1BlacklistChkIntv time.Duration
var runL1LifeCheck sync.Once

func addL1Blacklist(ip string, term time.Duration, ttl int, l2term time.Duration) {
	//1级刑期检查
	//runL1LifeCheck.Do(l1BlacklistLifeCheck)

	httpBlacklist.Lock()
	defer httpBlacklist.Unlock()

	if _, ok := httpBlacklist.data[ip]; !ok {
		var u l1BlacklistUnit
		logrus.Debugf("addL1Blacklist() 1级黑名单nil,新增%s", ip)
		u.term = time.Duration(time.Now().UnixNano()) + term
		u.captchaTTL = ttl
		u.l2Term = l2term
		httpBlacklist.data[ip] = &u
		return
	}

	httpBlacklist.data[ip].term = time.Duration(time.Now().UnixNano()) + term
	logrus.Debugf("addL1Blacklist() 刑期更新,%s", ip)
}

//返回值，是否需要下一层的处理器处理
func l1BlacklistCheck(wr http.ResponseWriter, req *http.Request) bool {
	cip := common.SplitHostPort(req.RemoteAddr, 0)

	vcode := req.FormValue("unknown_code")
	//logrus.Debugf("Post的验证码：%s", vcode)

	idkey, ttl, l2term, ok := loadL1Blacklist(cip)
	if !ok {
		logrus.Debugf("l1BlacklistCheck() cip %s不在1级黑名单中", cip)
		return true
	}

	if idkey == "" {
		logrus.Debugf("l1BlacklistCheck() idkey为空，发验证码")
		sendCaptcha(wr, req)
		return false
	}

	//验证码不一致
	if !base64Captcha.VerifyCaptcha(idkey, vcode) {
		if ttl == 1 { //如果当前是1，l1BlacklistTTLMinus1()后ttl就会变成0
			//加IPtables
			logrus.Warnf("l1BlacklistCheck() TTL==0,加2级黑名单：%s", vcode)
			delL1Blacklist(cip)
			go addL2Blacklist(cip, l2term)
			http.Error(wr, "SeverError: L2 blocked", 503)
			return false
		}
		l1BlacklistTTLMinus1(cip)
		logrus.Debugf("l1BlacklistCheck() 验证码验证失败：%s， 再发一次验证码", vcode)
		sendCaptcha(wr, req)
		return false
	}

	//验证码一致,把ip踢出黑名单
	logrus.Debugf("l1BlackListCheck() 验证码验证OK：%s, 踢出1级黑名单", vcode)
	delL1Blacklist(cip)
	//会有读消耗,导致后面处理器中req的content-length和实际不一致
	fmt.Fprintf(wr, nodelayJS, req.RequestURI)
	return false
}

func l1BlacklistLifeCheck() {
	tk := time.NewTicker(L1BlacklistChkIntv)
	for {
		select {
		case <-tk.C:
			//释放掉过期的软件黑名单
			httpBlacklist.Lock()
			for k, _ := range httpBlacklist.data {
				if time.Duration(time.Now().UnixNano()) > httpBlacklist.data[k].term {
					delete(httpBlacklist.data, k)
					logrus.Debugf("l1BlacklistLifeCheck() 刑满释放，key=%s", k)
				}
			}
			httpBlacklist.Unlock()
		}
	}
}

//TTL减1
func l1BlacklistTTLMinus1(ip string) {
	httpBlacklist.Lock()
	defer httpBlacklist.Unlock()
	b, ok := httpBlacklist.data[ip]
	if ok {
		b.captchaTTL--
	}
}

func delL1Blacklist(ip string) {
	//logrus.Debugf("踢出黑名单：%s ...", ip)
	httpBlacklist.Lock()
	delete(httpBlacklist.data, ip)
	httpBlacklist.Unlock()
	logrus.Debugf("踢出1级黑名单：%s [Done]", ip)
}

func loadL1Blacklist(ip string) (idkey string, ttl int, l2term time.Duration, ok bool) {
	httpBlacklist.RLock()
	u, ok := httpBlacklist.data[ip]
	if ok {
		idkey = u.idKey
		ttl = u.captchaTTL
		l2term = u.l2Term
	}
	httpBlacklist.RUnlock()

	return
}

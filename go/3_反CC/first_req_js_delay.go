package cc

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"

	"cc/core/module/anti_cc/common"
)

type JS_conf struct {
	AmongTime   time.Duration
	Threshold   int
	L1BlistTerm time.Duration
	CaptchaTTL  int
	L2BlistTerm time.Duration
}

var (
	CookieSalt string
	FRJsDelay  time.Duration
	FRJsHTML   string
)

const nodelayJS = `<html><script language='javascript' type='text/javascript'> setTimeout("javascript:location.href='%s'", 0); </script></html>`

func NewFirstReqJsConf(amongTime time.Duration, threshold int, L1BlistTerm time.Duration, captchaTTL int, L2BlistTerm time.Duration) *JS_conf {
	var jsc JS_conf
	jsc.AmongTime = amongTime
	jsc.Threshold = threshold
	jsc.L1BlistTerm = L1BlistTerm
	jsc.CaptchaTTL = captchaTTL
	jsc.L2BlistTerm = L2BlistTerm
	return &jsc
}

func (jsc *JS_conf) handle1stReqJS(wr http.ResponseWriter, req *http.Request) bool {
	cip := common.SplitHostPort(req.RemoteAddr, 0)
	cipUA := common.Md5sum(cip + req.UserAgent() + CookieSalt)

	ck, _ := req.Cookie("FRJS")
	logrus.Debugf("cookie:FRJS %p", ck)

	//有cookie
	if ck != nil {
		if ck.Value == cipUA {
			return true
		} else {
			// 有cookie,但是cookie值与hash(cip+UA+Salt)不一致，则说明
			// cookie被冒用(cookie从一个客户端拷贝到了另一个客户端)
			// 此种情况直接加1级黑名单
			logrus.Errorf("Cookie被冒用了,%s", ck.Value)

			addL1Blacklist(cip, jsc.L1BlistTerm, jsc.CaptchaTTL, jsc.L2BlistTerm)
			logrus.Debugf("添加黑名单，JS fail MAP删除%s(%s)", cipUA, cip)
			jsFailDB.Lock()
			delete(jsFailDB.data, cipUA)
			jsFailDB.Unlock()
			fmt.Fprintf(wr, nodelayJS, req.RequestURI)
			return false
		}
	}

	//无cookie
	// 如果请求没带FRJS cookie，则返回JS给客户端
	// 客户端等待了超过4秒再请求，则给客户端SetCookie
	// 客户端等待小于4秒就再次请求，则一直发JS，直到超过限制次数加入http黑名单
	if jsc.watch(cip, cipUA) {
		logrus.Debugf("cip %s in failDB,after handle: %v", cip, jsFailDB.data[cipUA])
		//生成cookie并发送给浏览器

		setck := http.Cookie{
			Name:     "FRJS",
			Value:    cipUA,
			HttpOnly: true,
			Path:     "/",
		}
		http.SetCookie(wr, &setck)
		fmt.Fprintf(wr, nodelayJS, req.RequestURI)
		return false
	}

	//不发cookie，直接发js延迟
	sendJS(wr, req)
	return false
}

//参数1：客户ip
//参数2： cip+UA的hash
//参数3： 是否冒用cookie
//返回值：是否需要设置cookie
func (jsc *JS_conf) watch(cip, cipUA string) (needSetCookie bool) {
	jsFailDB.Lock()
	defer jsFailDB.Unlock()

	//var isFirst bool
	fm, ok := jsFailDB.data[cipUA]
	if !ok {
		logrus.Debugf("首次JS访问，添加%s(%s)到map", cipUA, cip)
		fm = new(jsFailCounter)
		fm.dead = time.Duration(time.Now().UnixNano()) + jsc.AmongTime
		jsFailDB.data[cipUA] = fm
		//isFirst = true
	}

	logrus.Debugf("cipUA %s(%s) in failDB, derefer: %v", cipUA, cip, *(jsFailDB.data[cipUA]))

	//不是新生的结构体 && 延迟大于FRJsDelay && 延迟小于FRJsDelay+4s
	delay := time.Duration(time.Now().UnixNano()) - fm.jsBirth
	if fm.jsBirth != 0 && delay > FRJsDelay && delay < FRJsDelay+time.Second*4 {
		logrus.Debugf("发送js给客户端->收到客户端请求超过FRJsDelay,小于FRJsDelay+4s，可以发cookie了, %d", fm.jsBirth)
		delete(jsFailDB.data, cipUA)
		return true
	}
	fm.jsBirth = time.Duration(time.Now().UnixNano())

	fm.fail++
	logrus.Debugf("JS验证失败次数%d, %s(%s)", fm.fail, cipUA, cip)
	if fm.fail > jsc.Threshold {
		//js失败次数超限，加1级黑名单
		addL1Blacklist(cip, jsc.L1BlistTerm, jsc.CaptchaTTL, jsc.L2BlistTerm)
		logrus.Debugf("添加黑名单，JS fail MAP删除%s(%s)", cipUA, cip)
		delete(jsFailDB.data, cipUA)
	}

	return false
}

func sendJS(wr http.ResponseWriter, req *http.Request) {
	n, err := fmt.Fprintf(wr, FRJsHTML, req.RequestURI)
	if err != nil {
		logrus.Errorf("io.copy defaultJS->wr failed, %dB, %v", n, err)
	}
}

//****************** 客户端验证失败状态保存 *************
type jsFailCounter struct {
	fail    int
	jsBirth time.Duration //返回js的时间
	dead    time.Duration
}

type jsFailCounterMap struct {
	data map[string]*jsFailCounter
	sync.RWMutex
}

var jsFailDB = jsFailCounterMap{data: make(map[string]*jsFailCounter, 300)}

func jsFailMapLifeCheck() {
	tk := time.NewTicker(time.Second * 60)
	for {
		select {
		case <-tk.C:
			jsFailDB.Lock()
			for k, _ := range jsFailDB.data {
				if time.Duration(time.Now().UnixNano()) > jsFailDB.data[k].dead {
					delete(jsFailDB.data, k)
					logrus.Debugf("删除FRJS map，key=%s", k)
				}
			}
			jsFailDB.Unlock()
		}
	}
}

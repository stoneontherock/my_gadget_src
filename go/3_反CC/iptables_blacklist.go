package cc

import (
	"github.com/sirupsen/logrus"
	"os/exec"
	"sync"
	"time"
)

// ************* 下面是IPtables黑名单 ****************
//iptables黑名单唯一性：客户端IP地址
//IPtables黑名单唯一性：生命周期（到期释放）
//释放方式：到期释放/手动释放
type l2BlackListMap struct {
	data map[string]int64 //key:ip  value:释放时间UnixNano秒
	sync.RWMutex
}

var l2BlackList = l2BlackListMap{data: make(map[string]int64, 200)}
var L2BlacklistChkIntv time.Duration

var runL2LifeCheck sync.Once

func addL2Blacklist(ip string, l2term time.Duration) {

	logrus.Debugf("添加2级黑名单，%s", ip)
	l2BlackList.Lock()
	defer l2BlackList.Unlock()
	if _, ok := l2BlackList.data[ip]; ok {
		logrus.Infof("ip %s already in level2 blacklist")
		return
	}
	l2BlackList.data[ip] = time.Now().Add(l2term).UnixNano()

	go runIPtablesCMD(ip, "-A")
}

func runIPtablesCMD(ip, action string) {
	c := exec.Command("/bin/bash", "-c", `iptables `+action+` INPUT -s `+ip+` -j DROP`)

	//Run() command as block mode
	logrus.Debugf("2级黑名单操作'%s', ip:%s", action, ip)
	err := c.Run()
	if err != nil {
		logrus.Errorf("exec.Cmd.Run,%v", err)
	}
}

func l2BlacklistLifeCheck() {
	tk := time.NewTicker(L2BlacklistChkIntv)
	for {
		select {
		case <-tk.C:
			l2BlackList.Lock()
			for ip, dead := range l2BlackList.data {
				if time.Now().UnixNano() > dead {
					delete(l2BlackList.data, ip)
					go runIPtablesCMD(ip, "-D")
				}
			}
			l2BlackList.Unlock()
		}
	}
}

//下面代码是解决bug: 重启进程导致iptables规则残留
// todo: 需要将下面代码覆盖上面的代码
//
// const (
// 	periodicDelCMD = `iptables -L INPUT |
// awk -v t=$(date +%s) -F'L2TMOUT:' '
//     /L2TMOUT:/ { 
//         c=$2
//         split(c,a,"[, *]")
//         if (a[1] != "" && a[1] <= t) {
//            printf("iptables -D INPUT -s %s -j DROP -m comment --comment L2TMOUT:%s,%s\n",a[2],a[1],a[2]) 
//         }
//     }
//     END{
//       	print ":"
//     }' |bash
// `
// //	FrontendShowCMD = `iptables -L INPUT |
// //awk -v t=%d -F'L2TMOUT:' '
// //    /L2TMOUT:/ {
// //        c=$2
// //        split(c,a,"[, *]")
// //        if (a[1] != "" && a[1] > t) {
// //           print a[2]","a[1]
// //        }
// //    }`

// 	//delByIPnUnixTimeCMD = ``
// )

// func addL2Blacklist(ip string, l2term time.Duration) {
// 	elogger.Debugf("添加2级黑名单，%s", ip)
// 	cmdStr := fmt.Sprintf(`iptables -A INPUT -s %s -j DROP -m comment --comment "L2TMOUT:%d,%[1]s"`, ip, time.Now().Add(l2term).Unix())
// 	runCMD(cmdStr)
// }

// func runCMD(cmdStr string) {
// 	c := exec.Command("/bin/bash", "-c", cmdStr)
// 	//阻塞执行
// 	err := c.Run()
// 	if err != nil {
// 		elogger.Errorf("exec.Cmd.Run, %v", err)
// 	}
// }

// func l2BlacklistLifeCheck() {
// 	tk := time.NewTicker(L2BlacklistChkIntv)
// 	for range tk.C {
// 		runCMD(periodicDelCMD)
// 	}
// }

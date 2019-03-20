package circle_link_node

import (
	"github.com/sirupsen/logrus"
	"time"
)

type linkNode struct {
	count int
	next  *linkNode
}

type Circle struct {
	first   *linkNode //10个节点的环形
	refresh int64     //unixNano
	window  int64     //统计时长
}

var resolotion = int64(10)

func NewCircle(w int64) *Circle {
	first := new(linkNode)
	first.count = 1
	//logrus.Debugf("创建起点环%p", first)
	p := first
	//创建环形链表
	for i := int64(1); i < resolotion; i++ {
		p.next = new(linkNode)
		//logrus.Debugf("创建节点%d,地址%p", i, p.next)
		p = p.next
	}
	p.next = first //环末尾节点的next指向环起始节点

	var c Circle
	c.first = first
	c.window = w
	c.refresh = time.Now().UnixNano()
	//logrus.Debugf("环信息：%v", c)
	return &c
}

func (c *Circle) UpdateCircle() int {
	elapse := time.Now().UnixNano() - c.refresh

	jump := int64(elapse / (c.window / resolotion))
	//logrus.Debugf("elapse=%d,jump=%d,now=%s", elapse, jump, time.Now())
	//fmt.Printf("elapse=%d,jump=%d,now=%s\n", elapse, jump, time.Now())
	f := c.first
	for i := int64(0); i < jump; i++ {
		//fmt.Printf("Update:f=%p,f.next=%p\n", f, f.next)
		f = f.next
	}
	f.count++
	c.first = f
	c.refresh = time.Now().UnixNano()
	//fmt.Printf("Update:c.First=%p\n", c.First)

	return c.total()
}

func (c *Circle) total() int {
	if c == nil {
		return 0
	}

	var sum int
	f := c.first
	for i := int64(0); i < resolotion; i++ {
		sum += f.count
		f = f.next
	}

	logrus.Debugf("单位时间内访问总量%d", sum)
	return sum
}

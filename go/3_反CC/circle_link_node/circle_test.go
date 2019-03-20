package circle_link_node

import (
	"math/rand"
	"testing"
	"time"
)

func Test_Circle(t *testing.T) {
	t.Log("测试环型链表...")

	w := time.Second * 10
	circle := NewCircle(int64(w))

	rand.Seed(time.Now().UnixNano())
	var cc *linkNode = circle.first
	var list1 []int
	var r int
	for {
		if len(list1) > int(resolotion*2) {
			t.Fatal("失败，数量大于分辨率*2了，首尾还不相接，退出循环")
			break
		}

		for {
			r = rand.Intn(10)
			if r > 0 {
				break
			}
		}
		list1 = append(list1, r)
		cc.count = r
		t.Logf("count=%d, addr:%p, next:%p", r, cc, cc.next)

		if cc.next == circle.first {
			t.Logf("首尾节点已相接, 首:%p, 尾:%p, 尾.next:%p", circle.first, cc, cc.next)
			break
		}
		cc = cc.next
	}

	if len(list1) != int(resolotion) {
		t.Fatalf("失败，节点数(%d)不等于分辨率(%d)", len(list1), resolotion)
	}

	total := circle.total()
	sum := 0
	for _, v := range list1 {
		sum += v
	}

	if total != sum {
		t.Fatalf("链表求和(%d)错误,实际总和=%d", total, sum)
	}

	t.Log("ok,节点数等于分辨率,节点求和通过")

	time.Sleep(time.Second * 3)
	circle.UpdateCircle()
	cu := circle.first
	t.Logf("cu addr: %p", cu)
	var list2 []int
	for {
		list2 = append(list2, cu.count)
		if cu.next == circle.first {
			break
		}
		cu = cu.next
	}
	if list1[3]+1 == list2[0] {
		t.Logf("更新环形链表索引4,OK")
	} else {
		t.Errorf("更新环形链表索引4,失败")
	}

	t.Logf("------更新前vs更新后------\nlist1:%v\nlist2:%v\n", list1, list2)

	time.Sleep(time.Second * 7)
	circle.UpdateCircle()
	cx := circle.first
	t.Logf("cx addr: %p", cx)
	var list3 []int
	for {
		list3 = append(list3, cx.count)
		if cx.next == circle.first {
			break
		}
		cx = cx.next
	}
	t.Logf("list3:%v\n", list3)
	if list3[0] == 1 {
		t.Logf("更新环形链表索引8,OK")
	} else {
		t.Errorf("更新环形链表索引8,失败")
	}
}

package goskiplist

import (
	"fmt"
	"math/rand"
)

const (
	SKIPLIST_MAXLEVEL = 32
	SKIPLIST_P        = 0.25
)

type GskiplistLevel struct {
	forward *GskiplistNode
	span    uint64
}
type GskiplistNode struct {
	ele      string
	score    float64
	backward *GskiplistNode
	level    []GskiplistLevel
}
type Gskiplist struct {
	header *GskiplistNode
	tail   *GskiplistNode
	length uint64
	level  int
}

func (node *GskiplistNode) GetEleWithScore() string {
	return node.ele + fmt.Sprintf("%.2f", node.score)
}
func createNode(level int, score float64, ele string) *GskiplistNode {
	node := &GskiplistNode{
		ele:      ele,
		score:    score,
		level:    make([]GskiplistLevel, level),
		backward: nil,
	}
	return node
}

func CreateSkiplist() *Gskiplist {
	sl := &Gskiplist{
		level:  1,
		length: 0,
		header: createNode(SKIPLIST_MAXLEVEL, 0, ""), //创建头节点
		tail:   nil,
	}
	//初始化头节点的每一层
	for j := 0; j < SKIPLIST_MAXLEVEL; j++ {
		sl.header.level[j].forward = nil
		sl.header.level[j].span = 0
	}
	sl.header.backward = nil
	return sl
}

// 向跳表插入一个节点，同时返回插入好的节点。
// ele不能为空串，否则返回nil。
func (this *Gskiplist) Insert(score float64, ele string) *GskiplistNode {
	if ele == "" {
		return nil
	}
	update := make([]*GskiplistNode, SKIPLIST_MAXLEVEL)
	rank := make([]uint64, SKIPLIST_MAXLEVEL)
	var x *GskiplistNode

	x = this.header
	//更新update以及rank
	for i := this.level - 1; i >= 0; i-- {
		rank[i] = 0
		if i != this.level-1 {
			rank[i] = rank[i+1]
		}
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score && x.level[i].forward.ele < ele)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}
	level := this.randomLevel()
	//更新最大层数
	if level > this.level {
		for i := this.level; i < level; i++ {
			rank[i] = 0
			update[i] = this.header
			update[i].level[i].span = this.length
		}
		this.level = level
	}
	x = createNode(level, score, ele)

	//插入操作
	for i := 0; i < level; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x
		//更新x和前一个节点的span
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	//更新更高层
	for i := level; i < this.level; i++ {
		update[i].level[i].span++
	}

	//更新前节点指针指向
	x.backward = nil
	if update[0] != this.header {
		x.backward = update[0]
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		this.tail = x
	}
	this.length++
	return x
}

func (this *Gskiplist) randomLevel() int {
	level := 1
	f := float32(0x7fff)
	threshold := int(SKIPLIST_P * f)
	for rand.Intn(int(f)) < threshold {
		level += 1
	}
	level = min(SKIPLIST_MAXLEVEL, level)
	return level
}

// 删除节点，返回这个节点以及是否成功
func (this *Gskiplist) Delete(score float64, ele string) (*GskiplistNode, bool) {
	update := make([]*GskiplistNode, SKIPLIST_MAXLEVEL)
	var x *GskiplistNode

	x = this.header
	for i := this.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score && x.level[i].forward.ele < ele)) {
			x = x.level[i].forward
		}
		update[i] = x
	}

	x = x.level[0].forward
	//从底层删除
	if x != nil && x.score == score && x.ele == ele {
		this.deleteNode(x, update)
		return x, true
	}
	//未找到对应节点
	return nil, false
}

func (this *Gskiplist) deleteNode(x *GskiplistNode, update []*GskiplistNode) {
	for i := 0; i < this.level; i++ {
		if update[i].level[i].forward == x {
			//在这一层，存在x
			update[i].level[i].span += x.level[i].span - 1
			update[i].level[i].forward = x.level[i].forward
		} else {
			//不存在则只更新span
			update[i].level[i].span--
		}
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		this.tail = x.backward
	}

	//若x独占高层，需要逐个清除
	for this.level > 1 && this.header.level[this.level-1].forward == nil {
		this.level--
	}
	this.length--
}

// 获取节点的排名
func (this *Gskiplist) GetRank(score float64, ele string) uint64 {
	var x *GskiplistNode
	var rank uint64 = 0
	x = this.header
	for i := this.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score && x.level[i].forward.ele <= ele)) {
			rank += x.level[i].span
			x = x.level[i].forward
		}
		//需要检测x.ele是否为空串，因为它可能是头节点。
		if x.ele != "" && x.score == score && x.ele == ele {
			return rank
		}
	}
	return 0
}

// 根据排名获取节点
func (this *Gskiplist) GetElementByRank(rank uint64) *GskiplistNode {
	return this.getElementByRankFromNode(this.header, this.level-1, rank)
}
func (this *Gskiplist) getElementByRankFromNode(startNode *GskiplistNode, startLevel int, rank uint64) *GskiplistNode {
	x := startNode
	var traversed uint64
	for i := startLevel; i >= 0; i-- {
		for x.level[i].forward != nil && traversed+x.level[i].span <= rank {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		//遍历完一层，查看是否到达
		if traversed == rank {
			return x
		}
	}
	return nil
}

// 根据分数区间获取数据集合，返回数据的ele集合
func (this *Gskiplist) GetElementsRangeByScore(low float64, high float64) (ans []string) {
	x := this.header
	var i int
	for i = this.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && x.level[i].forward.score < low {
			x = x.level[i].forward
		}
	}
	x = x.level[0].forward
	for x != nil && x.score <= high {
		ans = append(ans, x.ele)
		x = x.level[0].forward
	}
	return ans
}

package raft

import "log"

// raft集群中leader数据变更后有个log replica 的过程，就是把数据同步到各个follower节点
type datalog struct {
	term   int
	action string
}

// 类似java里的equals()方法，用于比较两个对象是否相等
func (d *datalog) identical(data datalog) bool {
	return d.term == data.term && d.action == data.action
}

type submittedItems struct {
	logs     []datalog
	logIndex int
}

func (d *submittedItems) getLatestLogs() *datalog {
	log.Println("size of datalog:", len(d.logs))
	if len(d.logs) > 0 {
		// 获取logs数组里最后一个元素
		return &(d.logs[len(d.logs)-1])
	} else {
		return &datalog{}
	}
}

// 类似java里的equals()方法，比较两个submittedItems是否一致
func (d *submittedItems) identicalWith(b *submittedItems) bool {
	return d.logIndex == b.logIndex && d.getLatestLogs().term == b.getLatestLogs().term
}

func (s *submittedItems) add(data datalog) {
	s.logIndex++
	s.logs = append(s.logs, data)
}

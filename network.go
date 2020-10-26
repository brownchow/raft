package raft

import (
	"log"
	"time"
)

// 用go的channel模拟网络，感觉叫nodeList更合适，一个map包含了网络中所有节点id和它对应的chan
type network struct {
	// key:val   节点id:节点对应的消息通道 message queue
	recvQueue map[int]chan Message
}

//CreateNetwork nodes ...int 每个节点都有一个id
func CreateNetwork(nodes ...int) *network {
	// map容量为0
	nt := network{recvQueue: make(map[int]chan Message, 0)}

	//遍历map，初始化每个chan的容量为1024
	for _, node := range nodes {
		nt.recvQueue[node] = make(chan Message, 1024)
	}
	return &nt
}

//nodeNetwork 节点，还有节点所在的网络
type nodeNetwork struct {
	id  int
	net *network
}

// 创建一个节点，并指定所在网络
func (n *network) addNodeToNetwork(id int) nodeNetwork {
	return nodeNetwork{id: id, net: n}
}

// 发送消息
func (n *network) sendTo(m Message) {
	log.Println("Send message from:", m.from, " send to ", m.to, " val:", m.val, " typ:", m.typ)
	n.recvQueue[m.to] <- m
}

// 接受消息，入参：发送Msg的节点id(hashMap查表嘛，取出对应的chan)，出参：取出节点id对应的chan的Msg
func (n *network) recvFrom(id int) *Message {
	select {
	case retMsg := <-n.recvQueue[id]:
		log.Println("Recev msg from: ", retMsg.from, " send to:", retMsg.to, " val:", retMsg.val, "  typ:", retMsg.typ)
		return &retMsg
	case <-time.After(time.Second):
		log.Println("id: ", id, "don't get mesasage, timeout")
		return nil
	}
}

// 发送消息
func (n *nodeNetwork) send(m Message) {
	n.net.sendTo(m)
}

// 接收消息
func (n *nodeNetwork) recv() *Message {
	return n.net.recvFrom(n.id)
}

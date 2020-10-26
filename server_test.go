package raft

import (
	"log"
	"testing"
	"time"
)

func TestBasicNetwork(t *testing.T) {
	log.Println("TestBasicNetowk........................")
	// 入参是每个节点的id，函数体会初始化每个节点对应的chan，存放需要发送的Msg队列
	nt := CreateNetwork(1, 3, 5, 2, 4)
	// 必须起一个新的goroutine去检测Msg???
	go func() {
		log.Println("recv from 5:", nt.recvFrom(5))
		log.Println("recv from 1:", nt.recvFrom(1))
		log.Println("recv from 3:", nt.recvFrom(3))
		log.Println("recv from 2:", nt.recvFrom(2))
		m := nt.recvFrom(4)
		if m == nil {
			t.Errorf("No message detected.")
		} else {
			log.Println("recv from 4:", m)
		}
	}()
}

func TestFollowerElectionToLeader(t *testing.T) {
	log.Println("--------TestFollowerElectionToLeader-------")
	nt := CreateNetwork(1, 2, 3, 4, 5)
	// 参数： 节点id， 节点角色，节点所在网络，其他节点列表
	// 每个节点的默认id都是follower角色
	nServer1 := NewServer(1, Follower, nt.addNodeToNetwork(1), 2, 3, 4, 5)
	nServer2 := NewServer(2, Follower, nt.addNodeToNetwork(2), 1, 3, 4, 5)
	nServer3 := NewServer(3, Follower, nt.addNodeToNetwork(3), 2, 1, 4, 5)
	nServer4 := NewServer(4, Follower, nt.addNodeToNetwork(4), 2, 3, 1, 5)
	nServer5 := NewServer(5, Follower, nt.addNodeToNetwork(5), 2, 3, 1, 5)

	log.Println(nServer1.id, nServer2.id, nServer3.id, nServer4.id, nServer5.id)

	//Set server1 an action.
	nServer1.AppendEntries(datalog{term: 1, action: "x<-1"})
	log.Println("Assign value to server 1 ")

	//Wait server1 become Leader.
	for i := 0; i <= 10; i++ {
		if nServer1.WhoAreYou() == Leader {
			log.Println("1 become leader done.")
			return
		}
		log.Println("1 still not leader:", nServer1.WhoAreYou())
		time.Sleep(time.Second)
	}
	t.Error("No one become leader on basic")
}

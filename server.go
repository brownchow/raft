package raft

import (
	"log"
	"math/rand"
	"time"
)

type Role int

const (
	Follower Role = iota + 1
	Candidate
	Leader
)

//RaftServer 服务器，服务器的角色可以是leader, candidate, follower
type RaftServer struct {
	//Persistent state on all servers
	currentTerm int
	voteFor     int
	log         []int

	//Validate state for all servers
	commitIndex int
	lastApplied int

	//Leader state will reinit on election.
	nextIndex  []int
	matchIndex []int

	//Basic Server state
	id          int
	expiredTime int //Hearbit expired time (by millisecond.)
	role        Role
	nt          nodeNetwork
	msgRecvTime time.Time //Message receive time
	nodeList    []int     //id list exist in this network.包括自身
	term        int       //term about current time seq
	db          submittedItems

	isAlive bool //To determine if server still alive, for kill testing.

	//For candidator
	HasVoted      bool //To record if already vote others.
	acceptVoteMsg []Message
}

//NewServer :New a server and given a random expired time.
func NewServer(id int, role Role, nt nodeNetwork, nodeList ...int) *RaftServer {
	rand.Seed(time.Now().UnixNano())
	expiredMiliSec := rand.Intn(5) + 1
	serv := &RaftServer{
		id:          id,
		role:        role, // 指定角色，默认都是follower
		nt:          nt,
		expiredTime: expiredMiliSec, //随机的过期时间，感觉代码有问题啊，expiredTime这个参数根本没用到
		isAlive:     true,
		nodeList:    nodeList,
		db:          submittedItems{}}
	// 起一个goroutine在后台运行，去执行serverLoop
	go serv.runServerLoop()
	return serv
}

//AssignAction : Assign a assign to any of server.
func (sev *RaftServer) AppendEntries(action datalog) {
	//TODO. Add action into logs and leader will announce to all other servers.
	switch sev.role {
	case Leader:
		//Apply to all followers
	case Candidate:
		//TBC.
	case Follower:
		//Run election to leader
		sev.requestVote(action)
	}
}

//Whoareyou :Check rule function for testing verification.
func (sev *RaftServer) WhoAreYou() Role {
	return sev.role
}

// 启动服务器死循环
func (sev *RaftServer) runServerLoop() {

	for {
		// 不需要break语句
		switch sev.role {
		case Leader:
			sev.runLeaderLoop()
		case Candidate:
			sev.runCandidateLoop()
		case Follower:
			sev.runFollowerLoop()
		}
		//timer base on milli-second.
		time.Sleep(100 * time.Millisecond)
	}
}

//For flower -> candidate，超时时间到，从follower变成candidate，首先给自己投票，然后让集群中的其他节点给自己投票
func (srv *RaftServer) requestVote(action datalog) {
	m := Message{
		from:         srv.id,
		typ:          RequestVote,
		val:          action,
		term:         srv.term,
		lastLogIndex: 0,
	}
	// 给网络中包括自身在内的所有节点发送RequestVote Message
	for _, node := range srv.nodeList {
		m.to = node
		srv.nt.send(m)
	}
	// 一般情况下，所有节点都会返回同意，因此肯定超过一半的节点同意，直接变成candidate
	//Send request Vote and change self to Candidate.
	srv.roleChange(Candidate)
	log.Println("Now ID:", srv.id, " become candidate->", srv.role)
}

func (sev *RaftServer) sendHearbit() {
	latestData := sev.db.getLatestLogs()
	// 给网络中除自身以外的所有节点发送心跳信息，附带最新的log
	for _, node := range sev.nodeList {
		hbMsg := Message{from: sev.id, to: node, typ: Heartbeat, val: *latestData}
		sev.nt.send(hbMsg)
	}
}

// leader节点要做的事情
func (sev *RaftServer) runLeaderLoop() {
	log.Println("ID:", sev.id, " Run leader loop")
	//leader给follower发送心跳信息
	sev.sendHearbit()
	//从自身的chan中接收其他节点回复的信息
	recevMsg := sev.nt.recv()
	if recevMsg == nil {
		return
	}
	switch recevMsg.typ {
	case Heartbeat:
		//TODO. other leaders HB, should not happen.
		return

	case HeartbeatFeedback:
		// 正常应该返回这种信息
		//TODO. notthing happen in this.
		return

	case RequestVote:
		return
	case AcceptVote:
		return
	case WinningVote:
		return
	}
	//TODO. if get bigger TERM request, back to follower
}

// candidate节点应该做的事情
func (sev *RaftServer) runCandidateLoop() {
	log.Println("ID:", sev.id, " Run candidate loop")
	//TODO. send RequestVote to all others
	// 从自身chan中接收Msg
	recvMsg := sev.nt.recv()

	if recvMsg == nil {
		log.Println("ID:", sev.id, " no msg, return.")
		return
	}
	switch recvMsg.typ {
	case Heartbeat:
		//TODO. Leader heartbit
		return
	case RequestVote:
		//TODO. other candidate request vote.
		return
	case AcceptVote:
		sev.acceptVoteMsg = append(sev.acceptVoteMsg, *recvMsg)
		// candidate 接收到集群中超过半数节点的投票，就转变成leader角色
		log.Println("[candidate]: has ", len(sev.acceptVoteMsg), " still not reach ", sev.majorityCount())
		if len(sev.acceptVoteMsg) >= sev.majorityCount() {
			sev.roleChange(Leader)

			// send winvote to all and notify there is new leader
			for _, node := range sev.nodeList {
				hbMsg := Message{from: sev.id, to: node, typ: WinningVote}
				sev.nt.send(hbMsg)
			}
		}

		return
	case WinningVote:
		//Receive winvote from other candidate means we need goback to follower
		sev.roleChange(Follower)
		return

	}
	//TODO. check if prompt to leader.

	//TODO. If not, back to follower
}

// follower节点应该做的事情
func (sev *RaftServer) runFollowerLoop() {
	log.Println("ID:", sev.id, " Run follower loop")

	//TODO. check if leader no heartbeat to change to candidate.

	recvMsg := sev.nt.recv()

	if recvMsg == nil {
		log.Println("ID:", sev.id, " no msg, return.")
		return
	}
	switch recvMsg.typ {
	case Heartbeat:
		// return
		if !sev.db.getLatestLogs().identical(recvMsg.GetVal()) {
			//Data not exist, add it. (TODO)
			sev.db.add(recvMsg.GetVal())
		}

		//Send it back HeartBeat
		recvMsg.to = recvMsg.from
		recvMsg.from = sev.id
		recvMsg.typ = HeartbeatFeedback
		sev.nt.send(*recvMsg)
		return
	case RequestVote:
		//Handle Request Vote from candidate.
		//If doesn't vote before, will vote.
		if sev.voteFor == 0 {
			recvMsg.to = recvMsg.from
			recvMsg.from = sev.id
			recvMsg.typ = AcceptVote
			sev.nt.send(*recvMsg)
			sev.voteFor = recvMsg.from
		} else {
			//Don't do anything if you already vote.
			//Only vote when first candidate request vote comes.
		}
	case WinningVote:
		//Clean variables.
		sev.voteFor = recvMsg.from

	}
}

func (sev *RaftServer) majorityCount() int {
	return len(sev.nodeList)/2 + 1
}

func (sev *RaftServer) roleChange(newRole Role) {
	log.Println("note:", sev.id, " change role from ", sev.role, " to ", newRole)
	sev.role = newRole
}

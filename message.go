package raft

type msgType int

const (
	Heartbeat msgType = iota + 1
	HeartbeatFeedback
	RequestVote
	AcceptVote
	WinningVote
)

//From has different meaning depending on different state
// [AppendEntries]:  from = leaderID
// [RequestVote]: from = candidateID
type Message struct {
	from int
	to   int
	typ  msgType
	term int
	val  datalog

	// log replication 的过程
	lastLogIndex int
	lastLogTerm  int
	leaderCommit int
	success      bool // 是否成功
}

func (m *Message) GetMsgTerm() int {
	return m.term
}

func (m *Message) GetVal() datalog {
	return m.val
}

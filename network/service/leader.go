package service

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/messages"
	"encoding/json"
	"fmt"
)

const (
	DeadType = iota
	TranMiss
	BlockOvertime
)

var (
	Leader LeaderT
)

type LeaderT struct {
	OnChanging bool
	LeaderID int64
	LeaderMsgChan chan LeaderMsg
	DeadMsgMap []DeadMsg
	//key is PId'String
	TranMissMsgMap map[string][]TranMissMsg
	BlockOvertimeMsgMap map[int64][]BlockOvertimeMsg
}

func (leader *LeaderT) ProcessLeaderMsg() {
	for {
		msg := <- leader.LeaderMsgChan
		if msg.LeaderID != leader.LeaderID {
			fmt.Println("Outdated msg")
			continue
		}
		switch msg.Type {
		case DeadType:
			var deadMsg DeadMsg
			err := json.Unmarshal(msg.Data, deadMsg)
			if err != nil {
				fmt.Println("Process LeaderMsg failed", err)
				continue
			}
		case TranMiss:
		case BlockOvertime:
		default:
			fmt.Println("Unknown LeaderMsg Type")
		}
	}
}

func (*LeaderT) Parse(data []byte) *LeaderMsg {
	var msg LeaderMsg
	err := json.Unmarshal(data, msg)
	if err != nil {
		fmt.Println("Parse Leader Msg failed", err)
		return nil
	}
	return &msg
}

func (leader *LeaderT) LeaderVote(t int, id interface{}) {
	switch t {
	case DeadType:

	case TranMiss:
		var msg LeaderVoteMsg
		tId := id.(messages.PId).String()
		msg.Data = []byte(leader.TranMissMsgMap[tId][0].TId.String())
		for _, val := range leader.TranMissMsgMap[tId] {
			msg.Sigs = append(msg.Sigs, val.Sig)
		}

	case BlockOvertime:
	default:
		fmt.Println("Unknown LeaderVote type")
	}
}

type LeaderMsg struct {
	Type int
	LeaderID int64
	Data []byte
}

type DeadMsg struct {
	Sig []byte
}

type TranMissMsg struct {
	TId messages.PId
	Sig []byte
}

type BlockOvertimeMsg struct {
	BId int64
	Sig []byte
}

type LeaderVoteMsg struct {
	Data []byte
	Sigs [][]byte
}

type LeaderTInterface interface {
	ProcessLeaderMsg()
	Parse(data []byte) *LeaderMsg
	LeaderVote(t int, id interface{})
}

func init() {
	Leader = LeaderT{
		LeaderMsgChan: make(chan LeaderMsg, conf.BCDnsConfig.LeaderMsgBufferSize),
	}
}

func TurnLeader () {

}
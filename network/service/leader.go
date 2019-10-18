package service

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/messages"
	"encoding/json"
	"fmt"
	"log"
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
	LeaderId int64
	TermId int64
	ViewChangeMsgChan chan []byte
	RetrieveMsgChan chan []byte
	RetrieveMsgs map[int64]map[string]ViewRetrieveMsg
	ViewChangeMsgs map[ViewChangeMsgData][]ViewChangeMsg
}

func (leader *LeaderT) ProcessViewChangeMsg() {
	var msg ViewChangeMsg
	for {
		msgByte := <- leader.ViewChangeMsgChan
		err := json.Unmarshal(msgByte, msg)
		if err != nil {
			fmt.Println("Process viewchange msg failed", err)
			continue
		}
		if msg.TermId != leader.TermId {
			fmt.Println("Outdated msg")
			continue
		}
		if !checkType(msg.ViewChangeType) {
			fmt.Println("Illegal msg type")
			continue
		}
		dataBytes, err := json.Marshal(msg.ViewChangeMsgData)
		if err != nil {
			fmt.Println("Process viewchange msg failed", err)
			continue
		}
		if service.CertificateAuthorityX509.VerifySignature(msg.Sig, dataBytes, msg.HostName) {
			if _, ok := leader.ViewChangeMsgs[msg.ViewChangeMsgData]; !ok {
				leader.ViewChangeMsgs[msg.ViewChangeMsgData] = append(leader.ViewChangeMsgs[msg.ViewChangeMsgData],
					msg)

				if len(leader.ViewChangeMsgs[msg.ViewChangeMsgData]) > 2 * service.CertificateAuthorityX509.GetF() + 1 {
					go leader.LeaderVote(msg.ViewChangeMsgData)
				}
			}
		}
	}
}

func (leader *LeaderT) LeaderVote(id ViewChangeMsgData) {
	msg := LeaderVoteMsg{
		Msgs: leader.ViewChangeMsgs[id],
	}
	msgByte, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("LeaderVote failed", err)
		return
	}
	P2PNet.BroadcastMsg(msgByte)
}

func (leader *LeaderT) ProcessRetrieveMsg() {
	var msg ViewRetrieveMsg
	for {
		msgByte := <- leader.RetrieveMsgChan
		err := json.Unmarshal(msgByte, msg)
		if err != nil {
			fmt.Println("Process retrieve msg failed", err)
			continue
		}
		if msg.Retrieve {
			msg.LeaderId, msg.TermId, msg.HostName = leader.LeaderId, leader.TermId, conf.BCDnsConfig.HostName
		} else {
			if v, ok := leader.RetrieveMsgs[msg.TermId]; ok {
				if _, ok = v[msg.HostName]; !ok {
					if leader.RetrieveMsgs[msg.TermId][msg.HostName] = msg; len(leader.RetrieveMsgs[msg.TermId]) >= service.CertificateAuthorityX509.GetF() + 1 {
						if msg.TermId == -1 {
							leader.TermId, leader.LeaderId = 0, 0
						} else {
							leader.TermId, leader.LeaderId = msg.TermId, msg.LeaderId
						}
						leader.RetrieveMsgs = make(map[int64]map[string]ViewRetrieveMsg)
					}
				}
			} else {
				leader.RetrieveMsgs[msg.TermId] = make(map[string]ViewRetrieveMsg)
				leader.RetrieveMsgs[msg.TermId][msg.HostName] = msg
				if leader.RetrieveMsgs[msg.TermId][msg.HostName] = msg; len(leader.RetrieveMsgs[msg.TermId]) >= service.CertificateAuthorityX509.GetF() + 1 {
					if msg.TermId == -1 {
						leader.TermId, leader.LeaderId = 0, 0
					} else {
						leader.TermId, leader.LeaderId = msg.TermId, msg.LeaderId
					}
					leader.RetrieveMsgs = make(map[int64]map[string]ViewRetrieveMsg)
				}
			}
		}
	}
}

type ViewChangeMsg struct {
	ViewChangeMsgData
	Sig []byte
}

type ViewChangeMsgData struct {
	Type uint8
	HostName string
	ViewChangeType int
	TermId, BId int64
	//key is PId'String
	TId messages.PId
}

type LeaderVoteMsg struct {
	Type uint8
	Msgs []ViewChangeMsg
}

type LeaderTInterface interface {
	ProcessViewChangeMsg()
	LeaderVote(ViewChangeMsgData)
	ProcessRetrieveMsg()
}

type ViewRetrieveMsg struct {
	Type uint8
	Retrieve bool
	HostName string
	TermId, LeaderId int64
}

func init() {
	msg := ViewRetrieveMsg{
		Type:conf.ViewRetrieve,
		Retrieve:true,
	}
	msgByte, err := json.Marshal(msg)
	if err != nil {
		log.Fatal("Leader init failed", err)
	}
	P2PNet.BroadcastMsg(msgByte)
	Leader = LeaderT{
		OnChanging: false,
		LeaderId: -1,
		TermId: -1,
		ViewChangeMsgChan: make(chan []byte, conf.BCDnsConfig.LeaderMsgBufferSize),
		ViewChangeMsgs: make(map[ViewChangeMsgData][]ViewChangeMsg),
	}
}

//static method
func TurnLeader () {
	service.CertificateAuthorityX509.Mutex.Lock()
	defer service.CertificateAuthorityX509.Mutex.Unlock()

	Leader.LeaderId = (Leader.LeaderId + 1) % int64(service.CertificateAuthorityX509.GetNetworkSize())
}

func checkType(t int) bool {
	if t >= DeadType && t <= BlockOvertime {
		return true
	}
	return false
}
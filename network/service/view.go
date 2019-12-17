package service

import (
	"BCDns_0.1/blockChain"
	service2 "BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/messages"
	"encoding/json"
	"fmt"
	"github.com/sasha-s/go-deadlock"
)

type ViewManagerT struct {
	Mutex              deadlock.Mutex
	OnChange           bool
	View               int64
	LeaderId           int64
	ViewChangeMsgs     map[string]blockChain.ViewChangeMessage
	JoinReplyMessages  map[string]JoinReplyMessage
	JoinMessages map[string]JoinMessage
	InitLeaderMessages map[string]InitLeaderMessage
}

var (
	ViewManager *ViewManagerT
)

func NewViewManager() (*ViewManagerT, error) {
	manager := &ViewManagerT{
		ViewChangeMsgs:     map[string]blockChain.ViewChangeMessage{},
		JoinReplyMessages:  map[string]JoinReplyMessage{},
		JoinMessages: map[string]JoinMessage{},
		InitLeaderMessages: map[string]InitLeaderMessage{},
	}
	manager.View = -1
	manager.LeaderId = -1
	return manager, nil
}

func (m *ViewManagerT) Start() {
	for {
		select {
		case msg := <-JoinReplyChan:
			if msg.View != -1 {
				m.View = msg.View
				return
			}
			m.JoinReplyMessages[msg.From] = msg
			fmt.Println("1", len(m.JoinReplyMessages), msg)
			if service2.CertificateAuthorityX509.Check(len(m.JoinReplyMessages) + len(m.JoinMessages)) {
				initLeaderMsg, err := NewInitLeaderMessage(Net.GetAllNodeIds())
				if err != nil {
					logger.Warningf("[ViewManagerT.Start] NewInitLeaderMessage error=%v", err)
					panic(err)
				}
				jsonData, err := json.Marshal(initLeaderMsg)
				if err != nil {
					logger.Warningf("[ViewManagerT.Start] json.Marshal error=%v", err)
					panic(err)
				}
				Net.BroadCast(jsonData, InitLeaderMsg)
			}
		case msg := <- JoinChan:
			m.JoinMessages[msg.From] = msg
			fmt.Println("2", len(m.JoinMessages), msg)
			if service2.CertificateAuthorityX509.Check(len(m.JoinReplyMessages) + len(m.JoinMessages)) {
				initLeaderMsg, err := NewInitLeaderMessage(Net.GetAllNodeIds())
				if err != nil {
					logger.Warningf("[ViewManagerT.Start] NewInitLeaderMessage error=%v", err)
					panic(err)
				}
				jsonData, err := json.Marshal(initLeaderMsg)
				if err != nil {
					logger.Warningf("[ViewManagerT.Start] json.Marshal error=%v", err)
					panic(err)
				}
				Net.BroadCast(jsonData, InitLeaderMsg)
			}
		case msgByte := <-InitLeaderChan:
			var msg InitLeaderMessage
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				logger.Warningf("[ViewManagerT.Start] json.Unmarshal error+%v", err)
				continue
			}
			if !msg.VerifySignature() {
				logger.Warningf("[ViewManagerT.Start] InitLeaderMeseaderId + 1)sage.VerifySignature failed")
				continue
			}
			m.InitLeaderMessages[msg.From] = msg
			fmt.Println("initleader", len(m.InitLeaderMessages), msg)
			if service2.CertificateAuthorityX509.Check(len(m.InitLeaderMessages)) {
				m.View, m.LeaderId = m.GetLeaderNode()
				if m.View == -1 {
					panic("[ViewManagerT.Start] GetLeaderNode failed")
				}
				return
			}
		}
	}
}

func (m *ViewManagerT) GetLeaderNode() (int64, int64) {
	count := make([]int, service2.CertificateAuthorityX509.GetNetworkSize())
	for _, msg := range m.InitLeaderMessages {
		for _, id := range msg.NodeIds {
			count[id]++
		}
	}
	for i := int64(len(count) - 1); i >= 0; i-- {
		if service2.CertificateAuthorityX509.Check(count[i]) {
			return i, i
		}
	}
	return -1, -1
}

func (m *ViewManagerT) Run() {
	for {
		select {
		case msgByte := <-ViewChangeChan:
			var msg blockChain.ViewChangeMessage
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				logger.Warningf("[ViewManagerT.Run] json.Unmarshal error=%v", err)
				continue
			}
			if msg.View != m.View {
				continue
			}
			if !msg.VerifySignature() {
				logger.Warningf("[ViewManagerT.Run] ViewChangeMessage.Signature is invalid")
				continue
			}
			if !msg.VerifySignatures() {
				logger.Warningf("[ViewManagerT.Run] ViewChangeMessage.Signatures is invalid")
				continue
			}
			m.ViewChangeMsgs[msg.From] = msg
			if service2.CertificateAuthorityX509.Check(len(m.ViewChangeMsgs)) {
				localH, h := m.GetLatestHeight()
				if localH != h {
					StartDataSync(localH, h)
				}
				if m.IsNextLeader() {
					newViewMsg, err := blockChain.NewNewViewMessage(m.ViewChangeMsgs, m.View)
					if err != nil {
						logger.Warningf("[ViewManagerT.Run] NewNewViewMessage error=%v", err)
						continue
					}
					jsonData, err := json.Marshal(newViewMsg)
					if err != nil {
						logger.Warningf("[ViewManagerT.Run] json.Marshal error=%v", err)
						continue
					}
					Net.BroadCast(jsonData, NewViewMsg)
				}
			}
		case msgByte := <-NewViewChan:
			var msg blockChain.NewViewMessage
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				logger.Warningf("[ViewManagerT.Run] json.Unmarshal error=%v", err)
				continue
			}
			if m.View != msg.View {
				continue
			}
			if !msg.VerifySignature() {
				logger.Warningf("[ViewManagerT.Run] NewViewMessage.Signature is invalid")
				continue
			}
			if !msg.VerifyMsgs() {
				logger.Warningf("[ViewManagerT.Run] NewViewMessage.msgs is invalid")
				continue
			}
			m.View++
			m.LeaderId = m.View % int64(service2.CertificateAuthorityX509.GetNetworkSize())
			m.ViewChangeMsgs = map[string]blockChain.ViewChangeMessage{}
		}
	}
}

func (m *ViewManagerT) GetLatestHeight() (uint, uint) {
	lastBlock, err := blockChain.BlockChain.GetLatestBlock()
	if err != nil {
		return 0, 0
	}
	h := lastBlock.Height
	for _, msg := range m.ViewChangeMsgs {
		if h < msg.Height {
			h = msg.Height
		}
	}
	return lastBlock.Height, h
}

func (m *ViewManagerT) IsLeader() bool {
	return service2.CertificateAuthorityX509.IsLeaderNode(m.LeaderId)
}

func (m *ViewManagerT) IsNextLeader() bool {
	return service2.CertificateAuthorityX509.IsLeaderNode((m.LeaderId + 1) %
		int64(service2.CertificateAuthorityX509.GetNetworkSize()))
}

func (m *ViewManagerT) StartChanging() {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.OnChange = true
}

func (m *ViewManagerT) IsOnChanging() bool {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	return m.OnChange
}

func StartDataSync(lastH, h uint) {
	for i := lastH; i <= h; i++ {
		syncMsg, err := messages.NewDataSyncMessage(i)
		if err != nil {
			logger.Warningf("[DataSync] NewDataSyncMessage error=%v", err)
			continue
		}
		jsonData, err := json.Marshal(syncMsg)
		if err != nil {
			logger.Warningf("[DataSync] json.Marshal error=%v", err)
			continue
		}
		Net.BroadCast(jsonData, DataSyncMsg)
	}
}

package service

import (
	"BCDns_0.1/consensusMy/service"
	"encoding/json"
	"sync"
)

type ViewManagerT struct {
	Mutex sync.Mutex
	OnChange bool
	View int
	LeaderId int
	ViewChangeMsgs map[string]service.ViewChangeMessage
	Proof map[string][]byte
	RedoMsgs map[string]map[string]interface{}
}

var (
	ViewManager *ViewManagerT
)

func NewViewManager() (*ViewManagerT, error) {
	manager := new(ViewManagerT)
	manager.View = -1
	manager.LeaderId = -1
	return manager, nil
}

func (m *ViewManagerT) Start() {
	for {
		select {
		case msgByte := <- JoinReplyChan:
			var msg JoinReplyMessage
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				logger.Warningf("[ViewManagerT.Start] json.Unmarshal error=%v", err)
				continue
			}

		}
	}
}
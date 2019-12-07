package service

import (
	"BCDns_0.1/blockChain"
	"BCDns_0.1/messages"
	"BCDns_0.1/network/service"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"
)

var (
	ProposalMessageChan chan []byte
)

type Leader struct {
	Mutex        sync.Mutex
	MessagePool  ProposalMessagePool
	BlockConfirm bool
}

func init() {
	ProposalMessageChan = make(chan []byte, 1024)
}

func (l *Leader) Run(done chan uint) {
	defer close(done)
	interrupt := make(chan int)
	interruptTimer := make(chan int)
	go func() {
		for true {
			select {
			case <-time.After(10 * time.Second):
				if l.BlockConfirm {
					interrupt <- 1
				}
			case <-interruptTimer:
				interrupt <- 1
			}
		}
	}()
	for {
		select {
		case msgByte := <-ProposalMessageChan:
			var msg ProposalMessage
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				logger.Warningf("[Leader.Run] json.Unmarshal error=%v", err)
				continue
			}
			switch msg.Type {
			case Add:
				if !msg.ValidateAdd() {
					logger.Warningf("[Leader.Run] ValidateAdd failed")
					continue
				}
			case Del:
				if !msg.ValidateDel() {
					logger.Warningf("[Leader.Run] ValidateDel failed")
					continue
				}
			case Mod:
				if !msg.ValidateMod() {
					logger.Warningf("[Leader.Run] ValidateMod failed")
					continue
				}
			}
			l.MessagePool.AddProposal(msg)
			if l.BlockConfirm && l.MessagePool.Size() >= blockChain.BlockMaxSize {
				interruptTimer <- 1
			}
		case <-interrupt:
			if !service.Leader.IsLeader() {
				continue
			}
			if l.MessagePool.Size() <= 0 {
				fmt.Printf("[Leader.Run] CurrentBlock is empty\n")
				continue
			}
			validP, abandonedP := CheckProposals(l.MessagePool.ProposalMessages[:blockChain.BlockMaxSize])
			block, err := blockChain.BlockChain.MineBlock(validP)
			if err != nil {
				logger.Warningf("[Leader.Run] MineBlock error=%v", err)
				continue
			}
			blockMessage, err := blockChain.NewBlockMessage(block, abandonedP)
			if err != nil {
				logger.Warningf("[Leader.Run] NewBlockMessage error=%v", err)
				continue
			}
			jsonData, err := json.Marshal(blockMessage)
			if err != nil {
				logger.Warningf("[Leader.Run] json.Marshal error=%v", err)
				continue
			}
			service.Net.BroadCast(jsonData, service.BlockMsg)
			l.MessagePool.Clear(blockChain.BlockMaxSize)
			l.Mutex.Lock()
			l.BlockConfirm = false
			l.Mutex.Unlock()
		}
	}
}

func (l *Leader) Confirm() {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	l.BlockConfirm = true
}

func CheckProposals(proposals ProposalMessages) (ProposalMessages, ProposalMessages) {
	filter := make(map[string]ProposalMessages)
	abandoneP := ProposalMessagePool{}
	validP := ProposalMessagePool{}
	for _, p := range proposals {
		if fp, ok := filter[p.ZoneName]; !ok {
			filter[p.ZoneName] = append(filter[p.ZoneName], p)
			validP.AddProposal(p)
		} else {
			drop := false
			for _, tmpP := range filter[p.ZoneName] {
				if reflect.DeepEqual(p.Id, tmpP.Id) {
					drop = true
					break
				}
			}
			if !drop {
				//TODO: Two conflicted proposal
				tmpP := fp[len(fp)-1]
				switch p.Type {
				case Add:
					if tmpP.Owner != messages.Dereliction {
						abandoneP.AddProposal(p)
					} else {
						validP.AddProposal(p)
					}
				case Mod:
					if tmpP.Owner != p.Owner || tmpP.Owner != p.From {
						abandoneP.AddProposal(p)
					} else {
						validP.AddProposal(p)
					}
				case Del:
					if p.Owner != messages.Dereliction || tmpP.Owner != p.From {
						abandoneP.AddProposal(p)
					} else {
						validP.AddProposal(p)
					}
				}
			}
		}
	}
	return validP.ProposalMessages, abandoneP.ProposalMessages
}

func ValidateProposals(msg *blockChain.BlockMessage) bool {
	tmpPool := ProposalMessages{}
	tmpPool = append(tmpPool, msg.ProposalMessages...)
	tmpPool = append(tmpPool, msg.AbandonedProposal...)
	validP, _ := CheckProposals(tmpPool)
	return reflect.DeepEqual(validP, msg.ProposalMessages)
}

package service

import (
	"BCDns_0.1/blockChain"
	"BCDns_0.1/messages"
	"BCDns_0.1/network/service"
	"encoding/json"
	"fmt"
	"github.com/golang/snappy"
	"reflect"
	"time"
)

var (
	ProposalMessageChan chan messages.ProposalMessage
	BlockConfirmChan    chan uint
)

type Leader struct {
	MessagePool  messages.ProposalMessagePool
	BlockConfirm bool
	UnConfirmedH uint
}

func init() {
	ProposalMessageChan = make(chan messages.ProposalMessage, 1024)
	BlockConfirmChan = make(chan uint, 1024)
}

func NewLeader() *Leader {
	return &Leader{
		MessagePool:  messages.NewProposalMessagePool(),
		BlockConfirm: true,
		UnConfirmedH: 0,
	}
}

func (l *Leader) Run(done chan uint) {
	defer close(done)
	interrupt := make(chan int)
	go func() {
		for {
			select {
			case <-time.After(10 * time.Second):
				fmt.Println("Timeout", l.BlockConfirm, l.UnConfirmedH)
				if l.BlockConfirm {
					interrupt <- 1
				}
			}
		}
	}()
	for {
		select {
		case <-interrupt:
			l.generateBlock()
		case h := <-BlockConfirmChan:
			if h == l.UnConfirmedH {
				l.BlockConfirm = true
			}
		case msg := <-ProposalMessageChan:
			l.MessagePool.AddProposal(msg)
			if l.BlockConfirm && l.MessagePool.Size() >= blockChain.BlockMaxSize {
				l.generateBlock()
			}
		}
	}
}

func (l *Leader) generateBlock() {
	if !service.ViewManager.IsLeader() {
		return
	}
	if l.MessagePool.Size() <= 0 {
		fmt.Printf("[Leader.Run] CurrentBlock is empty\n")
		return
	}
	bound := blockChain.BlockMaxSize
	if len(l.MessagePool.ProposalMessages) < blockChain.BlockMaxSize {
		bound = len(l.MessagePool.ProposalMessages)
	}
	validP, abandonedP := CheckProposals(l.MessagePool.ProposalMessages[:bound])
	block, err := blockChain.BlockChain.MineBlock(validP)
	if err != nil {
		logger.Warningf("[Leader.Run] MineBlock error=%v", err)
		return
	}
	blockMessage, err := blockChain.NewBlockMessage(block, abandonedP)
	if err != nil {
		logger.Warningf("[Leader.Run] NewBlockMessage error=%v", err)
		return
	}
	jsonData, err := json.Marshal(blockMessage)
	if err != nil {
		logger.Warningf("[Leader.Run] json.Marshal error=%v", err)
		return
	}
	service.Net.BroadCast(jsonData, service.BlockMsg)
	l.MessagePool.Clear(bound)
	l.BlockConfirm = false
	l.UnConfirmedH = block.Height
	dc := snappy.Encode(nil, jsonData)
	fmt.Println("block broadcast fin", len(dc), block.Height, len(validP), validP[len(validP) - 1].Values)
}

func CheckProposals(proposals messages.ProposalMessages) (
	messages.ProposalMessages, messages.ProposalMessages) {
	filter := make(map[string]messages.ProposalMessages)
	abandoneP := messages.ProposalMessagePool{}
	validP := messages.ProposalMessagePool{}
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
				case messages.Add:
					if tmpP.Owner != messages.Dereliction {
						abandoneP.AddProposal(p)
					} else {
						validP.AddProposal(p)
					}
				case messages.Mod:
					if tmpP.Owner != p.Owner || tmpP.Owner != p.From {
						abandoneP.AddProposal(p)
					} else {
						validP.AddProposal(p)
					}
				case messages.Del:
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
	tmpPool := messages.ProposalMessages{}
	tmpPool = append(tmpPool, msg.ProposalMessages...)
	tmpPool = append(tmpPool, msg.AbandonedProposal...)
	validP, _ := CheckProposals(tmpPool)
	return reflect.DeepEqual(validP, msg.ProposalMessages)
}

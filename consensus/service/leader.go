package service

import (
	"BCDns_0.1/blockChain"
	service2 "BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/messages"
	"BCDns_0.1/network/service"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

var (
	LeaderNode *LeaderNodeT
)

type LeaderNodeT struct {
}

type LeaderNodeInterface interface {
	Run()
}

func NewLeaderNode() *LeaderNodeT {
	return &LeaderNodeT{}
}

func (l *LeaderNodeT) Run(done chan uint) {
	defer close(done)
	interrupt := make(chan int)
	interruptTimer := make(chan int)
	go func() {
		for true {
			select {
			case <-time.After(10 * time.Second):
				interrupt <- 1
			case <-interruptTimer:
			}
		}
	}()
	for true {
		select {
		case msgByte := <-service.CommitChan:
			if service.Leader.IsLeader() {
				var msg messages.AuditedProposal
				err := json.Unmarshal(msgByte, &msg)
				if err != nil {
					fmt.Printf("[LeaderNode] json.Unmarshal failed err=%v\n", err)
					continue
				}
				if !service.Leader.CheckTermId(msg.TermId) {
					fmt.Printf("[LeaderNode] termId is invalid\n")
					continue
				}
				if !service2.CertificateAuthorityX509.Exits(msg.From) {
					fmt.Printf("[LeaderNode] From node is unexist\n")
					continue
				}
				if !msg.VerifySignature() {
					fmt.Printf("[LeaderNode] Signature is illegal\n")
					continue
				}
				if !msg.VerifySignatures() {
					fmt.Printf("[LeaderNode] Signatures is illegal\n")
					continue
				}
				blockChain.LeaderAuditedProposalPool.AddProposal(msg)
				if blockChain.LeaderAuditedProposalPool.Len() >= 100 {
					interrupt <- 1
					interruptTimer <- 1
				}
			}
		case <-interrupt:
			if service.Leader.IsLeader() {
				if blockChain.LeaderAuditedProposalPool.Len() <= 0 {
					fmt.Printf("[LeaderNode] CurrentBlock is empty\n")
					continue
				}
				//validP, abandonedP := CheckProposal(*blockChain.LeaderAuditedProposalPool)
				//fmt.Println(validP)
				//fmt.Println(abandonedP)
				b, err := blockChain.BlockChain.MineBlock(blockChain.LeaderAuditedProposalPool.AuditedProposalSlice)
				if err != nil {
					fmt.Printf("[LeaderNode] MineBlock err=%v\n", err)
					continue
				}
				//b := blockChain.NewBlock(blockChain.LeaderAuditedProposalPool.AuditedProposalSlice, []byte("test"), 0)
				blockMessage, err := blockChain.NewBlockMessage(b)
				if err != nil {
					fmt.Printf("[LeaderNode] NewBlockMessage failed err=%v\n", err)
					continue
				}
				blockBytes, err := json.Marshal(*blockMessage)
				if err != nil {
					fmt.Printf("[LeaderNode] CurrentBlock marshal failed err=%v\n", err)
					continue
				}
				service.P2PNet.BroadcastMsg(blockBytes, service.Block)
				blockChain.LeaderAuditedProposalPool.Clear()
			}
		}
	}
}

func CheckProposal(proposals messages.AuditedProposalPool) (messages.AuditedProposalSlice,
	messages.AuditedProposalSlice) {
	filter := make(map[string]messages.AuditedProposalSlice)
	abandoneP := messages.AuditedProposalPool{}
	validP := messages.AuditedProposalPool{}
	for _, p := range proposals.AuditedProposalSlice {
		if fp, ok := filter[p.Proposal.Body.ZoneName]; !ok {
			filter[p.Proposal.Body.ZoneName] = append(filter[p.Proposal.Body.ZoneName], p)
			validP.AddProposal(p)
		} else {
			drop := false
			for _, tmpP := range filter[p.Proposal.Body.ZoneName] {
				if reflect.DeepEqual(p.Proposal.Body.Id, tmpP.Proposal.Body.Id) {
					drop = true
					break
				}
			}
			if !drop {
				//TODO: Two conflicted proposal
				tmpP := fp[len(fp)-1]
				switch p.Proposal.Body.Type {
				case messages.Add:
					if tmpP.Proposal.Body.Owner != messages.Dereliction {
						abandoneP.AddProposal(p)
					} else {
						validP.AddProposal(p)
					}
				case messages.Mod:
					if tmpP.Proposal.Body.Owner != p.Proposal.Body.Owner || tmpP.Proposal.Body.Owner != p.Proposal.Body.PId.Name {
						abandoneP.AddProposal(p)
					} else {
						validP.AddProposal(p)
					}
				case messages.Del:
					if p.Proposal.Body.Owner != messages.Dereliction || tmpP.Proposal.Body.Owner != p.Proposal.Body.PId.Name {
						abandoneP.AddProposal(p)
					} else {
						validP.AddProposal(p)
					}
				}
			}
		}
	}
	return validP.AuditedProposalSlice, abandoneP.AuditedProposalSlice
}

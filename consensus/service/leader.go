package service

import (
	"BCDns_0.1/blockChain"
	service2 "BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/messages"
	"BCDns_0.1/network/service"
	"encoding/json"
	"fmt"
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

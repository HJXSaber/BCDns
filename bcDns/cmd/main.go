package main

import (
	"BCDns_0.1/bcDns/conf"
	blockChain2 "BCDns_0.1/blockChain"
	service2 "BCDns_0.1/consensus/service"
	dao2 "BCDns_0.1/dao"
	service3 "BCDns_0.1/network/service"
	"fmt"
	"time"
)

func main() {
	var err error
	//fmt.Println("[Init Certificate]")
	//service.CertificateAuthorityX509 = new(service.CAX509)
	fmt.Println("[Init Dao]")
	dao2.Dao, err = NewDao()
	if err != nil {
		panic(err)
	}
	defer blockChain2.BlockChain.Close()
	fmt.Println("[Init NetWork]")
	service3.P2PNet = service3.NewDnsNet()
	if service3.P2PNet == nil {
		panic("NewDnsNet failed")
	}
	fmt.Println("[Init Leader]")
	service3.Leader = service3.NewLeader()
	if service3.Leader == nil {
		panic("NewLeader failed")
	}
	fmt.Println("[Init Proposer]")
	service2.Proposer = service2.NewProposer(15 * time.Second)
	if service2.Proposer == nil {
		panic("NewProposer failed")
	}
	fmt.Println("[Init Node]")
	service2.Node = service2.NewNode()
	if service2.Node == nil {
		panic("NewNode failed")
	}
	fmt.Println("[Init LeaderNode]")
	service2.LeaderNode = service2.NewLeaderNode()
	if service2.LeaderNode == nil {
		panic("NewLeaderNode failed")
	}
	fmt.Println("[Run]")
	done := make(chan uint)
	go service2.Proposer.Run(done)
	go service2.Node.Run(done)
	go service2.LeaderNode.Run(done)
	_ = <-done
	fmt.Println("[Err] System exit")
}

func NewDao() (*dao2.DAO, error) {
	var err error
	blockChain2.BlockChain, err = blockChain2.NewBlockchain(conf.BCDnsConfig.HostName)
	if err != nil {
		return nil, err
	}
	dao2.Db, err = dao2.NewDB(conf.BCDnsConfig.HostName)
	if err != nil {
		return nil, err
	}
	storage := dao2.NewStorage(dao2.Db, blockChain2.BlockChain)
	return &dao2.DAO{
		Storage: storage,
	}, nil
}

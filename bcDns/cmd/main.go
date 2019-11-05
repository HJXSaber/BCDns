package main

import (
	"BCDns_0.1/bcDns/conf"
	blockChain2 "BCDns_0.1/blockChain"
	"BCDns_0.1/certificateAuthority/service"
	service2 "BCDns_0.1/consensus/service"
	dao2 "BCDns_0.1/dao"
	service3 "BCDns_0.1/network/service"
	"time"
)

func main() {
	var err error
	service.CertificateAuthorityX509 = new(service.CAX509)
	dao2.Dao, err = NewDao()
	if err != nil {
		panic(err)
	}
	service3.P2PNet = service3.NewDnsNet()
	if service3.P2PNet == nil {
		panic("NewDnsNet failed")
	}
	service2.Proposer = service2.NewProposer(10 * time.Second)
	if service2.Proposer == nil {
		panic("NewProposer failed")
	}
	service2.Node = service2.NewNode()
	if service2.Node == nil {
		panic("NewNode failed")
	}

}

func NewDao() (*dao2.DAO, error) {
	blockChain, err := blockChain2.NewBlockchain(conf.BCDnsConfig.HostName)
	if err != nil {
		return nil, err
	}
	db, err := dao2.NewDB(conf.BCDnsConfig.HostName)
	if err != nil {
		return nil, err
	}
	storage := dao2.NewStorage(db, blockChain)
	return &dao2.DAO{
		Storage: storage,
	}, nil
}

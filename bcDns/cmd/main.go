package main

import (
	"BCDns_0.1/bcDns/conf"
	blockChain2 "BCDns_0.1/blockChain"
	service2 "BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/consensusMy/service"
	dao2 "BCDns_0.1/dao"
	service3 "BCDns_0.1/network/service"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"
)

func main() {
	go func() {
		http.HandleFunc("/debug/pprof/block", pprof.Index)
		http.HandleFunc("/debug/pprof/goroutine", pprof.Index)
		http.HandleFunc("/debug/pprof/heap", pprof.Index)
		http.HandleFunc("/debug/pprof/threadcreate", pprof.Index)

		http.ListenAndServe("0.0.0.0:8888", nil)
	}()
	var err error
	fmt.Println("[Init Dao]")
	dao2.Dao, err = NewDao()
	if err != nil {
		panic(err)
	}
	defer blockChain2.BlockChain.Close()
	fmt.Println("[Init NetWork]")
	service3.Net, err = service3.NewDNet()
	if err != nil {
		panic(err)
	}
	if service3.Net == nil {
		panic("NewDNet failed")
	}
	fmt.Println("[Init Leader]")
	service3.ViewManager, err = service3.NewViewManager()
	if err != nil {
		panic(err)
	}
	fmt.Println("[Join]")
	err = service3.Net.Join(service2.CertificateAuthorityX509.GetSeeds())
	if err != nil {
		panic(err)
	}
	fmt.Println("[Leader.Start]")
	service3.ViewManager.Start()
	fmt.Println("[Init Consensus]")
	done := make(chan uint)
	service.ConsensusCenter, err = service.NewConsensus(done)
	if err != nil {
		panic(err)
	}
	fmt.Println("[System running]")
	fmt.Println("[Start Time]", time.Now())
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

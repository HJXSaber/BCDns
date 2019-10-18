package service

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/certificateAuthority/service"
	"fmt"
	"github.com/hashicorp/memberlist"
	"log"
)

type DnsNet struct {
	Network *memberlist.Memberlist
	broadCasts *memberlist.TransmitLimitedQueue
}

//Can not broadcast msg whose size is longer than 1350B
//When the size of msg is longer than 1350B. We have to transfer it by reliable channel
func (net DnsNet) BroadcastMsg(jsonData []byte) {
	if len(jsonData) >= 1350 {
		//TODO
		for _, node := range net.Network.Members() {
			err := net.Network.SendReliable(node, jsonData)
			if err != nil {
				fmt.Println("Broadcast msg failed", err)
				continue
			}
		}
	} else {
		net.broadCasts.QueueBroadcast(&Broadcast{
			Msg: jsonData,
			Notify:nil,
		})
	}
}

var (
	P2PNet DnsNet
)

func init() {
	config := memberlist.DefaultLANConfig()
	config.BindPort = conf.BCDnsConfig.Port
	config.Delegate = &Delegate{}
	config.Name = conf.BCDnsConfig.HostName

	var err error
	P2PNet.Network, err = memberlist.Create(config)
	if err != nil {
		//TODO
		log.Fatal("Initial network failed", err)
	}

	seeds := service.CertificateAuthorityX509.GetSeeds()
	_, err = P2PNet.Network.Join(seeds)
	for _, member := range P2PNet.Network.Members() {
		for i, cert := range service.CertificateAuthorityX509.CertificatesOrder {
			if cert.Cert.IPAddresses[0].Equal(member.Addr) {
				service.CertificateAuthorityX509.CertificatesOrder[i].Member = member
			}
		}
	}
	if err != nil {
		//TODO
		log.Fatal("Join failed ", err)
	}

	P2PNet.broadCasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return P2PNet.Network.NumMembers()
		},
		RetransmitMult: 3,
	}
}

type Broadcast struct {
	Msg    []byte
	Notify chan<- struct{}
}

func (b *Broadcast) Finished() {
	if b.Notify != nil {
		close(b.Notify)
	}
}

func (*Broadcast) Invalidates(b memberlist.Broadcast) bool {
	return false
}

func (b *Broadcast) Message() []byte {
	return b.Msg
}

type Delegate struct {}

func (*Delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

func (*Delegate) NotifyMsg(data []byte) {

}

func (*Delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return P2PNet.broadCasts.GetBroadcasts(overhead, limit)
}

//exchange local data with remote peer. certificate verify through this func
func (*Delegate) LocalState(join bool) []byte {
	_, certBytes := service.CertificateAuthorityX509.GetLocalCertificate()
	if certBytes == nil {
		return nil
	}
	return certBytes
}

func (*Delegate) MergeRemoteState(buf []byte, join bool) {
	if !join {
		fmt.Println("MergeState TODO")
	}
}
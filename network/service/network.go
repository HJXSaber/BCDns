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
	Leader memberlist.Node
}

//Can not broadcast msg whose size is longer than 1350B
//When the size of msg os longer than 1350B. We have to transfer it by reliable channel
func (net DnsNet) BroadcastMsg(jsonData []byte) {
	if len(jsonData) >= 1350 {
		//TODO
	}
	net.broadCasts.QueueBroadcast(&Broadcast{
		Msg: jsonData,
		Notify:nil,
	})
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
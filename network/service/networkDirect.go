package service

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"github.com/sasha-s/go-deadlock"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	logger          *logging.Logger // package-level logger
	ListenAddr      = "0.0.0.0"
	MaxPacketLength = 65536
	Net             *DNet

	ChanSize            = 1024
	ProposalChan        chan []byte
	BlockChan           chan []byte
	BlockConfirmChan    chan []byte
	DataSyncChan        chan []byte
	DataSyncRespChan    chan []byte
	ProposalReplyChan   chan []byte
	ProposalConfirmChan chan []byte
	ViewChangeChan      chan []byte
	NewViewChan         chan []byte
	InitLeaderChan      chan []byte
	JoinReplyChan       chan JoinReplyMessage
	JoinChan chan JoinMessage
)

func init() {
	logger = logging.MustGetLogger("networks")
	ProposalChan = make(chan []byte, ChanSize)
	BlockChan = make(chan []byte, ChanSize)
	BlockConfirmChan = make(chan []byte, ChanSize)
	DataSyncChan = make(chan []byte, ChanSize)
	DataSyncRespChan = make(chan []byte, ChanSize)
	ProposalReplyChan = make(chan []byte, ChanSize)
	ProposalConfirmChan = make(chan []byte, ChanSize)
	ViewChangeChan = make(chan []byte, ChanSize)
	NewViewChan = make(chan []byte, ChanSize)
	InitLeaderChan = make(chan []byte, ChanSize)
	JoinReplyChan = make(chan JoinReplyMessage, ChanSize)
	JoinChan = make(chan JoinMessage, ChanSize)
}

type MessageTypeT uint8

const (
	ProposalMsg MessageTypeT = iota + 1
	BlockMsg
	BlockConfirmMsg
	DataSyncMsg
	DataSyncRespMsg
	ProposalReplyMsg
	ProposalConfirmMsg
	InitLeaderMsg
	ViewChangeMsg
	NewViewMsg
	JoinMsg
)

const (
	Packed = iota
	Packing
)

type DNet struct {
	Mutex   deadlock.Mutex
	Members []DNode
	Map     map[string]DNode
	Node    DNode
}

type DNode struct {
	Pass       bool
	RemoteAddr string
	Name       string
	NodeId     int64
	Conn       net.Conn
}

func NewDNet() (*DNet, error) {
	dNet := &DNet{
		Members: []DNode{},
		Map:     map[string]DNode{},
	}
	go dNet.handleStram()
	return dNet, nil
}

func (n *DNet) handleStram() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4",
		strings.Join([]string{ListenAddr, conf.BCDnsConfig.Port}, ":"))
	if err != nil {
		panic(err)
	}
	tcpListen, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := tcpListen.Accept()
		if err != nil {
			logger.Warningf("[Network] NewDNet listen.accept error=%v", err)
			continue
		}
		go n.handleConn(conn)
	}
}

func (n *DNet) handleConn(conn net.Conn) {
	var (
		msg Message
		header *PacketHeader
		length = 0
		data = make([]byte, 0)
	)
	state := Packed
	for {
		dataBuf := make([]byte, MaxPacketLength)
		l, err := conn.Read(dataBuf)
		if err != nil {
			if err == io.EOF {
				logger.Warningf("[Network] handleConn closed", conn.RemoteAddr())
				return
			}
			logger.Warningf("[Network] handleConn Read error=%v", err)
			continue
		}
		length += l
		data = append(data, dataBuf[:l]...)
		if state == Packed {
			header, err = GetPacketHeader(data)
			if err != nil {
				length, data = 0, data[:0]
				logger.Warningf("[NetWork] handleConn GetPacketHeader error=%v", err)
				continue
			}
		}
		if length < header.Len {
			state = Packing
			continue
		} else {
			state = Packed
			msg, err = UnpackMessage(data[:header.Len])
			data = data[header.Len:]
			length -= header.Len
			if err != nil {
				logger.Warningf("[NetWork] handleConn UnpackMessage error=%v", err)
				continue
			}
		}
		switch msg.MessageTypeT {
		case JoinMsg:
			var message JoinMessage
			err := json.Unmarshal(msg.Payload, &message)
			if err != nil {
				logger.Warningf("[Network] handleConn json.Unmarshal error=%v", err)
				continue
			}
			if !message.VerifySignature() {
				logger.Warningf("[Network] handleConn signature is invalid")
				continue
			}
			if !service.CertificateAuthorityX509.VerifyCertificate(message.Cert) {
				logger.Warningf("[Network] handleConn cert is invalid")
				continue
			}
			node := DNode{
				Pass:       true,
				RemoteAddr: conn.RemoteAddr().String(),
				Name:       message.From,
				NodeId:     message.NodeId,
				Conn:       conn,
			}
			n.Members = append(n.Members, node)
			n.Mutex.Lock()
			n.Map[node.Name] = node
			n.Mutex.Unlock()
			replyMsg, err := NewJoinReplyMessage(ViewManager.View, map[string][]byte{})
			if err != nil {
				logger.Warningf("[Network] handleConn NewJoinReplyMessage error=%v", err)
				continue
			}
			jsonData, err := json.Marshal(replyMsg)
			if err != nil {
				logger.Warningf("[Network] handleConn json.Marshal error=%v", err)
				continue
			}
			JoinChan <- message
			node.Send(jsonData)
		case ProposalMsg:
			ProposalChan <- msg.Payload
		case BlockMsg:
			fmt.Println("block")
			BlockChan <- msg.Payload
		case BlockConfirmMsg:
			fmt.Println("blockConfirm")
			BlockConfirmChan <- msg.Payload
		case DataSyncMsg:
			fmt.Println("dataSync")
			DataSyncChan <- msg.Payload
		case DataSyncRespMsg:
			fmt.Println("dataSyncReply")
			DataSyncRespChan <- msg.Payload
		case ProposalReplyMsg:
			ProposalReplyChan <- msg.Payload
		case ProposalConfirmMsg:
			fmt.Println("proposalConfirm")
			ProposalConfirmChan <- msg.Payload
		case InitLeaderMsg:
			fmt.Println("InitLeaderMsg")
			InitLeaderChan <- msg.Payload
		case ViewChangeMsg:
			fmt.Println("ViewChangeMsg")
			ViewChangeChan <- msg.Payload
		case NewViewMsg:
			fmt.Println("NewVIewMsg")
			NewViewChan <- msg.Payload
		default:
			logger.Warningf("[Network] handleConn Unknown message type")
		}
	}
}

// If non-node is reached, return error
func (n *DNet) Join(seeds []string) error {
	msg, err := NewJoinMessage()
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	localData, err := PackMessage(NewMessage(JoinMsg, jsonData))
	if err != nil {
		return err
	}
	success := int32(0)
	wg := sync.WaitGroup{}
	fmt.Println(seeds)
	for _, seed := range seeds {
		wg.Add(1)
		go func(seed string) {
			defer wg.Done()
			conn, err := JoinNode(seed)
			if err != nil {
				logger.Warningf("[Network] JoinNode error=%v", err)
				return
			}
			err = n.PushAndPull(conn, localData)
			if err != nil {
				logger.Warningf("[Network] Join push&pull error=%v", err)
				return
			}
			go n.handleConn(conn)
			atomic.AddInt32(&success, 1)
		}(seed)
	}
	wg.Wait()
	if success == 0 {
		return errors.New("[NetWork] Join failed")
	}
	return nil
}

func (n *DNet) PushAndPull(conn net.Conn, localData []byte) error {
	_, err := conn.Write(localData)
	if err != nil {
		return err
	}

	remoteData := make([]byte, MaxPacketLength)
	l, err := conn.Read(remoteData)
	if err != nil {
		return err
	}
	var msg JoinReplyMessage
	err = json.Unmarshal(remoteData[:l], &msg)
	if err != nil {
		return err
	}
	if !msg.VerifySignature() {
		return errors.New("[PushAndPull] JoinReplyMessage.VerifySignature failed")
	}
	if msg.From != conf.BCDnsConfig.HostName {
		node := DNode{
			Pass:       true,
			RemoteAddr: conn.RemoteAddr().String(),
			Name:       msg.From,
			NodeId:     msg.NodeId,
			Conn:       conn,
		}
		n.Members = append(n.Members, node)
		n.Mutex.Lock()
		defer n.Mutex.Unlock()
		n.Map[node.Name] = node
		JoinReplyChan <- msg
	}
	return nil
}

func JoinNode(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (n *DNet) BroadCast(payload []byte, t MessageTypeT) {
	data, err := PackMessage(NewMessage(t, payload))
	if err != nil {
		logger.Warningf("[Network] BroadCast json.Marshal error=%v", err)
		return
	}
	for _, m := range n.Members {
		_, err := m.Send(data)
		if err != nil {
			logger.Warningf("[Network] BroadCast send error=%v", err)
		}
	}
}

func (n *DNet) SendTo(payload []byte, t MessageTypeT, to string) {
	data, err := PackMessage(NewMessage(t, payload))
	if err != nil {
		logger.Warningf("[Network] BroadCast json.Marshal error=%v", err)
		return
	}
	n.Mutex.Lock()
	defer n.Mutex.Unlock()
	_, _ = n.Map[to].Send(data)
}

func (n *DNet) SendToLeader(payload []byte, t MessageTypeT) {
	data, err := PackMessage(NewMessage(t, payload))
	if err != nil {
		logger.Warningf("[Network] BroadCast json.Marshal error=%v", err)
		return
	}
	name, err := utils.GetCertId(*service.CertificateAuthorityX509.CertificatesOrder[ViewManager.LeaderId])
	if err != nil {
		logger.Warningf("[SendToLeader] GetCertId failed err=%v", err)
		return
	}
	n.Mutex.Lock()
	defer n.Mutex.Unlock()
	_, _ = n.Map[name].Send(data)
}

func (n *DNet) GetAllNodeIds() []int64 {
	var ids []int64
	for _, node := range n.Members {
		ids = append(ids, node.NodeId)
	}
	return ids
}

func (n DNode) Send(msg []byte) (int, error) {
	return n.Conn.Write(msg)
}
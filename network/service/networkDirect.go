package service

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/op/go-logging"
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

type DNet struct {
	Mutex   sync.Mutex
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
	var msg Message
	for {
		dataBuf := make([]byte, MaxPacketLength)
		l, err := conn.Read(dataBuf)
		if err != nil {
			logger.Warningf("[Network] handleConn Read error=%v", err)
			continue
		}
		err = json.Unmarshal(dataBuf[:l], &msg)
		if err != nil {
			logger.Warningf("[Network] handleConn json.Unmarshal error=%v", err)
			continue
		}
		fmt.Println("start handle")
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
			_, _ = node.Send(jsonData)
		case ProposalMsg:
			ProposalChan <- msg.Payload
		case BlockMsg:
			BlockChan <- msg.Payload
		case BlockConfirmMsg:
			BlockConfirmChan <- msg.Payload
		case DataSyncMsg:
			DataSyncChan <- msg.Payload
		case DataSyncRespMsg:
			DataSyncRespChan <- msg.Payload
		case ProposalReplyMsg:
			ProposalReplyChan <- msg.Payload
		case ProposalConfirmMsg:
			ProposalConfirmChan <- msg.Payload
		case InitLeaderMsg:
			InitLeaderChan <- msg.Payload
		case ViewChangeMsg:
			ViewChangeChan <- msg.Payload
		default:
			logger.Warningf("[Network] handleConn Unknown message type")
		}
		fmt.Println("finish handle")
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
	localData, err := json.Marshal(NewMessage(JoinMsg, jsonData))
	if err != nil {
		return err
	}
	success := int32(0)
	wg := sync.WaitGroup{}
	for _, seed := range seeds {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := JoinNode(seed)
			if err != nil {
				logger.Warningf("[Network] JoinNode error=%v", err)
				return
			}
			joinSelf, err := n.PushAndPull(conn, localData)
			if err != nil {
				logger.Warningf("[Network] Join push&pull error=%v", err)
				return
			}
			atomic.AddInt32(&success, 1)
			if !joinSelf {
				go n.handleConn(conn)
			}
		}()
	}
	wg.Wait()
	if success == 0 {
		return errors.New("[NetWork] Join failed")
	}
	return nil
}

func (n *DNet) PushAndPull(conn net.Conn, localData []byte) (bool, error) {
	_, err := conn.Write(localData)
	if err != nil {
		return false, err
	}

	remoteData := make([]byte, MaxPacketLength)
	l, err := conn.Read(remoteData)
	if err != nil {
		return false, err
	}
	var msg JoinReplyMessage
	err = json.Unmarshal(remoteData[:l], &msg)
	if err != nil {
		return false, err
	}
	if !msg.VerifySignature() {
		return false, errors.New("[PushAndPull] JoinReplyMessage.VerifySignature failed")
	}
	JoinReplyChan <- msg
	if _, ok := n.Map[msg.From]; !ok {
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
		return true, nil
	}
	return false, nil
}

func JoinNode(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (n *DNet) BroadCast(payload []byte, t MessageTypeT) {
	data, err := json.Marshal(NewMessage(t, payload))
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
	data, err := json.Marshal(NewMessage(t, payload))
	if err != nil {
		logger.Warningf("[Network] BroadCast json.Marshal error=%v", err)
		return
	}
	n.Mutex.Lock()
	defer n.Mutex.Unlock()
	_, _ = n.Map[to].Send(data)
}

func (n *DNet) SendToLeader(payload []byte, t MessageTypeT) {
	data, err := json.Marshal(NewMessage(t, payload))
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

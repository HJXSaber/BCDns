package service

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/utils"
	"encoding/json"
	"errors"
	"github.com/op/go-logging"
	"net"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	logger          *logging.Logger // package-level logger
	ListenAddr      = "0,0,0,0"
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
	JoinReplyChan chan JoinReplyMessage
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
	ViewChangeMsg
	ViewChangeResult
	RetrieveLeader
	RetrieveLeaderResponse
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
	conn       net.Conn
}

func NewDNet() *DNet {
	dNet := new(DNet)
	go dNet.handleStram()
	return dNet
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
		switch message := msg.(type) {
		case JoinMessage:
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
				conn:       conn,
			}
			n.Members = append(n.Members, node)
			n.Mutex.Lock()
			n.Map[node.Name] = node
			n.Mutex.Unlock()
			replyMsg, err := NewJoinReplyMessage(ViewManager.View, ViewManager.Proof)
			if err != nil {
				logger.Warningf("[Network] handleConn NewJoinReplyMessage error=%v", err)
				continue
			}
			jsonData, err := json.Marshal(replyMsg)
			if err != nil {
				logger.Warningf("[Network] handleConn json.Marshal error=%v", err)
				continue
			}
			node.Send(jsonData)
		case MessageProposal:
			ProposalChan <- message.Payload
		case MessageBlock:
			BlockChan <- message.Payload
		case MessageBlockConfirm:
			BlockConfirmChan <- message.Payload
		case MessageDataSync:
			DataSyncChan <- message.Payload
		case MessageDataSyncResp:
			DataSyncRespChan <- message.Payload
		case MessageProposalReply:
			ProposalReplyChan <- message.Payload
		case MessageViewChange:
			ViewChangeChan <- message.Payload
		case MessageRetrieveLeader:
			RetrieveLeaderMsgChan <- message.Payload
		case MessageProposalConfirm:
			ProposalConfirmChan <- message.Payload
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
	localData, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	success := int32(0)
	wg := sync.WaitGroup{}
	for _, seed := range seeds {
		go func() {
			wg.Add(1)
			defer wg.Done()
			conn, err := JoinNode(seed)
			if err != nil {
				return
			}
			err = n.PushAndPull(conn, localData)
			if err != nil {
				logger.Warningf("[Network] Join push&pull error=%v", err)
				return
			}
			atomic.AddInt32(&success, 1)
			go n.handleConn(conn)
		}()
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
	node := DNode{
		Pass:       true,
		RemoteAddr: conn.RemoteAddr().String(),
		Name:       msg.From,
		conn:       conn,
	}
	n.Members = append(n.Members, node)
	n.Mutex.Lock()
	defer n.Mutex.Unlock()
	n.Map[node.Name] = node
	JoinReplyChan <- msg
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
	msg, err := ConvertMessage(payload, t)
	if err != nil {
		logger.Warningf("[Network] BroadCast ConvertMessage error=%v", err)
		return
	}
	data, err := json.Marshal(msg)
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
	msg, err := ConvertMessage(payload, t)
	if err != nil {
		logger.Warningf("[Network] BroadCast ConvertMessage error=%v", err)
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Warningf("[Network] BroadCast json.Marshal error=%v", err)
		return
	}
	n.Mutex.Lock()
	defer n.Mutex.Unlock()
	_, _ = n.Map[to].Send(data)
}

func (n *DNet) SendToLeader(payload []byte, t MessageTypeT) {
	msg, err := ConvertMessage(payload, t)
	if err != nil {
		logger.Warningf("[Network] BroadCast ConvertMessage error=%v", err)
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Warningf("[Network] BroadCast json.Marshal error=%v", err)
		return
	}
	name, err := utils.GetCertId(*service.CertificateAuthorityX509.CertificatesOrder[Leader.LeaderId])
	if err != nil {
		logger.Warningf("[SendToLeader] GetCertId failed err=%v", err)
		return
	}
	n.Mutex.Lock()
	defer n.Mutex.Unlock()
	_, _ = n.Map[name].Send(data)
}

func ConvertMessage(payload []byte, t MessageTypeT) (interface{}, error) {
	var msg Message
	switch t {
	case ProposalMsg:
		msg = MessageProposal{
			Payload: payload,
		}
	case BlockMsg:
		msg = MessageBlock{
			Payload: payload,
		}
	case BlockConfirmMsg:
		msg = MessageBlockConfirm{
			Payload: payload,
		}
	case DataSyncMsg:
		msg = MessageDataSync{
			Payload: payload,
		}
	case DataSyncRespMsg:
		msg = MessageDataSyncResp{
			Payload: payload,
		}
	case ProposalReplyMsg:
		msg = MessageProposalReply{
			Payload: payload,
		}
	case ViewChangeMsg:
		msg = MessageViewChange{
			Payload: payload,
		}
	case RetrieveLeader:
		msg = MessageRetrieveLeader{
			Payload: payload,
		}
	case ProposalConfirmMsg:
		msg = MessageProposalConfirm{
			Payload: payload,
		}
	default:
		return nil, errors.New("[ConvertMessage] Unknown messageType")
	}
	return msg, nil
}

func (n *DNode) Send(msg []byte) (int, error) {
	return n.conn.Write(msg)
}

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
)

func init() {
	logger = logging.MustGetLogger("networks")
}

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
		case MessageJoin:
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
				Name:       message.Name,
				conn:       conn,
			}
			n.Members = append(n.Members, node)
			n.Mutex.Lock()
			n.Map[node.Name] = node
			n.Mutex.Unlock()
		case MessageProposal:
			ProposalChan <- message.Payload
		case MessageEndorsement:
			EndorsementChan <- message.Payload
		case MessageCommit:
			CommitChan <- message.Payload
		case MessageBlock:
			BlockChan <- message.Payload
		case MessageViewChange:
			ViewChangeMsgChan <- message.Payload
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
	_, certData := service.CertificateAuthorityX509.GetLocalCertificate()
	msg := MessageJoin{
		Cert: certData,
		Name: conf.BCDnsConfig.HostName,
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
	var msg MessageJoin
	err = json.Unmarshal(remoteData[:l], &msg)
	if err != nil {
		return err
	}
	node := DNode{
		Pass:       true,
		RemoteAddr: conn.RemoteAddr().String(),
		Name:       msg.Name,
		conn:       conn,
	}
	n.Members = append(n.Members, node)
	n.Mutex.Lock()
	defer n.Mutex.Unlock()
	n.Map[node.Name] = node
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
	case ProposalMsgT:
		msg = MessageProposal{
			Payload: payload,
		}
	case Endorsement:
		msg = MessageEndorsement{
			Payload: payload,
		}
	case Commit:
		msg = MessageCommit{
			Payload: payload,
		}
	case Block:
		msg = MessageBlock{
			Payload: payload,
		}
	case ViewChange:
		msg = MessageViewChange{
			Payload: payload,
		}
	case RetrieveLeader:
		msg = MessageRetrieveLeader{
			Payload: payload,
		}
	case ProposalConfirmT:
		msg = MessageProposalConfirm{
			Payload:payload,
		}
	default:
		return nil, errors.New("[ConvertMessage] Unknown messageType")
	}
	return msg, nil
}

func (n *DNode) Send(msg []byte) (int, error) {
	return n.conn.Write(msg)
}

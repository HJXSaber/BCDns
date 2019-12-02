package service


var (
	LeaderG *LeaderT
)

type LeaderT struct {
	ProposalMessageChan chan []byte

}

func (l *LeaderT) Run() {
	for {
		select {
		case msgByte := <- l.ProposalMessageChan:

		}
	}
}
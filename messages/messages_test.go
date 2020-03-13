package messages

import (
	"BCDns_0.1/utils"
	"fmt"
	"testing"
	"time"
)

func TestProposalMessage_GetPow(t *testing.T) {
	tt := int64(0)
	for i := 0; i < 5; i++ {
		t1 := time.Now()
		msg := ProposalMessage{
			Base:utils.Base{
				From: "s1",
				TimeStamp:time.Now().Unix(),
			},
			Type:Add,
			ZoneName:"com.",
			Owner: "s1",
			Values: map[string]string{
				"A":"1.1.1.1",
			},
		}
		msg.GetPow()
		t2 := time.Now()
		tt += t2.Sub(t1).Milliseconds()
	}
	fmt.Println(float64(tt) / float64(5))
}
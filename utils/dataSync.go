package utils

import (
	"BCDns_0.1/messages"
	"BCDns_0.1/network/service"
	"encoding/json"
	"fmt"
)

func StartDataSync(lastH, h uint) {
	for i := lastH; i <= h; i++ {
		syncMsg, err := messages.NewDataSyncMessage(i)
		if err != nil {
			fmt.Printf("[DataSync] NewDataSyncMessage error=%v", err)
			continue
		}
		jsonData, err := json.Marshal(syncMsg)
		if err != nil {
			fmt.Printf("[DataSync] json.Marshal error=%v", err)
			continue
		}
		service.Net.BroadCast(jsonData, service.DataSyncMsg)
	}
}

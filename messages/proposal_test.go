package messages

import (
	"encoding/json"
	"fmt"
	"github.com/rs/xid"
	"log"
	"testing"
)

func TestUUID(t *testing.T) {
	id := xid.New()
	fmt.Println(id)
	data, err := json.Marshal(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(data)
	var id2 xid.ID
	err = json.Unmarshal(data, id2.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id2)
}
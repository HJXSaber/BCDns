package conf

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	CAPath string
	CAPort int64

	//system info
	Port     string
	HostName string

	ProposalBufferSize int
	ProposalTimeout    time.Duration

	LeaderMsgBufferSize int
	PowDifficult        int
}

var (
	path        string
	BCDnsConfig Config
)

func init() {
	if val, ok := os.LookupEnv("BCDNSConfFile"); ok {
		path = val
	} else {
		log.Fatal("System's config is not set")
	}
	dir, file := filepath.Split(path)
	viper.SetConfigName(file)
	viper.AddConfigPath(dir)
	viper.SetConfigType("json")
	err := viper.ReadInConfig()
	if err != nil {
		//TODO
		log.Fatal("Read system config failed", err)
	}

	BCDnsConfig.Port = viper.GetString("PORT")
	BCDnsConfig.HostName = viper.GetString("HOSTNAME")
	BCDnsConfig.ProposalBufferSize = 10000
	BCDnsConfig.ProposalTimeout = 100 * time.Second
}

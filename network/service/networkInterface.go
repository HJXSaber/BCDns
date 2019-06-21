package service

type NetWorkInterface interface {
	BroadcastMsg(jsonData []byte)
}
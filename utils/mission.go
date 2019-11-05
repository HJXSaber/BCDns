package utils

import (
	"github.com/golang/protobuf/ptypes/any"
	"reflect"
)

//mission begin

var (
	HandlerCert = reflect.TypeOf(*new(HandlerCertT))
)

type HandlerCertT func(cert any.Any)

type MissionTable struct {
}

type Mission struct {
	Data    interface{}
	Handler interface{}
}

type MissionInterface interface {
	Call()
}

func (m *Mission) Call() {
	switch reflect.TypeOf(m.Handler) {
	case HandlerCert:
		f := m.Handler.(HandlerCertT)
		f(m.Data.(any.Any))
	}
}

//mission end

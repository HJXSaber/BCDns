package dao

import (
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)

var (
	Dao DAO
)

type DAO struct {
	mutex sync.Mutex
	db    *leveldb.DB
}

type Record struct {
	ZoneName string
	HostName string
}

type DAOInterface interface {
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Add(key, value []byte) error
}

func init() {
	db, err := leveldb.OpenFile("db", nil)
	if err != nil {
		panic(err)
	}
	Dao = DAO{
		mutex: sync.Mutex{},
		db:    db,
	}
}

func (d *DAO) Get(key []byte) ([]byte, error) {
	return d.db.Get(key, nil)
}

func (d *DAO) Has(key []byte) (bool, error) {
	return d.db.Has(key, nil)
}

func (d *DAO) Add(key, value []byte) error {
	return d.db.Put(key, value, nil)
}

func (d *DAO) DelEX(key []byte) error {
	return d.db.Delete(key, nil)
}

package dao

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"sync"
)

var (
	Dao DAO
)

type DAO struct {
	mutex sync.Mutex
	db *leveldb.DB
}

type DAOInterface interface {
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Add(key, value []byte) error
}

func init() {
	db, err := leveldb.OpenFile("db", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	Dao = DAO{
		mutex: sync.Mutex{},
		db: db,
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

func test() {
	//d, _ := leveldb.OpenFile("db", nil)

}
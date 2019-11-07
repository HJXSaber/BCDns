package dao

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"reflect"
)

var (
	Dao *DAO
)

type DAO struct {
	// key:zonename value:proposalMessage
	*Storage
}

type Record struct {
	ZoneName string
	HostName string
}

type DAOInterface interface {
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) error
}

const cacheFile = "blockchain_cache_%s.db"

type DB struct {
	*leveldb.DB
}

func NewDB(nodeId string) (*DB, error) {
	dbFile := fmt.Sprintf(cacheFile, nodeId)
	db, err := leveldb.OpenFile(dbFile, nil)
	if err != nil {
		return nil, err
	}
	return &DB{
		DB: db,
	}, nil
}

func (d *DB) Get(key []byte) ([]byte, error) {
	return d.DB.Get(key, nil)
}

func (d *DB) Set(key, value []byte) error {
	return d.DB.Put(key, value, nil)
}

type Storage struct {
	Storages []DAOInterface
}

func NewStorage(storage ...DAOInterface) *Storage {
	storages := new(Storage)
	for _, s := range storage {
		storages.Storages = append(storages.Storages, s)
	}
	return storages
}

func (s *Storage) GetZoneName(name string) ([]byte, error) {
	return s.GetZoneNameByIndex(name, 0)
}

var (
	ErrNotFoundT = reflect.TypeOf(leveldb.ErrNotFound)
)

func (s *Storage) GetZoneNameByIndex(name string, index int) ([]byte, error) {
	if data, err := s.Storages[index].Get([]byte(name)); err == leveldb.ErrNotFound {
		if index == len(s.Storages)-1 {
			return nil, err
		}
		data, err := s.GetZoneNameByIndex(name, index+1)
		if err == leveldb.ErrNotFound {
			_ = s.Storages[index].Set([]byte(name), data)
			return nil, err
		} else if err != nil {
			return nil, err
		} else {
			_ = s.Storages[index].Set([]byte(name), data)
			return data, nil
		}
	} else if err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

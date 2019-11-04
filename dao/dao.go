package dao

import (
	"github.com/syndtr/goleveldb/leveldb"
	"reflect"
)

var (
	Dao DAO
)

type DAO struct {
	// key:zonename value:hostname
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

func init() {
	db, err := leveldb.OpenFile("db", nil)
	if err != nil {
		panic(err)
	}
	Dao = DAO{
		db: db,
	}
}

func (d *DAO) Get(key []byte) ([]byte, error) {
	return d.db.Get(key, nil)
}

func (d *DAO) Set(key, value []byte) error {
	return d.db.Put(key, value, nil)
}

func (d *DAO) Has(key []byte) (bool, error) {
	return d.db.Has(key, nil)
}

func (d *DAO) DelEX(key []byte) error {
	return d.db.Delete(key, nil)
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

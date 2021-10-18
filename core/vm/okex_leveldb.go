package vm

import (
	"path/filepath"

	"github.com/ethereum/go-ethereum/ethdb/leveldb"
)

func init() {
	dbCreator := func(name string, dir string) (OKDB, error) {
		return NewGoLevelDB(name, dir)
	}
	registerDBCreator(GoLevelDBBackend, dbCreator, false)
}

type GoLevelDB struct {
	db *leveldb.Database
}

var _ OKDB = (*GoLevelDB)(nil)

func NewGoLevelDB(name string, dir string) (*GoLevelDB, error) {
	dbPath := filepath.Join(dir, name)
	db, err := leveldb.New(dbPath, 128, 128, "", false)
	if err != nil {
		return nil, err
	}
	database := &GoLevelDB{
		db: db,
	}
	return database, nil
}

// Get implements OKDB.
func (db *GoLevelDB) Get(key []byte) ([]byte, error) {
	key = nonNilBytes(key)
	return db.db.Get(key)
}

// Put implements OKDB.
func (db *GoLevelDB) Put(key []byte, value []byte) error {
	key = nonNilBytes(key)
	value = nonNilBytes(value)
	return db.db.Put(key, value)
}

// Close implements OKDB.
func (db *GoLevelDB) Close() error {
	if err := db.db.Close(); err != nil {
		return err
	}
	return nil
}

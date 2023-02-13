package vm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

type OKDB interface {
	// Get fetches the value of the given key, or nil if it does not exist.
	// CONTRACT: key, value readonly []byte
	Get([]byte) ([]byte, error)

	// Put sets the value for the given key, replacing it if it already exists.
	// CONTRACT: key, value readonly []byte
	Put([]byte, []byte) error

	// Close closes the database connection.
	Close() error
}

type BackendType string

// These are valid backend types.
const (
	GoLevelDBBackend BackendType = "goleveldb"
	RocksDBBackend   BackendType = "rocksdb"
)

type dbCreator func(name string, dir string) (OKDB, error)

var backends = map[BackendType]dbCreator{}

func registerDBCreator(backend BackendType, creator dbCreator, force bool) {
	_, ok := backends[backend]
	if !force && ok {
		return
	}
	backends[backend] = creator
}

var (
	txDB    OKDB
	blockDB OKDB
	tokenDB OKDB
)

func InitDB(dir string, backend BackendType) error {
	var dbDir string
	if dir != "" {
		dbDir = filepath.Join(dir, "okdb")
	} else {
		dbDir = filepath.Join("okdb")
	}
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		err := os.MkdirAll(dbDir, 0700)
		if err != nil {
			return fmt.Errorf("could not create directory %v: %w", dir, err)
		}
	}

	txDB = newDB("InnerTxDB", backend, dbDir)
	blockDB = newDB("InnerBlockDB", backend, dbDir)
	tokenDB = newDB("TokenDB", backend, dbDir)

	return nil
}

func newDB(name string, backend BackendType, dir string) OKDB {
	dbCreator, ok := backends[backend]
	if !ok {
		keys := make([]string, len(backends))
		i := 0
		for k := range backends {
			keys[i] = string(k)
			i++
		}
		panic(fmt.Sprintf("Unknown db_backend %s, expected either %s", backend, strings.Join(keys, " or ")))
	}

	db, err := dbCreator(name, dir)
	if err != nil {
		panic(fmt.Sprintf("Error initializing DB: %v", err))
	}
	return db
}

func CloseDB() (errors []error) {
	if err := txDB.Close(); err != nil {
		errors = append(errors, err)
	}

	if err := blockDB.Close(); err != nil {
		errors = append(errors, err)
	}

	if err := tokenDB.Close(); err != nil {
		errors = append(errors, err)
	}

	log.Info("Close innerTx DB")
	return errors
}

func ReadToken(key []byte) []byte {
	rtn, _ := tokenDB.Get(key)
	return rtn
}

func WriteTx(hash string, ix []*InnerTx) error {
	row, _ := rlp.EncodeToBytes(ix)
	err := txDB.Put([]byte(hash), row)
	if err != nil {
		log.Info("WriteTx error:" + hash)
		return err
	}
	return nil
}

func WriteBlockDB(blockhash string, hash []string) error {
	if len(blockhash) != 0 && len(hash) != 0 {
		row, _ := rlp.EncodeToBytes(hash)
		err := blockDB.Put([]byte(blockhash), row)
		if err != nil {
			log.Info("WriteBlock err:" + blockhash)
			return err
		}
	}
	return nil
}

func WriteToken(key []byte, value []byte) error {
	err := tokenDB.Put(key, value)
	if err != nil {
		log.Info("WriteToken error:" + string(key))
		return err
	}
	return nil
}

func GetFromDB(hash string) []InnerTx {
	result, err := txDB.Get([]byte(hash))
	if err != nil {
		return nil
	}

	innerTxs := make([]InnerTx, 0)
	rlp.DecodeBytes(result, &innerTxs)
	return innerTxs
}

func GetBlockDB(blockHash string) []string {
	result, err := blockDB.Get([]byte(blockHash))
	if err != nil {
		return nil
	}

	var innerTxs []string
	rlp.DecodeBytes(result, &innerTxs)
	return innerTxs
}

// We defensively turn nil keys or values into []byte{} for
// most operations.
func nonNilBytes(bz []byte) []byte {
	if bz == nil {
		return []byte{}
	}
	return bz
}

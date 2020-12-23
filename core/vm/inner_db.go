package vm

import (
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"path/filepath"
)

var db  *leveldb.Database
var blockDB * leveldb.Database
var tokenDB *leveldb.Database
var inited bool
var blockInited bool
var initedToken bool

func InitDB(datadir string) error {
	if err := InitBlockDB(datadir); err != nil {
		return err
	}
	if err := InitTxDB(datadir); err != nil {
		return err
	}
	if err := InitTokenDB(datadir); err != nil {
		return err
	}
	return nil
}

func CloseDB(){
	db.Close()
	blockDB.Close()
	tokenDB.Close()
	log.Info("Close innerTx DB")
}

func InitTokenDB(datadir string) error{
	var path string
	if datadir != "" {
		path = filepath.Join(datadir, "okdb", "TokenDB")
	}else{
		path = filepath.Join("okdb", "TokenDB")
	}
	var token, err = leveldb.New(path, 256, 128, "tokenInfo")
	if err != nil {
		initedToken = false
		log.Info("Init tokenlevelDB failed", err)
		return err
	}else{
		tokenDB = token
		initedToken = true;
		log.Info("Init tokenDB")
		return nil
	}
}

func InitTxDB(datadir string) error{
	var path string
	if datadir != "" {
		path = filepath.Join(datadir, "okdb", "InnerTxDB")
	}else{
		path = filepath.Join("okdb", "InnerTxDB")
	}
	var createDB, err = leveldb.New(path, 256, 128, "contract")
	if err != nil {
		inited = false
		log.Info("Init levelDB failed", err)
		return err
	}else{
		db = createDB
		inited = true;
		log.Info("Init levelDB")
		return nil
	}
}

func InitBlockDB(datadir string) error{
	var path string
	if datadir != "" {
		path = filepath.Join(datadir, "okdb", "InnerBlockDB")
	}else{
		path = filepath.Join("okdb", "InnerBlockDB")
	}
	var db, err = leveldb.New(path, 256, 128, "blockTx")
	if err != nil {
		blockInited = false
		log.Info("Init blockDB failed", err)
		return err
	}else{
		blockDB = db
		blockInited = true;
		log.Info("Init blockDB")
		return nil
	}
}

func ReadToken(key []byte) []byte{
	rtn,_ := tokenDB.Get(key)
	return rtn
}

func WriteTx(hash string, ix []*InnerTx) error{
	row, _ := rlp.EncodeToBytes(ix)
	err := db.Put([]byte(hash), row)
	if err != nil {
		log.Info("Writetx error:" + hash)
		return err
	}
	return nil
}

func WriteBlockDB(blockhash string, hash []string) error{
	if len(blockhash) != 0 && len(hash) != 0 {
		row,_ := rlp.EncodeToBytes(hash)
		err := blockDB.Put([]byte(blockhash), row)
		if err != nil {
			log.Info("Writeblock err:" + blockhash)
			return err
		}
	}
	return nil
}

func WriteToken(key []byte, value []byte) error{
	err := tokenDB.Put(key, value)
	if err != nil {
		log.Info("Writetoken error:" + string(key))
		return err
	}
	return nil
}

func GetFromDB(hash string) []InnerTx{
	result, err := db.Get([]byte(hash))
	if err == nil {
		innerTxs := make([]InnerTx,0)
		rlp.DecodeBytes(result, &innerTxs)
		return innerTxs
	}else {
		return nil
	}
}

func GetBlockDB(blockhash string) []string{
	result, err := blockDB.Get([]byte(blockhash))
	if err == nil {
		var innerTxies []string
		rlp.DecodeBytes(result, &innerTxies)
		return innerTxies
	}else {
		return nil
	}
}
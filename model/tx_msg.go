package model

import (
	"github.com/irisnet/rainbow-sync/db"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type TxMsg struct {
	Time      int64    `bson:"time"`
	TxFee     Coin     `bson:"tx_fee"`
	Height    int64    `bson:"height"`
	TxHash    string   `bson:"tx_hash"`
	Type      string   `bson:"type"`
	MsgIndex  int      `bson:"msg_index"`
	TxIndex   uint32   `bson:"tx_index"`
	TxStatus  string   `bson:"tx_status"`
	TxMemo    string   `bson:"tx_memo"`
	TxLog     string   `bson:"tx_log"`
	GasUsed   int64    `bson:"gas_used"`
	GasWanted int64    `bson:"gas_wanted"`
	Events    []Event  `bson:"events"`
	Msg       DocTxMsg `bson:"msg"`
	Addrs     []string `bson:"addrs"`
	TxSigners []string `bson:"tx_signers"`
}

const (
	IrisTxMsgStatus         = "tx_status"
	IrisTxMsgSigners        = "tx_signers"
	CollectionNameIrisTxMsg = "sync_iris_tx_msg"
)

func (d TxMsg) Name() string {
	return CollectionNameIrisTxMsg
}

func (d TxMsg) PkKvPair() map[string]interface{} {
	return bson.M{"tx_hash": d.TxHash, "msg_index": d.MsgIndex}
}

func (d TxMsg) EnsureIndexes() {
	var indexes []mgo.Index
	indexes = append(indexes,
		mgo.Index{
			Key:        []string{"-height"},
			Background: true},
		mgo.Index{
			Key:        []string{"-tx_hash", "-msg_index"},
			Unique:     true,
			Background: true},
		mgo.Index{
			Key:        []string{"-type"},
			Background: true},
		mgo.Index{
			Key:        []string{"-" + IrisTxMsgSigners},
			Background: true},
		mgo.Index{
			Key:        []string{"-addrs"},
			Background: true},
	)

	db.EnsureIndexes(d.Name(), indexes)
}

package block

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/irisnet/rainbow-sync/service/iris/constant"
	model "github.com/irisnet/rainbow-sync/service/iris/db"
	"github.com/irisnet/rainbow-sync/service/iris/helper"
	"github.com/irisnet/rainbow-sync/service/iris/logger"
	imodel "github.com/irisnet/rainbow-sync/service/iris/model"
	docTxMsg "github.com/irisnet/rainbow-sync/service/iris/model/msg"
	"github.com/irisnet/rainbow-sync/service/iris/utils"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/types"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2/txn"
	"strconv"
	"time"
)

const (
	IRIS = "Iris"
)

type Iris_Block struct{}

func (iris *Iris_Block) Name() string {
	return IRIS
}

func (iris *Iris_Block) SaveDocsWithTxn(blockDoc *imodel.Block, irisTxs []*imodel.IrisTx, taskDoc imodel.SyncTask) error {
	var (
		ops, irisTxsOps []txn.Op
	)

	if blockDoc.Height == 0 {
		return fmt.Errorf("invalid block, height equal 0")
	}

	blockOp := txn.Op{
		C:      imodel.CollectionNameBlock,
		Id:     bson.NewObjectId(),
		Insert: blockDoc,
	}

	length_txs := len(irisTxs)
	if length_txs > 0 {
		irisTxsOps = make([]txn.Op, 0, length_txs)
		for _, v := range irisTxs {
			op := txn.Op{
				C:      imodel.CollectionNameIrisTx,
				Id:     bson.NewObjectId(),
				Insert: v,
			}
			irisTxsOps = append(irisTxsOps, op)
		}
	}

	updateOp := txn.Op{
		C:      imodel.CollectionNameSyncTask,
		Id:     taskDoc.ID,
		Assert: txn.DocExists,
		Update: bson.M{
			"$set": bson.M{
				"current_height":   taskDoc.CurrentHeight,
				"status":           taskDoc.Status,
				"last_update_time": taskDoc.LastUpdateTime,
			},
		},
	}

	ops = make([]txn.Op, 0, length_txs+2)
	ops = append(append(ops, blockOp, updateOp))
	ops = append(ops, irisTxsOps...)

	if len(ops) > 0 {
		err := model.Txn(ops)
		if err != nil {
			return err
		}
	}

	return nil
}

func (iris *Iris_Block) ParseBlock(b int64, client *helper.Client) (resBlock *imodel.Block, resIrisTxs []*imodel.IrisTx, resErr error) {

	defer func() {
		if err := recover(); err != nil {
			logger.Error("parse iris block fail", logger.Int64("height", b),
				logger.Any("err", err), logger.String("Chain Block", iris.Name()))

			resBlock = &imodel.Block{}
			resIrisTxs = nil
			resErr = fmt.Errorf("%v", err)
		}
	}()
	irisTxs, err := iris.ParseIrisTxs(b, client)
	if err != nil {
		logger.Error("parse iris txs", logger.String("error", err.Error()), logger.String("Chain Block", iris.Name()))
	}

	resBlock = &imodel.Block{
		Height:     b,
		CreateTime: time.Now().Unix(),
	}
	resIrisTxs = irisTxs
	resErr = err

	return
}

// parse iris txs  from block result txs
func (iris *Iris_Block) ParseIrisTxs(b int64, client *helper.Client) ([]*imodel.IrisTx, error) {
	resblock, err := client.Block(&b)
	if err != nil {
		logger.Warn("get block result err, now try again", logger.String("err", err.Error()),
			logger.String("Chain Block", iris.Name()), logger.Any("height", b))
		// there is possible parse block fail when in iterator
		var err2 error
		client2 := helper.GetClient()
		resblock, err2 = client2.Block(&b)
		client2.Release()
		if err2 != nil {
			return nil, err2
		}
	}

	irisTxs := make([]*imodel.IrisTx, 0, len(resblock.Block.Txs))
	for _, tx := range resblock.Block.Txs {
		iristx := iris.ParseIrisTxModel(tx, resblock.Block)
		irisTxs = append(irisTxs, &iristx)
	}

	return irisTxs, nil
}

// parse iris tx from iris block result tx
func (iris *Iris_Block) ParseIrisTxModel(txBytes types.Tx, block *types.Block) imodel.IrisTx {

	var (
		authTx     auth.StdTx
		methodName = "ParseTx"
		docTx      imodel.IrisTx
		docMsgs    []imodel.DocTxMsg
	)

	cdc := utils.GetCodec()

	err := cdc.UnmarshalBinaryLengthPrefixed(txBytes, &authTx)
	if err != nil {
		logger.Error(err.Error())
		return docTx
	}

	height := block.Height
	txTime := block.Time
	txHash := utils.BuildHex(txBytes.Hash())
	fee := utils.BuildFee(authTx.Fee)
	memo := authTx.Memo

	// get tx status, gasUsed, gasPrice and actualFee from tx result
	status, result, err := utils.QueryTxResult(txBytes.Hash())
	if err != nil {
		logger.Error("get txResult err", logger.String("method", methodName), logger.String("err", err.Error()))
	}
	msgs := authTx.GetMsgs()
	if len(msgs) <= 0 {
		logger.Error("can't get msgs", logger.String("method", methodName))
		return docTx
	}

	docTx = imodel.IrisTx{
		Height: height,
		Time:   txTime,
		TxHash: txHash,
		Fee:    &fee,
		Memo:   memo,
		Status: status,
		Log:    result.Log,
		Code:   result.Code,
		Events: parseEvents(&result),
	}
	for _, msg := range msgs {
		switch msg.(type) {
		case imodel.MsgTransfer:
			msg := msg.(imodel.MsgTransfer)
			docTx.Initiator = msg.FromAddress.String()
			docTx.From = msg.FromAddress.String()
			docTx.To = msg.ToAddress.String()
			docTx.Amount = utils.ParseCoins(msg.Amount)
			docTx.Type = constant.TxTypeTransfer
			break
		case imodel.IBCBankMsgTransfer:
			msg := msg.(imodel.IBCBankMsgTransfer)
			docTx.Initiator = msg.Sender
			docTx.From = docTx.Initiator
			docTx.To = msg.Receiver
			docTx.Amount = buildCoins(msg.Denomination, msg.Amount.String())
			docTx.Type = constant.TxTypeIBCBankTransfer
			docTx.IBCPacketHash = buildIBCPacketHashByEvents(docTx.Events)
			txMsg := docTxMsg.DocTxMsgIBCBankTransfer{}
			txMsg.BuildMsg(msg)
			docTx.Msgs = append(docMsgs, imodel.DocTxMsg{
				Type: txMsg.Type(),
				Msg:  &txMsg,
			})
			break
		case imodel.IBCBankMsgReceivePacket:
			msg := msg.(imodel.IBCBankMsgReceivePacket)
			docTx.Initiator = msg.Signer.String()
			docTx.Type = constant.TxTypeIBCBankRecvTransferPacket

			if transPacketData, err := buildIBCPacketData(msg.Packet.Data()); err != nil {
				logger.Error("build ibc packet data fail", logger.String("packetData", string(msg.Packet.Data())),
					logger.String("err", err.Error()))
			} else {
				docTx.From = transPacketData.Sender
				docTx.To = transPacketData.Receiver
				docTx.Amount = buildCoins(transPacketData.Denomination, transPacketData.Amount)
			}

			if hash, err := buildIBCPacketHashByPacket(msg.Packet.(imodel.IBCPacket)); err != nil {
				logger.Error("build ibc packet hash fail", logger.String("err", err.Error()))
			} else {
				docTx.IBCPacketHash = hash
			}

			txMsg := docTxMsg.DocTxMsgIBCBankReceivePacket{}
			txMsg.BuildMsg(msg)
			docTx.Msgs = append(docMsgs, imodel.DocTxMsg{
				Type: txMsg.Type(),
				Msg:  &txMsg,
			})
			break
		default:
			logger.Warn("unknown msg type")
		}
	}

	return docTx

}

func parseEvents(result *abci.ResponseDeliverTx) []imodel.Event {

	var events []imodel.Event
	for _, val := range result.GetEvents() {
		one := imodel.Event{
			Type: val.Type,
		}
		one.Attributes = make(map[string]string, len(val.Attributes))
		for _, attr := range val.Attributes {
			one.Attributes[string(attr.Key)] = string(attr.Value)
		}
		events = append(events, one)
	}

	return events
}

func buildCoins(denom string, amountStr string) []*imodel.Coin {
	var coins []*imodel.Coin
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		logger.Error("convert str to float64 fail", logger.String("amountStr", amountStr),
			logger.String("err", err.Error()))
		amount = 0
	}
	coin := imodel.Coin{
		Denom:  denom,
		Amount: amount,
	}
	return append(coins, &coin)
}

func buildIBCPacketHashByEvents(events []imodel.Event) string {
	var packetStr string
	if len(events) == 0 {
		return ""
	}

	for _, e := range events {
		if e.Type == constant.EventTypeSendPacket {
			for k, v := range e.Attributes {
				if k == constant.EventAttributesKeyPacket {
					packetStr = v
					break
				}
			}
		}
	}

	if packetStr == "" {
		return ""
	}

	return utils.Md5Encrypt([]byte(packetStr))
}

func buildIBCPacketHashByPacket(packet imodel.IBCPacket) (string, error) {
	data, err := packet.MarshalJSON()
	if err != nil {
		return "", err
	}
	return utils.Md5Encrypt(data), nil
}

func buildIBCPacketData(packetData []byte) (imodel.IBCTransferPacketDataValue, error) {
	var transferPacketData imodel.IBCTransferPacketData
	err := json.Unmarshal(packetData, &transferPacketData)
	if err != nil {
		return transferPacketData.Value, err
	}

	return transferPacketData.Value, nil
}

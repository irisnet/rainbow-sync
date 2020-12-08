package service

import (
	"encoding/hex"
	. "github.com/irisnet/rainbow-sync/msgs"
	"github.com/irisnet/rainbow-sync/utils"
	"strings"
)

type (
	DocMsgKillRequestContext struct {
		RequestContextId string `bson:"request_context_id" yaml:"request_context_id"`
		Consumer         string `bson:"consumer" yaml:"consumer"`
	}
)

func (m *DocMsgKillRequestContext) GetType() string {
	return MsgTypeKillRequestContext
}

func (m *DocMsgKillRequestContext) BuildMsg(v interface{}) {
	msg := v.(*MsgKillRequestContext)

	m.RequestContextId = strings.ToUpper(hex.EncodeToString(msg.RequestContextId))
	m.Consumer = msg.Consumer.String()
}

func (m *DocMsgKillRequestContext) HandleTxMsg(v SdkMsg) MsgDocInfo {
	var (
		addrs []string
		msg   MsgKillRequestContext
	)

	utils.UnMarshalJsonIgnoreErr(utils.MarshalJsonIgnoreErr(v), &msg)
	addrs = append(addrs, msg.Consumer.String())
	handler := func() (Msg, []string) {
		return m, addrs
	}

	return CreateMsgDocInfo(v, handler)
}
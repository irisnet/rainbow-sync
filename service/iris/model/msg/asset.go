package msg

import (
	"strings"
	"github.com/irisnet/rainbow-sync/service/iris/constant"
	"github.com/irisnet/rainbow-sync/service/iris/logger"
	imodel "github.com/irisnet/rainbow-sync/service/iris/model"
)

type (
	DocTxMsgIssueToken struct {
		Family          string           `bson:"family"`
		Source          string           `bson:"source"`
		Gateway         string           `bson:"gateway"`
		Symbol          string           `bson:"symbol"`
		CanonicalSymbol string           `bson:"canonical_symbol"`
		Name            string           `bson:"name"`
		Decimal         uint8            `bson:"decimal"`
		MinUnitAlias    string           `bson:"min_unit_alias"`
		InitialSupply   uint64           `bson:"initial_supply"`
		MaxSupply       uint64           `bson:"max_supply"`
		Mintable        bool             `bson:"mintable"`
		Owner           string           `bson:"owner"`
		UdInfo          assetTokenUdInfo `bson:"ud_info"`
	}

	DocTxMsgEditToken struct {
		TokenId         string           `bson:"token_id"`         //  id of token
		Owner           string           `bson:"owner"`            //  owner of token
		CanonicalSymbol string           `bson:"canonical_symbol"` //  canonical_symbol of token
		MinUnitAlias    string           `bson:"min_unit_alias"`   //  min_unit_alias of token
		MaxSupply       uint64           `bson:"max_supply"`
		Mintable        string           `bson:"mintable"` //  mintable of token
		Name            string           `bson:"name"`
		UdInfo          assetTokenUdInfo `bson:"ud_info"`
	}

	DocTxMsgMintToken struct {
		TokenId string           `bson:"token_id"` // the unique id of the token
		Owner   string           `bson:"owner"`    // the current owner address of the token
		To      string           `bson:"to"`       // address of mint token to
		Amount  uint64           `bson:"amount"`   // amount of mint token
		UdInfo  assetTokenUdInfo `bson:"ud_info"`
	}

	DocTxMsgTransferTokenOwner struct {
		SrcOwner string           `bson:"src_owner"` // the current owner address of the token
		DstOwner string           `bson:"dst_owner"` // the new owner
		TokenId  string           `bson:"token_id"`
		UdInfo   assetTokenUdInfo `bson:"ud_info"`
	}

	DocTxMsgCreateGateway struct {
		Owner    string `bson:"owner"`    //  the owner address of the gateway
		Moniker  string `bson:"moniker"`  //  the globally unique name of the gateway
		Identity string `bson:"identity"` //  the identity of the gateway
		Details  string `bson:"details"`  //  the description of the gateway
		Website  string `bson:"website"`  //  the external website of the gateway
	}

	DocTxMsgEditGateway struct {
		Owner    string `bson:"owner"`    // Owner of the gateway
		Moniker  string `bson:"moniker"`  // Moniker of the gateway
		Identity string `bson:"identity"` // Identity of the gateway
		Details  string `bson:"details"`  // Details of the gateway
		Website  string `bson:"website"`  // Website of the gateway
	}

	DocTxMsgTransferGatewayOwner struct {
		Owner   string `bson:"owner"`   // the current owner address of the gateway
		Moniker string `bson:"moniker"` // the unique name of the gateway to be transferred
		To      string `bson:"to"`      // the new owner to which the gateway ownership will be transferred
	}

	assetTokenUdInfo struct {
		Source  string `bson:"source"`
		Gateway string `bson:"gateway"`
		Symbol  string `bson:"symbol"`
	}
)

const (
	separatorOfSymbolInTokenId = "."
	externalTokenPrefix        = "x"
	SourceNative               = "native"
	SourceExternal             = "external"
	SourceGateWay              = "gateway"
)

func (m *DocTxMsgIssueToken) Type() string {
	return constant.TxMsgTypeAssetIssueToken
}

func (m *DocTxMsgIssueToken) BuildMsg(txMsg interface{}) {
	msg := txMsg.(imodel.AssetIssueToken)

	m.Family = msg.Family.String()
	m.Source = msg.Source.String()
	m.Gateway = msg.Gateway
	m.Symbol = msg.Symbol
	m.CanonicalSymbol = msg.CanonicalSymbol
	m.Name = msg.Name
	m.Decimal = msg.Decimal
	m.MinUnitAlias = msg.MinUnitAlias
	m.InitialSupply = msg.InitialSupply
	m.MaxSupply = msg.MaxSupply
	m.Mintable = msg.Mintable
	m.Owner = msg.Owner.String()
	m.UdInfo = assetTokenUdInfo{
		Source:  msg.Source.String(),
		Gateway: msg.Gateway,
		Symbol:  msg.Symbol,
	}
}

func (m *DocTxMsgEditToken) Type() string {
	return constant.TxMsgTypeAssetEditToken
}

func (m *DocTxMsgEditToken) BuildMsg(txMsg interface{}) {
	msg := txMsg.(imodel.AssetEditToken)

	m.TokenId = msg.TokenId
	m.Owner = msg.Owner.String()
	m.CanonicalSymbol = msg.CanonicalSymbol
	m.MinUnitAlias = msg.MinUnitAlias
	m.MaxSupply = msg.MaxSupply
	m.Mintable = string(msg.Mintable)
	m.Name = msg.Name
	m.UdInfo = getAssetTokenUdInfo(msg.TokenId)
}

func (m *DocTxMsgMintToken) Type() string {
	return constant.TxMsgTypeAssetMintToken
}

func (m *DocTxMsgMintToken) BuildMsg(txMsg interface{}) {
	msg := txMsg.(imodel.AssetMintToken)

	m.TokenId = msg.TokenId
	m.Owner = msg.Owner.String()
	m.To = msg.To.String()
	m.Amount = msg.Amount
	m.UdInfo = getAssetTokenUdInfo(msg.TokenId)
}

func (m *DocTxMsgTransferTokenOwner) Type() string {
	return constant.TxMsgTypeAssetTransferTokenOwner
}

func (m *DocTxMsgTransferTokenOwner) BuildMsg(txMsg interface{}) {
	msg := txMsg.(imodel.AssetTransferTokenOwner)

	m.SrcOwner = msg.SrcOwner.String()
	m.DstOwner = msg.DstOwner.String()
	m.TokenId = msg.TokenId
	m.UdInfo = getAssetTokenUdInfo(msg.TokenId)
}

func (m *DocTxMsgCreateGateway) Type() string {
	return constant.TxMsgTypeAssetCreateGateway
}

func (m *DocTxMsgCreateGateway) BuildMsg(txMsg interface{}) {
	msg := txMsg.(imodel.AssetCreateGateway)

	m.Owner = msg.Owner.String()
	m.Moniker = msg.Moniker
	m.Identity = msg.Identity
	m.Details = msg.Details
	m.Website = msg.Website
}

func (m *DocTxMsgEditGateway) Type() string {
	return constant.TxMsgTypeAssetEditGateway
}

func (m *DocTxMsgEditGateway) BuildMsg(txMsg interface{}) {
	msg := txMsg.(imodel.AssetEditGateWay)

	m.Owner = msg.Owner.String()
	m.Moniker = msg.Moniker
	m.Identity = msg.Identity
	m.Details = msg.Details
	m.Website = msg.Website
}

func (m *DocTxMsgTransferGatewayOwner) Type() string {
	return constant.TxMsgTypeAssetTransferGatewayOwner
}

func (m *DocTxMsgTransferGatewayOwner) BuildMsg(txMsg interface{}) {
	msg := txMsg.(imodel.AssetTransferGatewayOwner)

	m.Owner = msg.Owner.String()
	m.Moniker = msg.Moniker
	m.To = msg.To.String()
}

// get assetTokenUdInfo by parse tokenId
// Global Unique Token ID Generation Rule
// When Source is native: ID = [Symbol], e.g. iris
// When Source is external: ID = x.[Symbol], e.g. x.btc
// When Source is gateway: ID = [Gateway].[Symbol], e.g. cats.kitty
func getAssetTokenUdInfo(tokenId string) assetTokenUdInfo {
	var (
		source, gateway, symbol string
	)
	strs := strings.Split(tokenId, separatorOfSymbolInTokenId)

	switch len(strs) {
	case 1:
		source = SourceNative
		symbol = tokenId
		break
	case 2:
		if strs[0] == externalTokenPrefix {
			source = SourceExternal
			symbol = strs[1]
		} else {
			source = SourceGateWay
			gateway = strs[0]
			symbol = strs[1]
		}
	default:
		logger.Warn("can't get assetUserDefinedInfo by tokenId", logger.String("tokenId", tokenId))
	}

	return assetTokenUdInfo{
		Source:  source,
		Gateway: gateway,
		Symbol:  symbol,
	}
}

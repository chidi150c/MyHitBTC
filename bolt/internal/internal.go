package internal

import (
	"myhitbtcv4/model"

	proto "github.com/golang/protobuf/proto"
)

//go generate internal.proto
// MarshalApp encodes an App to binary format.
func MarshalAppData(WebJSON *model.AppData) ([]byte, error) {
	return proto.Marshal(&AppData{
		ID:                     uint64(WebJSON.ID),
		UsrID:                  uint64(WebJSON.UsrID),
		SessID:                 string(WebJSON.SessID),
		PublicKey:              WebJSON.PublicKey,
		Secret:                 WebJSON.Secret,
		Host:                   WebJSON.Host,
		SymbolCode:             WebJSON.SymbolCode,
		Side:                   WebJSON.Side,
		MrktQuantity:           WebJSON.MrktQuantity,
		MrktBuyPrice:           WebJSON.MrktBuyPrice,
		MrktSellPrice:          WebJSON.MrktSellPrice,
		NeverBought:            WebJSON.NeverBought,
		NeverSold:              WebJSON.NeverSold,
		QuantityIncrement:      WebJSON.QuantityIncrement,
		Message:                WebJSON.Message,
		TickSize:               WebJSON.TickSize,
		TakeLiquidityRate:      WebJSON.TakeLiquidityRate,
		SuccessfulOrders:       WebJSON.SuccessfulOrders,
		MadeProfitOrders:       WebJSON.MadeProfitOrders,
		MadeLostOrders:         WebJSON.MadeLostOrders,
		StopLostPoint:          WebJSON.StopLostPoint,
		BaseCurrency:           WebJSON.BaseCurrency,
		QuoteCurrency:          WebJSON.QuoteCurrency,
		TrailPoints:            WebJSON.TrailPoints,
		LeastProfitMargin:      WebJSON.LeastProfitMargin,
		SpinOutReason:          WebJSON.SpinOutReason,
		SureTradeFactor:        WebJSON.SureTradeFactor,
		Hodler:                 WebJSON.Hodler,
		GoodBiz:                WebJSON.GoodBiz,
		AlternateData:          WebJSON.AlternateData,
		InstantProfit:          WebJSON.InstantProfit,
		InstantLost:            WebJSON.InstantLost,
		TotalProfit:            WebJSON.TotalProfit,
		TotalLost:              WebJSON.TotalLost,
		PriceTradingStarted:    WebJSON.PriceTradingStarted,
		MainStartPointSell:     WebJSON.MainStartPointSell,
		SoldQuantity:           WebJSON.SoldQuantity,
		BoughtQuantity:         WebJSON.BoughtQuantity,
		MainStartPointBuy:      WebJSON.MainStartPointBuy,
		MainQuantity:           WebJSON.MainQuantity,
		NextStartPointNegPrice: WebJSON.NextStartPointNegPrice,
		NextStartPointPrice:    WebJSON.NextStartPointPrice,
		ProfitPointFactor:      WebJSON.ProfitPointFactor,
		HodlerQuantity:         WebJSON.HodlerQuantity,
		PendingA:               WebJSON.PendingA,
		PendingB:               WebJSON.PendingB,
		HeartbeatBuy:           WebJSON.HeartbeatBuy,
		MessageFilter:          WebJSON.MessageFilter,
		HeartbeatSell:          WebJSON.HeartbeatSell,
		NextMarketBuyPoint:     WebJSON.NextMarketBuyPoint,
		NextMarketSellPoint:    WebJSON.NextMarketSellPoint,
		DisableTransaction:     WebJSON.DisableTransaction,
		ProfitPriceUsed:        WebJSON.ProfitPriceUsed,
		ProfitPrice:            WebJSON.ProfitPrice,
	})
}

// UnmarshalAppData decodes a User from a binary data.
func UnmarshalAppData(data []byte, d *model.AppData) error {
	var pb AppData
	if err := proto.Unmarshal(data, &pb); err != nil {
		return err
	}
	d.ID = model.AppID(pb.ID)
	d.UsrID = model.UserID(pb.UsrID)
	d.SessID = model.SessionID(pb.SessID)
	d.PublicKey = pb.PublicKey
	d.Secret = pb.Secret
	d.Host = pb.Host
	d.SymbolCode = pb.SymbolCode
	d.Side = pb.Side
	d.MrktQuantity = pb.MrktQuantity
	d.MrktBuyPrice = pb.MrktBuyPrice
	d.MrktSellPrice = pb.MrktSellPrice
	d.NeverBought = pb.NeverBought
	d.NeverSold = pb.NeverSold
	d.QuantityIncrement = pb.QuantityIncrement
	d.Message = pb.Message
	d.TickSize = pb.TickSize
	d.TakeLiquidityRate = pb.TakeLiquidityRate
	d.SuccessfulOrders = pb.SuccessfulOrders
	d.MadeProfitOrders = pb.MadeProfitOrders
	d.MadeLostOrders = pb.MadeLostOrders
	d.StopLostPoint = pb.StopLostPoint
	d.BaseCurrency = pb.BaseCurrency
	d.QuoteCurrency = pb.QuoteCurrency
	d.TrailPoints = pb.TrailPoints
	d.LeastProfitMargin = pb.LeastProfitMargin
	d.SpinOutReason = pb.SpinOutReason
	d.SureTradeFactor = pb.SureTradeFactor
	d.Hodler = pb.Hodler
	d.GoodBiz = pb.GoodBiz
	d.AlternateData = pb.AlternateData
	d.InstantProfit = pb.InstantProfit
	d.InstantLost = pb.InstantLost
	d.TotalProfit = pb.TotalProfit
	d.TotalLost = pb.TotalLost
	d.PriceTradingStarted = pb.PriceTradingStarted
	d.MainStartPointSell = pb.MainStartPointSell
	d.SoldQuantity = pb.SoldQuantity
	d.BoughtQuantity = pb.BoughtQuantity
	d.MainStartPointBuy = pb.MainStartPointBuy
	d.MainQuantity = pb.MainQuantity
	d.NextStartPointNegPrice = pb.NextStartPointNegPrice
	d.NextStartPointPrice = pb.NextStartPointPrice
	d.ProfitPointFactor = pb.ProfitPointFactor
	d.HodlerQuantity = pb.HodlerQuantity
	d.PendingA = pb.PendingA
	d.PendingB = pb.PendingB
	d.HeartbeatBuy = pb.HeartbeatBuy
	d.MessageFilter = pb.MessageFilter
	d.HeartbeatSell = pb.HeartbeatSell
	d.NextMarketBuyPoint = pb.NextMarketBuyPoint
	d.NextMarketSellPoint = pb.NextMarketSellPoint
	d.DisableTransaction = pb.DisableTransaction
	d.ProfitPriceUsed = pb.ProfitPriceUsed
	d.ProfitPrice = pb.ProfitPrice
	return nil
}

// MarshalUser encodes a User to binary format.
func MarshalUser(WebJSON *model.User) ([]byte, error) {
	p := &User{
		ID:        uint64(WebJSON.ID),
		SessID:    string(WebJSON.SessID),
		Username:  WebJSON.Username,
		Password:  WebJSON.Password,
		Firstname: WebJSON.Firstname,
		Lastname:  WebJSON.Lastname,
		Email:     WebJSON.Email,
		ImageURL:  WebJSON.ImageURL,
		Token:     WebJSON.Token,
		Url:       WebJSON.Url,
		Expiry:    WebJSON.Expiry,
		Host:      WebJSON.Host,
	}
	p.ApIDs = make(map[string]uint64)
	for k, v := range WebJSON.ApIDs {
		p.ApIDs[k] = uint64(v)
	}
	return proto.Marshal(p)
}

// UnmarshalUser decodes a User from a binary data.
func UnmarshalUser(data []byte, d *model.User) error {
	var pb User
	if err := proto.Unmarshal(data, &pb); err != nil {
		return err
	}
	d.ID = model.UserID(pb.ID)
	d.SessID = model.SessionID(pb.SessID)
	d.Username = pb.Username
	d.Password = pb.Password
	d.Firstname = pb.Firstname
	d.Lastname = pb.Lastname
	d.Email = pb.Email
	d.ImageURL = pb.ImageURL
	d.Token = pb.Token
	d.Url = pb.Url
	d.Expiry = pb.Expiry
	d.Host = pb.Host
	d.Level = pb.Level
	d.ApIDs = make(map[string]model.AppID)
	for k, v := range pb.ApIDs {
		d.ApIDs[k] = model.AppID(v)
	}
	return nil
}

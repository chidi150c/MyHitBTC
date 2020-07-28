package app

import (
	"fmt"
	"log"
	"myhitbtcv4/app/api"
	"myhitbtcv4/model"
	"strconv"
	"strings"
	"time"
)

//NewApp is the constructor of the App struct
func AdaptApp(w WorkerAppService, md *model.App, from string) (*model.App, error) {
	var (
		floatHolder float64
	)
	//log.Printf("For %s: For %s: AdaptApp started with Main Quantity %.8f", md.Data.SymbolCode, w.user.Username, md.Data.MainQuantity)
	if from == "" {
		md.Chans.CloseDownAutoTradeChan = make(chan bool)
		md.Chans.CloseDownWorkerChan = make(chan bool)
		md.Chans.MessageChan = make(chan string, 1000)
		md.Chans.SetParamChan = make(chan SetParam, 1)
		md.Chans.ShutDownMessageChan = make(chan string)
		md.Chans.PriceTradingNextStartChan = make(chan priceTradingVehicle, 1)
		md.Chans.CloseDownSellPriceTradingChan = make(chan bool)
		md.Chans.MarketResetInfoChan = make(chan chan bool)
		md.Chans.CloseDownBuyPriceTradingChan = make(chan bool)
		md.Chans.CancelMyOrderChan = make(chan bool)
		if md.FromVersionUpdate {
			if md.Data.PriceTradingStarted == "buy" {
				md.Data.Side = "sell"
				if !isClosed(md.Chans.CloseDownSellPriceTradingChan) {
					close(md.Chans.CloseDownSellPriceTradingChan)
				}
				md.Chans.CloseDownBuyPriceTradingChan = make(chan bool)
				go w.priceTrading(md, "")
			} else if md.Data.PriceTradingStarted == "sell" {
				md.Data.Side = "buy"
				if !isClosed(md.Chans.CloseDownBuyPriceTradingChan) {
					close(md.Chans.CloseDownBuyPriceTradingChan)
				}
				md.Chans.CloseDownSellPriceTradingChan = make(chan bool)
				go w.priceTrading(md, "")
			} else {
				md.FromVersionUpdate = false
				close(md.Chans.CloseDownSellPriceTradingChan)
				close(md.Chans.CloseDownBuyPriceTradingChan)
			}
			// md.Data.Hodler = "no"
			// md.Data.TrailPoints =
			if w.user.Username == "chinedu" {
				//md.Data.LeastProfitMargin = md.Data.LeastProfitMargin - 0.005
				//md.Data.MessageFilter = ""
				//md.Data.QuantityIncrement = md.Data.QuantityIncrement / 2
				//md.Data.GoodBiz = md.Data.QuantityIncrement * 4.0
			} else if w.user.Username == "chidi150c" {
				//md.Data.LeastProfitMargin = md.Data.LeastProfitMargin - 0.005
				//md.Data.MessageFilter = ""
				//md.Data.QuantityIncrement = md.Data.QuantityIncrement * 2
				//md.Data.GoodBiz = md.Data.QuantityIncrement * 4.0
			}
			// md.Data.DisableTransaction = ""
			return md, nil
		} else {
			close(md.Chans.CloseDownSellPriceTradingChan)
			close(md.Chans.CloseDownBuyPriceTradingChan)
		}
	}
	var (
		s                *api.Symbol
		JackportErr, err error
		qty, ts, tl      float64
	)
	s, err = w.API.GetSymbol(md.Data.SymbolCode)
	if err != nil {
		log.Printf("NewMarketData: For %s: Unable to get Symbol!!! due to: %v\n", md.Data.SymbolCode, err)
		ErrorMessage := fmt.Sprintf("%v", err)
		if strings.Contains(ErrorMessage, "502 Bad Gateway") || strings.Contains(ErrorMessage, "503 Service Unavailable") || strings.Contains(ErrorMessage, "500 Internal Server") || strings.Contains(ErrorMessage, "Exchange temporary closed") {
			time.Sleep(time.Millisecond * time.Duration(15000+<-randGen800()))
		} else {
			time.Sleep(time.Second * 5)
		}
		s, err = w.API.GetSymbol(md.Data.SymbolCode)
		if err != nil {
			log.Printf("NewMarketData: For %s: Unable to get Symbol!!! due to: %v\n", md.Data.SymbolCode, err)
			ErrorMessage := fmt.Sprintf("%v", err)
			if strings.Contains(ErrorMessage, "502 Bad Gateway") || strings.Contains(ErrorMessage, "503 Service Unavailable") || strings.Contains(ErrorMessage, "500 Internal Server") || strings.Contains(ErrorMessage, "Exchange temporary closed") {
				time.Sleep(time.Millisecond * time.Duration(60000+<-randGen800()))
			} else {
				time.Sleep(time.Second * 30)
			}
			s, err = w.API.GetSymbol(md.Data.SymbolCode)
			if err != nil {
				log.Printf("NewMarketData: For %s: Unable to get Symbol!!! due to: %v\n", md.Data.SymbolCode, err)
				return nil, err
			}
		}
	}
	qty, _ = strconv.ParseFloat(s.QuantityIncrement, 64)
	ts, _ = strconv.ParseFloat(s.TickSize, 64)
	tl, _ = strconv.ParseFloat(s.TakeLiquidityRate, 64)
	if from == "" { //Parameters not to be resetted
		md.Data.MessageFilter = ""
		md.Data.Hodler = "no"
		md.Data.TrailPoints, md.Data.LeastProfitMargin = w.trailProfitPointsFunc(md)
		md.Data.QuantityIncrement = qty * 1
		md.Data.GoodBiz = md.Data.QuantityIncrement * 4.0
		if w.user.Username == "chinedu" {
			md.Data.QuantityIncrement = qty * 1
			md.Data.GoodBiz = md.Data.QuantityIncrement * 1
		} else if w.user.Username == "chidi150c" {
			md.Data.QuantityIncrement = qty * 10
			md.Data.GoodBiz = md.Data.QuantityIncrement * 1
		}
		md.Data.DisableTransaction = ""
	}
	md.Data.Side = "buy"
	md.Data.MrktQuantity = 1
	md.Data.TickSize = ts
	md.Data.TakeLiquidityRate = tl
	md.Data.BaseCurrency = s.BaseCurrency
	md.Data.QuoteCurrency = s.QuoteCurrency
	if (md.Data.NeverBought != "no more aply") && (md.Data.NeverSold != "no more aply") {
		md.Data.NeverBought = "yes"
		md.Data.NeverSold = "yes"
	}
	md.Data.SpinOutReason = "normalActivity"
	md.Data.SureTradeFactor = 1.0
	lastPrice := 0.0
	for {
		tkr, err := w.API.GetTicker(md.Data.SymbolCode)
		if err != nil {
			ErrorMessage := fmt.Sprintf("%v", err)
			md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: Unable to get Ticker\n", md.Data.SymbolCode, w.user.Username)
			if strings.Contains(ErrorMessage, "502 Bad Gateway") || strings.Contains(ErrorMessage, "503 Service Unavailable") || strings.Contains(ErrorMessage, "500 Internal Server") || strings.Contains(ErrorMessage, "Exchange temporary closed") {
				time.Sleep(time.Millisecond * time.Duration(30000+<-w.Rand800))
			} else {
				time.Sleep(time.Second * 5)
			}
			continue
		}
		lastPrice, _ = strconv.ParseFloat(tkr.Last, 64)
		break
	}
	base, _, err := w.GetTradingBalance(md)
	if (from == "all") || (from == "buy") || (from == "sell") {
		if from == "sell" {
			floatHolder, JackportErr = w.jackpotOrderQuantity(md, lastPrice, md.Data.GoodBiz, "AdaptAppSell")
		}
		md.Chans.MessageChan <- fmt.Sprintf("AdapApp %s %s Starting AdaptApp made at Side = %s from = %s base = %.8f jackpotQ = %.8f JackportErr = %v /n", md.Data.SymbolCode, w.user.Username, md.Data.Side, from, base, floatHolder, JackportErr)
		md.Chans.SetParamChan <- SetParam{"", 0}
		md.Chans.SetParamChan <- SetParam{"", 0}
	}
	if from == "all" {
		if !isClosed(md.Chans.CloseDownBuyPriceTradingChan) {
			close(md.Chans.CloseDownBuyPriceTradingChan)
		}
		if !isClosed(md.Chans.CloseDownSellPriceTradingChan) {
			close(md.Chans.CloseDownSellPriceTradingChan)
		}
		if base > md.Data.QuantityIncrement && err == nil {
			md.Data.Side = "buy"
			md.Data.MrktQuantity = base
			md.Data.AlternateData = lastPrice
			md.Chans.CloseDownSellPriceTradingChan = make(chan bool)
			go w.priceTrading(md, "")
		}
	} else if from == "buy" {
		if !isClosed(md.Chans.CloseDownBuyPriceTradingChan) {
			close(md.Chans.CloseDownBuyPriceTradingChan)
		}
		if base > md.Data.QuantityIncrement && err == nil {
			if !isClosed(md.Chans.CloseDownSellPriceTradingChan) {
				close(md.Chans.CloseDownSellPriceTradingChan)
			}
			md.Data.Side = "buy"
			md.Data.MrktQuantity = base
			md.Data.AlternateData = lastPrice
			md.Chans.CloseDownSellPriceTradingChan = make(chan bool)
			go w.priceTrading(md, "")
		}
	} else if from == "sell" {
		if !isClosed(md.Chans.CloseDownSellPriceTradingChan) {
			close(md.Chans.CloseDownSellPriceTradingChan)
		}
		if JackportErr == nil {
			if !isClosed(md.Chans.CloseDownBuyPriceTradingChan) {
				close(md.Chans.CloseDownBuyPriceTradingChan)
			}
			md.Data.Side = "sell"
			md.Data.MrktQuantity = floatHolder
			md.Data.AlternateData = lastPrice
			md.Chans.CloseDownBuyPriceTradingChan = make(chan bool)
			go w.priceTrading(md, "")
		} else {
			md.Chans.MessageChan <- fmt.Sprintf("AdaptAppSell: %v", err)
		}
	} else {
		if (base > md.Data.QuantityIncrement) && (err == nil) { //md.Data.QuantityIncrement
			md.Data.Side = "buy"
			md.Data.MrktQuantity = base
			md.Data.AlternateData = lastPrice
			md.Chans.CloseDownSellPriceTradingChan = make(chan bool)
			go w.priceTrading(md, "NeverSold")
		}
	}
	md.Data.NextMarketBuyPoint = lastPrice * (1.0 - (md.Data.LeastProfitMargin * 4))
	md.Data.NextMarketSellPoint = lastPrice * (1.0 + (md.Data.LeastProfitMargin * 4))

	return md, nil
}

type priceTradingVehicle struct {
	LastPrice                 float64
	PriceTradingStartQuantity float64
}

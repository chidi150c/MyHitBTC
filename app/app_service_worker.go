package app

//run with go run main.go > output.txt 2>&1
import (
	// "io"
	// "net/http"
	// "bufio"
	// "net"
	"fmt"
	"log"
	"math"
	"math/rand"
	"myhitbtcv4/app/api"
	"myhitbtcv4/model"
	"strconv"
	"strings"
	"time"

	"github.com/apourchet/investment/lib/ema"
	hatchuid "github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

type ppPendingData struct {
	action      string
	profitPrice float64
}

//WorkerAppService is a tool provider for the methods, a tool can be a method or infrastructural data struct
type WorkerAppService struct {
	API                        api.Client
	Limit                      string
	Period                     string
	Period2                    string
	UUIDChan                   chan string
	Rand800                    <-chan int
	session                    *Session
	user                       *model.User
	profitPriceGenAChan        <-chan float64
	profitPriceGenBChan        <-chan float64
	profitPriceOriginatorAChan chan ppPendingData
	profitPriceResetAChan      chan bool
	profitPriceOriginatorBChan chan ppPendingData
	profitPriceResetBChan      chan bool
}

//NewWorkerAppService is the constructor of the WorkerAppService
func NewWorkerAppService(mdData *model.AppData, s *Session, uuidch chan string) WorkerAppService {
	c := api.NewClient(mdData.Host, mdData.PublicKey, mdData.Secret)
	pPOAChan := make(chan ppPendingData)
	pPRAChan := make(chan bool)
	pPOBChan := make(chan ppPendingData)
	pPRBChan := make(chan bool)
	w := WorkerAppService{
		API:                        c,
		Limit:                      "200",
		Period:                     "M15",
		Period2:                    "M30",
		UUIDChan:                   uuidch,
		Rand800:                    randGen800(),
		profitPriceOriginatorAChan: pPOAChan,
		profitPriceResetAChan:      pPRAChan,
		profitPriceOriginatorBChan: pPOBChan,
		profitPriceResetBChan:      pPRBChan,
	}
	w.profitPriceGenAChan = w.ProfitPriceAFunc(mdData, pPRAChan, pPOAChan)
	w.profitPriceGenBChan = w.ProfitPriceBFunc(mdData, pPRBChan, pPOBChan)
	return w
}

//MarginVeh  ....
type MarginVeh struct {
	ID     model.AppID
	Margin marginParam
}

//AutoTradeManager starts the App
func (w WorkerAppService) AutoTradeManager(md *App, marginChan chan MarginVeh) (<-chan AppVehicle, error) {
	user, err := w.session.Authenticate()
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}
	if user.ApIDs[md.Data.SymbolCode] != md.Data.ID {
		log.Printf("%v", errors.New("UnAuthorized"))
		return nil, err
	}
	w.user = user
	w.session.workerAppService = w
	mdChan := make(chan AppVehicle)
	s := md.Data.SymbolCode
	mdHolder := md
	//log.Printf("For %s: For %s: AutoTradeManager started", s, w.user.Username)
	for {
		select {
		case <-md.Chans.CloseDown:
			return nil, errors.New("App closed appruptly")
		default:
		}
		mdHolder, err = AdaptApp(w, md, "")
		if err != nil {
			log.Printf("For %s: Unable to Adapt App: Error %v", s, err)
			continue
		}
		md = mdHolder
		break
	}
	data := SetParam{}
	go w.marketTrading(md)
	go func() {
		myCloseDown := md.Chans.CloseDown
		messageChan := md.Chans.MessageChan
		returnChan := make(chan bool)
		for {
			select {
			case <-myCloseDown:
				myCloseDown = make(chan bool)
				messageChan = nil
				if isClosed(returnChan) {
					md.Chans.ShutDownMessageChan <- "App Shutdown Completed!!!"
					return
				}
				go func() {
					for {
						select {
						case md.Data.Message = <-md.Chans.MessageChan:
							log.Print(md.Data.Message)
							if strings.Contains(md.Data.Message, "App Shuting Down") && strings.Contains(md.Data.Message, "Market") {
								break
							} else if strings.Contains(md.Data.Message, "App Shuting Down") {
								for {
									msg := <-md.Chans.MessageChan
									if strings.Contains(msg, "App Shuting Down") && strings.Contains(msg, "Market") {
										break
									}
								}
								break
							} else {
								continue
							}
						}
						break
					}
					close(myCloseDown)
					close(returnChan)
				}()
			case mdChan <- AppVehicle{md, returnChan}: //MyChan: partly used by handler to feed UI
				<-returnChan
			case marginChan <- MarginVeh{md.Data.ID, marginParam{md.Data.SymbolCode, md.Data.SuccessfulOrders, md.Data.MadeProfitOrders, md.Data.MadeLostOrders, md.Data.TotalLost + md.Data.TotalProfit}}:
			case md.Data.Message = <-messageChan:
				if (strings.Contains(md.Data.Message, md.Data.MessageFilter) && md.Data.MessageFilter != "") || strings.Contains(md.Data.Message, "made") || strings.Contains(md.Data.Message, "Reset Authorized") || strings.Contains(md.Data.Message, "success:") || strings.Contains(md.Data.Message, "started") {
					log.Print(md.Data.Message)
				}
			case data = <-md.Chans.SetParamChan:
				switch data.Key {
				case "":
				case "StopLostPoint":
					md.Data.StopLostPoint = data.Value.(float64)
				case "HeartbeatBuy":
					md.Data.HeartbeatBuy = data.Value.(string)
				case "HeartbeatSell":
					md.Data.HeartbeatSell = data.Value.(string)
				case "SuccessfulOrders":
					md.Data.SuccessfulOrders += data.Value.(float64)
				case "SymbolCode":
					md.Data.SymbolCode = data.Value.(string)
				case "Side":
					md.Data.Side = data.Value.(string)
				case "AlternateData":
					md.Data.AlternateData = data.Value.(float64)
				case "NeverBought":
					md.Data.NeverBought = data.Value.(string)
				case "NeverSold":
					md.Data.NeverSold = data.Value.(string)
				case "MrktQuantity":
					md.Data.MrktQuantity = data.Value.(float64)
				case "MrktBuyPrice":
					md.Data.MrktBuyPrice = data.Value.(float64)
				case "MrktSellPrice":
					md.Data.MrktSellPrice = data.Value.(float64)
				case "ProfitPriceUsed":
					md.Data.ProfitPriceUsed = data.Value.(string)
				case "ProfitPrice":
					md.Data.ProfitPrice = data.Value.(float64)
				case "TotalLost":
					md.Data.InstantLost = data.Value.(float64)
					md.Data.TotalLost += data.Value.(float64)
				case "MadeLostOrders":
					md.Data.MadeLostOrders += data.Value.(float64)
				case "MadeProfitOrders":
					md.Data.MadeProfitOrders += data.Value.(float64)
				case "TotalProfit":
					md.Data.InstantProfit = data.Value.(float64)
					md.Data.TotalProfit += data.Value.(float64)
				case "SpinOutReason":
					md.Data.SpinOutReason = data.Value.(string)
				case "TrailPoints":
					md.Data.TrailPoints = data.Value.(float64)
				case "LeastProfitMargin":
					md.Data.LeastProfitMargin = data.Value.(float64)
				case "SureTradeFactor":
					md.Data.SureTradeFactor = data.Value.(float64)
				case "GoodBiz":
					md.Data.GoodBiz = data.Value.(float64)
					// 	case "ID":
					// 		md.Data.ID = data.Value.(AppID)
					// 	case "usrID":
					// 		md.Data.usrID = data.Value.(UserID)
					// 	case "SessID":
					// 		md.Data.SessID = data.Value.(SessionID)
					// 	case "PublicKey":
					// 		md.Data.PublicKey = data.Value.(string)
					// 	case "Secret":
					// 		md.Data.Secret = data.Value.(string)
					// 	case "Host":
					// 		md.Data.Host = data.Value.(string)
					// 	case "QuantityIncrement":
					// 		md.Data.QuantityIncrement = data.Value.(float64)
					// 	case "TickSize":
					// 		md.Data.TickSize = data.Value.(float64)
					// 	case "TakeLiquidityRate":
					// 		md.Data.TakeLiquidityRate = data.Value.(float64)
					// 	case "BaseCurrency":
					// 		md.Data.BaseCurrency = data.Value.(string)
					// 	case "QuoteCurrency":
					// 		md.Data.QuoteCurrency = data.Value.(string)
				}
			}
		}
	}()
	return mdChan, nil
}

//GetTradingBalance returns the trading balance in the trading account
func (w WorkerAppService) GetTradingBalance(md *App) (baseCurrencyBalance float64, quoteCurrencyBalance float64, err error) {
	var (
		bals api.Balances
		v    api.Balance
		i    float64
	)
	bals, err = w.API.GetTradingBalances()
	i = 1
	for _, v = range bals {
		if v.Currency == md.Data.BaseCurrency || v.Currency == md.Data.QuoteCurrency {
			if v.Currency == md.Data.BaseCurrency {
				baseCurrencyBalance, _ = strconv.ParseFloat(v.Available, 64)
				i++
			} else if v.Currency == md.Data.QuoteCurrency {
				quoteCurrencyBalance, _ = strconv.ParseFloat(v.Available, 64)
				i++
			}
			if i == 3 {
				break
			}
		}
	}
	return baseCurrencyBalance, quoteCurrencyBalance, err
}

//AppShutDown for
func (w WorkerAppService) AppShutDown(md *App) error {
	user, err := w.session.Authenticate()
	if err != nil {
		return err
	}
	if user.ApIDs[md.Data.SymbolCode] == md.Data.ID {
		if !isClosed(md.Chans.CloseDown) {
			close(md.Chans.CloseDown)
		}
		if msg := <-md.Chans.ShutDownMessageChan; strings.Contains(msg, "App Shutdown Completed!!!") {
			log.Printf(msg)
			return nil
		}
	}
	return fmt.Errorf("Unable to Delete/Shutdown App")
}
func waiting(t time.Duration) {
	time.Sleep(time.Millisecond * t)
}

//ResetApp resetts some params of md to their default
func (w WorkerAppService) ResetApp(md *App, rType string, sync chan bool) error {
	var (
		mdHolder *App
	)
	user, err := w.session.Authenticate()
	if err != nil {
		log.Printf("%v", err)
		sync <- false
		return err
	}
	if user.ApIDs[md.Data.SymbolCode] != md.Data.ID {
		sync <- false
		return errors.New("UnAuthorized")
	}
	respChan := make(chan bool)
	md.Chans.MarketResetInfoChan <- respChan
	md.Chans.MessageChan <- fmt.Sprintf("ResetApp: %s: %s: Reset Authorized | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
	s := md.Data.SymbolCode
	md.Chans.MessageChan <- fmt.Sprintf("ResetApp: %s: %s: Resetting AdaptApp params .... | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
	<-respChan
	for {
		select {
		case <-md.Chans.CloseDown:
			return errors.New("App closed appruptly")
		default:
		}
		mdHolder, err = AdaptApp(w, md, rType)
		if err != nil {
			md.Chans.MessageChan <- fmt.Sprintf("ResetApp: %s: %s: Resetting market params .... | disabled: \"%s\"\n", s, w.user.Username, md.Data.DisableTransaction)
			log.Printf("%v", err)
			continue
		}
		md = mdHolder
		break
	}
	if rType == "sell" || rType == "all" {
		if !isClosed(w.profitPriceResetAChan) {
			md.Chans.MessageChan <- fmt.Sprintf("ResetApp: %s: %s: Resetting profitPrice pendingA .... | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
			w.profitPriceResetAChan <- true
		}
	}
	if rType == "buy" || rType == "all" {
		if !isClosed(w.profitPriceResetBChan) {
			md.Chans.MessageChan <- fmt.Sprintf("ResetApp: %s: %s: Resetting profitPrice pendingB .... | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
			w.profitPriceResetBChan <- true
		}
	}
	md.Chans.MessageChan <- fmt.Sprintf("ResetApp: %s: %s: Resetting market params .... | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
	<-respChan
	md.Chans.MessageChan <- fmt.Sprintf("ResetApp: %s: %s: App Resetting Completed: | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
	sync <- true
	return nil
}

//MarketTrading is a trading algorithm provided by the WorkerAppService
func (w WorkerAppService) marketTrading(md *App) {
	var (
		ema5515Diff, sureTradeMUp, sureTradeMDown                      float64
		prevBuyLPoint, prevSellLPoint, lastPrice, gapPrice             float64
		profitOrLost                                                   float64
		ema55Market, ema15Market, floatHolder                          float64
		tkr                                                            *api.Ticker
		err                                                            error
		e                                                              EmaData
		ErrorMessage, MarketStatus, stringHolder                       string
		ordrStatus, profitOrLostStatus, pastNeverSold, pastNeverBought string
		wait                                                           time.Duration
	)
	gapPrice = md.Data.TickSize * md.Data.TrailPoints //to clear error region
	respChan := make(chan bool)
	prevBuyLPoint = math.MaxFloat64
	i := 0.0
	j := 0.0
	sync := make(chan bool, 1)
	for {
		select {
		case <-md.Chans.CloseDown:
			if !isClosed(w.profitPriceResetAChan) {
				md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: profitPriceFunc Shuting Down A: | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
				w.profitPriceResetAChan <- false
			}
			if !isClosed(w.profitPriceResetBChan) {
				md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: profitPriceFunc Shuting Down B: | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
				w.profitPriceResetBChan <- false
			}
			if !isClosed(md.Chans.StopSellPriceTradingChan) {
				close(md.Chans.StopSellPriceTradingChan)
				time.Sleep(time.Millisecond * 3)
				if !strings.Contains(md.Data.Message, "App Shuting Down") || !strings.Contains(md.Data.Message, "PriceTrading") {
					log.Printf("last messages at Shutdown: %s\n", md.Data.Message)
					continue
				}
			}
			if !isClosed(md.Chans.StopBuyPriceTradingChan) {
				close(md.Chans.StopBuyPriceTradingChan)
				time.Sleep(time.Millisecond * 3)
				if !strings.Contains(md.Data.Message, "App Shuting Down") || !strings.Contains(md.Data.Message, "PriceTrading") {
					log.Printf("last messages at Shutdown: %s\n", md.Data.Message)
					continue
				}
			}
			md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: App Shuting Down: | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
			return
		case respChan = <-md.Chans.MarketResetInfoChan:
			md.Data.ProfitPrice = 0.0
			md.Data.ProfitPriceUsed = ""
			md.Data.PriceTradingStarted = ""
			md.Data.MainStartPointSell = 0.0
			md.Data.SoldQuantity = 0.0
			md.Data.BoughtQuantity = 0.0
			md.Data.MainStartPointBuy = 0.0
			md.Data.MainQuantity = 0.0
			md.Data.NextStartPointPrice = 0.0
			md.Data.HodlerQuantity = 0.0
			respChan <- true
			respChan <- true
		default:
			if strings.Contains(md.Data.DisableTransaction, "market") || strings.Contains(md.Data.DisableTransaction, "all trades") {
				time.Sleep(time.Second * 300)
				continue

			}
		}
		e = w.GetEmaFunc(md)
		ema55Market = e.Ema55
		ema15Market = e.Ema15
		ema5515Diff = math.Abs(ema55Market - ema15Market)
		//error region
		if (gapPrice > ema5515Diff) && (ema5515Diff > 0.0 && ema5515Diff < math.MaxFloat64) && (gapPrice > 0.0 && gapPrice < math.MaxFloat64) {
			md.Chans.MessageChan <- fmt.Sprintf("Error Region, because ema5515diff is less than gapPrice, so no Bussiness/n")
			wait = 30000
			waiting(wait)
			gapPrice = md.Data.TickSize * md.Data.TrailPoints //to clear error region
			continue
		}
		for {
			tkr, err = w.API.GetTicker(md.Data.SymbolCode)
			if err != nil {
				ErrorMessage = fmt.Sprintf("%v", err)
				md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: Unable to get Ticker | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
				if strings.Contains(ErrorMessage, "502 Bad Gateway") || strings.Contains(ErrorMessage, "503 Service Unavailable") || strings.Contains(ErrorMessage, "500 Internal Server") || strings.Contains(ErrorMessage, "Exchange temporary closed") {
					time.Sleep(time.Millisecond * time.Duration(30000+<-w.Rand800))
				} else {
					time.Sleep(time.Second * 5)
				}
				continue
			} else if err == nil {
				lastPrice, _ = strconv.ParseFloat(tkr.Last, 64)
				md.Chans.MessageChan <- fmt.Sprintf("got lastPrice = %.8f", lastPrice)
				break
			}
		}
		//intaol MarketStatus
		MarketStatus = ""
		if ema15Market > ema55Market {
			MarketStatus = "Up"
			md.Chans.SetParamChan <- SetParam{"StopLostPoint", lastPrice - gapPrice}
			md.Chans.SetParamChan <- SetParam{"", 0}
			md.Chans.SetParamChan <- SetParam{"", 0}
			if md.Data.StopLostPoint < prevSellLPoint {
				md.Chans.SetParamChan <- SetParam{"StopLostPoint", prevSellLPoint}
				md.Chans.SetParamChan <- SetParam{"", 0}
				md.Chans.SetParamChan <- SetParam{"", 0}
			} else {
				prevSellLPoint = md.Data.StopLostPoint
			}
		} else if ema55Market > ema15Market {
			MarketStatus = "Down"
			md.Chans.SetParamChan <- SetParam{"StopLostPoint", lastPrice + gapPrice}
			md.Chans.SetParamChan <- SetParam{"", 0}
			md.Chans.SetParamChan <- SetParam{"", 0}
			if md.Data.StopLostPoint > prevBuyLPoint {
				md.Chans.SetParamChan <- SetParam{"StopLostPoint", prevBuyLPoint}
				md.Chans.SetParamChan <- SetParam{"", 0}
				md.Chans.SetParamChan <- SetParam{"", 0}
			} else {
				prevBuyLPoint = md.Data.StopLostPoint
			}
		} else if md.Data.StopLostPoint == 0.0 {
			md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: GlobalInfor: Market not decided, ema5515Diff = %.8f: gapPrice = %.8f: Side = %s LastPrice = %.8f StopLostPoint = %.8f md.Data.NextMarketBuyPoint = %.8f and md.Data.NextMarketSellPoint = %.8f NeverSold = %s NeverBought = %s /n", MarketStatus, md.Data.SymbolCode, w.user.Username, ema5515Diff, gapPrice, md.Data.Side, lastPrice, md.Data.StopLostPoint, md.Data.NextMarketBuyPoint, md.Data.NextMarketSellPoint, md.Data.NeverSold, md.Data.NeverBought)
			md.Chans.SetParamChan <- SetParam{"", 0}
			md.Chans.SetParamChan <- SetParam{"", 0}
			prevSellLPoint = 0.0
			prevBuyLPoint = math.MaxFloat64
			gapPrice = md.Data.TickSize * md.Data.TrailPoints //to clear error region
			continue
		} else {
			md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: GlobalInfor: Market not decided, ema5515Diff = %.8f: gapPrice = %.8f: Side = %s LastPrice = %.8f StopLostPoint = %.8f md.Data.NextMarketBuyPoint = %.8f and md.Data.NextMarketSellPoint = %.8f NeverSold = %s NeverBought = %s /n", MarketStatus, md.Data.SymbolCode, w.user.Username, ema5515Diff, gapPrice, md.Data.Side, lastPrice, md.Data.StopLostPoint, md.Data.NextMarketBuyPoint, md.Data.NextMarketSellPoint, md.Data.NeverSold, md.Data.NeverBought)
			md.Chans.SetParamChan <- SetParam{"", 0}
			md.Chans.SetParamChan <- SetParam{"", 0}
			gapPrice = md.Data.TickSize * md.Data.TrailPoints //to clear error region
			continue
		}

		//General message
		md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: GlobalInfo  before market decission: Side = %s LastPrice = %.8f StopLostPoint = %.8f md.Data.NextMarketBuyPoint = %.8f and md.Data.NextMarketSellPoint = %.8f NeverSold = %s NeverBought = %s /n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side, lastPrice, md.Data.StopLostPoint, md.Data.NextMarketBuyPoint, md.Data.NextMarketSellPoint, md.Data.NeverSold, md.Data.NeverBought)
		md.Chans.SetParamChan <- SetParam{"", 0}
		md.Chans.SetParamChan <- SetParam{"", 0}

		//Perform sell or buy decition making
		if ema15Market > ema55Market { //market is going up: !!!!!!!!!!!!!!!!!!
			MarketStatus = "Up"
			//Market: Up Should Bu
			sureTradeMDown = 0.0
			md.Chans.SetParamChan <- SetParam{"Side", "buy"}
			md.Chans.SetParamChan <- SetParam{"", 0}
			md.Chans.SetParamChan <- SetParam{"", 0}
			md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: go ahead buy StopLostPoint = %.8f md.Data.NextMarketBuyPoint = %.8f and md.Data.NextMarketSellPoint = %.8f /n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.StopLostPoint, md.Data.NextMarketBuyPoint, md.Data.NextMarketSellPoint)
			md.Chans.SetParamChan <- SetParam{"", 0}
			md.Chans.SetParamChan <- SetParam{"", 0}
			//Market Up but StopLostPoint is Reached
			if (lastPrice < md.Data.StopLostPoint) && (md.Data.StopLostPoint > 0.0 && md.Data.StopLostPoint < math.MaxFloat64) { //has crossed stop lost so sell
				//Market: Up: StopLost crossing: Don't Sell Below NextMarketSellPointflter 1
				if (md.Data.NextMarketSellPoint > lastPrice) && (md.Data.NextMarketSellPoint > 0.0 && md.Data.NextMarketSellPoint < math.MaxFloat64) {
					md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: StopLost %.8f crossing: Don't Sell Below NextMarketSellPoint %.8f at lastPrice %.8f | disabled: \"%s\"\n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.StopLostPoint, md.Data.NextMarketSellPoint, lastPrice, md.Data.DisableTransaction)
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
					wait = 3000
					waiting(wait)
					continue
				}
				//filter4
				if sureTradeMUp < md.Data.SureTradeFactor {
					md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: SreTradeFactor %.2f not reached | StopLost is %.8f  at NextMarketSellPoint %.8f at lastPrice %.8f | disabled: \"%s\"\n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.SureTradeFactor, md.Data.StopLostPoint, md.Data.NextMarketSellPoint, lastPrice, md.Data.DisableTransaction)
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
					sureTradeMUp++
					continue
				} else {
					sureTradeMUp = 0.0
				}
				md.Chans.SetParamChan <- SetParam{"Side", "sell"}
				md.Chans.SetParamChan <- SetParam{"", 0}
				md.Chans.SetParamChan <- SetParam{"", 0}
				md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: StoplostCross Side is sell at lastPrice = %.8f StopLostPoint = %.8f md.Data.NextMarketBuyPoint = %.8f and md.Data.NextMarketSellPoint = %.8f /n", MarketStatus, md.Data.SymbolCode, w.user.Username, lastPrice, md.Data.StopLostPoint, md.Data.NextMarketBuyPoint, md.Data.NextMarketSellPoint)
			}
		} else if ema15Market < ema55Market { //market is going down !!!!!!!!!!!!!!!!!!!
			MarketStatus = "Down"
			//Markwt Down should sell
			sureTradeMUp = 0.0
			md.Chans.SetParamChan <- SetParam{"Side", "sell"}
			md.Chans.SetParamChan <- SetParam{"", 0}
			md.Chans.SetParamChan <- SetParam{"", 0}
			md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: go ahead sell StopLostPoint = %.8f md.Data.NextMarketBuyPoint = %.8f and md.Data.NextMarketSellPoint = %.8f /n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.StopLostPoint, md.Data.NextMarketBuyPoint, md.Data.NextMarketSellPoint)
			//Market Down: but stoplostPoint is Reached
			if (lastPrice > md.Data.StopLostPoint) && (md.Data.StopLostPoint > 0.0 && md.Data.StopLostPoint < math.MaxFloat64) { //has crossed stop lost
				//filter 2
				if (md.Data.NextMarketBuyPoint < lastPrice) && (md.Data.NextMarketBuyPoint > 0.0 && md.Data.NextMarketBuyPoint < math.MaxFloat64) {
					md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: StopLost %.8f crossing: Don't Buy above NextMarketBuyPoint %.8f at lastPrice %.8f | disabled: \"%s\"\n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.StopLostPoint, md.Data.NextMarketBuyPoint, lastPrice, md.Data.DisableTransaction)
					wait = 3000
					waiting(wait)
					continue
				}
				//filter4
				if sureTradeMDown < md.Data.SureTradeFactor {
					md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: SreTradeFactor %.2f not reached | StopLost is %.8f  at NextMarketSellPoint %.8f at lastPrice %.8f | disabled: \"%s\"\n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.SureTradeFactor, md.Data.StopLostPoint, md.Data.NextMarketSellPoint, lastPrice, md.Data.DisableTransaction)
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
					sureTradeMDown++
					continue
				} else {
					sureTradeMDown = 0.0
				}
				md.Chans.SetParamChan <- SetParam{"Side", "buy"}
				md.Chans.SetParamChan <- SetParam{"", 0}
				md.Chans.SetParamChan <- SetParam{"", 0}
				md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: StoplostCross Side is buy at lastPrice = %.8f StopLostPoint = %.8f md.Data.NextMarketBuyPoint = %.8f and md.Data.NextMarketSellPoint = %.8f /n", MarketStatus, md.Data.SymbolCode, w.user.Username, lastPrice, md.Data.StopLostPoint, md.Data.NextMarketBuyPoint, md.Data.NextMarketSellPoint)
			}
		}

		//General message
		md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: GlobalInfo after market decision: Side = %s LastPrice = %.8f StopLostPoint = %.8f md.Data.NextMarketBuyPoint = %.8f and md.Data.NextMarketSellPoint = %.8f NeverSold = %s NeverBought = %s /n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side, lastPrice, md.Data.StopLostPoint, md.Data.NextMarketBuyPoint, md.Data.NextMarketSellPoint, md.Data.NeverSold, md.Data.NeverBought)
		md.Chans.SetParamChan <- SetParam{"", 0}
		md.Chans.SetParamChan <- SetParam{"", 0}

		//SellBuyfilter
		if ((md.Data.MrktSellPrice > 0.0 && md.Data.MrktSellPrice < math.MaxFloat64) && md.Data.Side == "buy") || ((md.Data.MrktBuyPrice > 0.0 && md.Data.MrktBuyPrice < math.MaxFloat64) && md.Data.Side == "sell") { //it bought before now sell higher
			md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s:  Entered sellBuyfilter with %s | disabled: \"%s\"\n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side, md.Data.DisableTransaction)
			md.Chans.SetParamChan <- SetParam{"", 0}
			md.Chans.SetParamChan <- SetParam{"", 0}
			if md.Data.Side == "sell" {
				floatHolder = lastPrice - md.Data.TickSize*2
				profitOrLost = (floatHolder - md.Data.MrktBuyPrice) * md.Data.MrktQuantity * (1.0 - md.Data.TakeLiquidityRate)
				if profitOrLost <= 0.0 { //Lost to be allowed
					if (md.Data.NeverSold == "yes" && MarketStatus == "Down") && profitOrLost < 0.0 {
						md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s: NeverSold so go aheah and make it to jackport side is %s", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side)
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.SetParamChan <- SetParam{"", 0}
					} else if (md.Data.NeverBought == "yes" && MarketStatus == "Up") && profitOrLost < 0.0 {
						md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s: %s: so go aheah and make it to jackport side is %s profitOrLost = %.8f/n", MarketStatus, md.Data.SymbolCode, w.user.Username, stringHolder, md.Data.Side, profitOrLost)
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.SetParamChan <- SetParam{"", 0}
					} else {
						wait = 3000
						waiting(wait)
						md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s: if allowed to %s will make lost so no need to allow", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side)
						continue //if allowed will make lost so no need to allow
					}
				} else {
					md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s: go aheah and make it to jackport side is %s", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side)
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
				}
			} else if md.Data.Side == "buy" {
				floatHolder = lastPrice + md.Data.TickSize*2
				profitOrLost = (md.Data.MrktSellPrice - floatHolder) * md.Data.MrktQuantity * (1.0 - md.Data.TakeLiquidityRate)
				if profitOrLost <= 0.0 { //lost you want to allow but same is gone or lost
					if (md.Data.NeverBought == "yes" && MarketStatus == "Up") && (profitOrLost < 0.0) {
						md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s: %s: so go aheah and make it to jackport side is %s profitOrLost = %.8f/n", MarketStatus, md.Data.SymbolCode, w.user.Username, stringHolder, md.Data.Side, profitOrLost)
					} else if (md.Data.NeverSold == "yes" && MarketStatus == "Down") && (profitOrLost < 0.0) {
						stringHolder = "NeverSold"
						md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s: %s: so go aheah and make it to jackport side is %s profitOrLost = %.8f/n", MarketStatus, md.Data.SymbolCode, w.user.Username, stringHolder, md.Data.Side, profitOrLost)
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.SetParamChan <- SetParam{"", 0}
					} else {
						wait = 3000
						waiting(wait)

						md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s:  if allowed to %s will make lost so no need to allow profitOrLost = %.8f/n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side, profitOrLost)
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						continue //if allowed will make lost so no need to allow
					}
				} else {
					md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s:  go aheah and make it to jackport side is %s /n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side)
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
				}
			} else {
				panic("Side not decided")
			}
		} else {
			if (profitOrLost > 0.0 && profitOrLost < math.MaxFloat64) && (md.Data.NeverBought == "yes" && MarketStatus == "Up") && (md.Data.MrktSellPrice <= 0 || md.Data.MrktSellPrice >= math.MaxFloat64) {
				md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s: NeverBought so go aheah and make it to jackport side is %s /n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side)
				md.Chans.SetParamChan <- SetParam{"", 0}
				md.Chans.SetParamChan <- SetParam{"", 0}
			} else if (profitOrLost > 0.0 && profitOrLost < math.MaxFloat64) && (md.Data.NeverSold == "yes" && MarketStatus == "Down") && (md.Data.MrktBuyPrice <= 0 || md.Data.MrktBuyPrice >= math.MaxFloat64) {
				md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s: NeverSold so go aheah and make it to jackport side is %s", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side)
				md.Chans.SetParamChan <- SetParam{"", 0}
				md.Chans.SetParamChan <- SetParam{"", 0}
			} else if i >= 5.0 && (profitOrLost == 0.0) && j <= 0.0 {
				j++
				md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s: i >= %.8f NeverS go aheah and make it to jackport side is %s", MarketStatus, md.Data.SymbolCode, w.user.Username, i, md.Data.Side)
				md.Chans.SetParamChan <- SetParam{"", 0}
				md.Chans.SetParamChan <- SetParam{"", 0}
			} else {
				i++
				md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s: Repaet: failed to decide Market profitOrLost = %.8f NeverBought = %s, NeverSold = %s, Side = %s, md.Data.MrktBuyPrice = %.8f,md.Data.MrktSellPrice = %.8f /n", MarketStatus, md.Data.SymbolCode, w.user.Username, profitOrLost, md.Data.NeverBought, md.Data.NeverSold, md.Data.Side, md.Data.MrktBuyPrice, md.Data.MrktSellPrice)
				md.Chans.SetParamChan <- SetParam{"", 0}
				md.Chans.SetParamChan <- SetParam{"", 0}
				continue
			}
		}

		if md.Data.Side == "buy" || md.Data.Side == "sell" {
			md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s:  Entered jackpot!!! with %s | disabled: \"%s\"\n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side, md.Data.DisableTransaction)
			md.Chans.SetParamChan <- SetParam{"", 0}
			md.Chans.SetParamChan <- SetParam{"", 0}
			//Mareket Decieded to make Order
			floatHolder, err = w.jackpotOrderQuantity(md, lastPrice, md.Data.GoodBiz, "MarketTrading")
			md.Chans.SetParamChan <- SetParam{"MrktQuantity", floatHolder}
			md.Chans.SetParamChan <- SetParam{"", 0}
			md.Chans.SetParamChan <- SetParam{"", 0}
			if err != nil {
				//MarKrt Decieded failed jackport
				if strings.Contains(fmt.Sprintf("%v", err), "below QuantityIncrement") {
					md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: Unable to get %s jackpot: The balance was below QuantityIncrement | profitOrLost = %.8f NeverBought = %s NeverSold = %s | disabled: \"%s\"\n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side, profitOrLost, md.Data.NeverBought, md.Data.NeverSold, md.Data.DisableTransaction)
				} else {
					md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: Unable to get %s %.8f jackpot: unable to connect to exchange | profitOrLost = %.8f NeverBought = %s NeverSold = %s | disabled: \"%s\"\n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side, floatHolder, profitOrLost, md.Data.NeverBought, md.Data.NeverSold, md.Data.DisableTransaction)
				}
				time.Sleep(time.Millisecond * 30000)
				continue
			} else {
				md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: Able to get %s %.8f jackpot: | disabled: \"%s\"\n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side, floatHolder, md.Data.DisableTransaction)
				md.Chans.SetParamChan <- SetParam{"", 0}
				md.Chans.SetParamChan <- SetParam{"", 0}
				if (profitOrLost > 0.0 && profitOrLost < math.MaxFloat64) && (md.Data.NeverBought == "yes" && MarketStatus == "Up") {
					md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s: NeverBought so go aheah and make it to Order side is %s /n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side)
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
				} else if (profitOrLost > 0.0 && profitOrLost < math.MaxFloat64) && (md.Data.NeverSold == "yes" && MarketStatus == "Down") {
					md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s: NeverSold so go aheah and make it to Order side is %s", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side)
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
				} else if i >= 5.0 && (profitOrLost == 0.0) && j <= 1.0 {
					md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s: i = %.8f so go aheah and make it to Order side is %s", MarketStatus, md.Data.SymbolCode, w.user.Username, i, md.Data.Side)
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
				} else {
					md.Chans.MessageChan <- fmt.Sprintf("Market %s: %s: %s: failed to Order and will Repeat: i=%.8f Side = %s, profitOrLost = %.8f, md.Data.MrktBuyPrice = %.8f,md.Data.MrktSellPrice = %.8f /n", MarketStatus, md.Data.SymbolCode, w.user.Username, i, md.Data.Side, profitOrLost, md.Data.MrktBuyPrice, md.Data.MrktSellPrice)
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
					if i < 5.0 {
						i++
					}
					continue
				}
				ordrStatus, _, floatHolder, err = w.placeOrder(md, md.Data.MrktQuantity, lastPrice, "MarketTrading")
				//Martket Decieded order failied
				if err != nil || !strings.Contains(ordrStatus, "filled") {
					md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: Place %s order error: | disabled: \"%s\"\n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side, md.Data.DisableTransaction)
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
					wait = 3000
					waiting(wait)
					continue
				} else {
					//Market Decieded order success
					md.Chans.SetParamChan <- SetParam{"SuccessfulOrders", 1.0}
					if md.Data.MrktQuantity != floatHolder {
						md.Chans.SetParamChan <- SetParam{"MrktQuantity", floatHolder}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.SetParamChan <- SetParam{"", 0}
					}
					md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: Place order success: %s %.8f amount at %.8f %s | NeverSold = %s NeverBought = %s | disabled: \"%s\"\n", MarketStatus, md.Data.SymbolCode, w.user.Username, md.Data.Side, md.Data.MrktQuantity, lastPrice, ordrStatus, md.Data.NeverSold, md.Data.NeverBought, md.Data.DisableTransaction)
					//Market: Decieded: Successful Buy Order
					if md.Data.Side == "buy" {
						pastNeverBought = md.Data.NeverBought
						md.Chans.SetParamChan <- SetParam{"NeverBought", "no more aply"}
						md.Chans.SetParamChan <- SetParam{"NeverSold", "yes"}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						if profitOrLost > 0.0 {
							profitOrLostStatus = "Profit"
						} else if profitOrLost < 0.0 {
							profitOrLostStatus = "Lost"
						} else if profitOrLost == 0.0 { //profitOrlost is same
							profitOrLostStatus = "same"
						} else {
							panic("same can not get here")
						}
						md.Chans.SetParamChan <- SetParam{"TotalProfit", profitOrLost}
						md.Chans.SetParamChan <- SetParam{"MadeProfitOrders", 1.0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: made %s %.8f: Total Profit as at now is %.8f | disabled: \"%s\"\n", MarketStatus, md.Data.SymbolCode, w.user.Username, profitOrLostStatus, profitOrLost, md.Data.TotalProfit, md.Data.DisableTransaction)
						if !isClosed(md.Chans.StopBuyPriceTradingChan) || (pastNeverBought == "yes") {
							prevSellLPoint = 0.0
							prevBuyLPoint = math.MaxFloat64
							j = 0.0
							go w.ResetApp(md, "buy", sync)
							<-sync
							continue
						}
						md.Chans.SetParamChan <- SetParam{"MrktBuyPrice", lastPrice}
						md.Chans.SetParamChan <- SetParam{"AlternateData", lastPrice}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						if !isClosed(md.Chans.StopBuyPriceTradingChan) {
							close(md.Chans.StopBuyPriceTradingChan)
						}
						if isClosed(md.Chans.StopSellPriceTradingChan) {
							md.Chans.StopSellPriceTradingChan = make(chan bool)
							go w.priceTrading(md, "")
						} else {
							md.Chans.PriceTradingNextStartChan <- priceTradingVehicle{lastPrice, md.Data.MrktQuantity}
						}
						//Market: Decieded: Successful Sell Order
					} else if md.Data.Side == "sell" {
						pastNeverSold = md.Data.NeverSold
						md.Chans.SetParamChan <- SetParam{"NeverSold", "no more aply"}
						md.Chans.SetParamChan <- SetParam{"NeverBought", "yes"}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						if profitOrLost > 0.0 {
							profitOrLostStatus = "Profit"
						} else if profitOrLost < 0.0 {
							profitOrLostStatus = "Lost"
						} else if i >= 1.0 && (profitOrLost == 0.0) { //profitOrlost is same
							profitOrLostStatus = "same"
						} else { //profitOrLost is same
							panic("it should not get here proftOrLost is same")
						}
						md.Chans.SetParamChan <- SetParam{"TotalProfit", profitOrLost}
						md.Chans.SetParamChan <- SetParam{"MadeProfitOrders", 1.0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: %s: made %s %.8f: Total Profit as at now is %.8f | disabled: \"%s\"\n", MarketStatus, md.Data.SymbolCode, w.user.Username, profitOrLostStatus, profitOrLost, md.Data.TotalProfit, md.Data.DisableTransaction)
						if !isClosed(md.Chans.StopSellPriceTradingChan) || (pastNeverSold == "yes") {
							prevSellLPoint = 0.0
							prevBuyLPoint = math.MaxFloat64
							j = 0.0
							go w.ResetApp(md, "sell", sync)
							<-sync
							time.Sleep(time.Millisecond * 3000)
							continue
						}
						md.Chans.SetParamChan <- SetParam{"MrktSellPrice", lastPrice} //Will be used for profit calculation
						md.Chans.SetParamChan <- SetParam{"AlternateData", lastPrice}
						md.Chans.SetParamChan <- SetParam{"SpinOutReason", "normalActivity"}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						if !isClosed(md.Chans.StopSellPriceTradingChan) {
							close(md.Chans.StopSellPriceTradingChan)
						}
						if isClosed(md.Chans.StopBuyPriceTradingChan) {
							md.Chans.StopBuyPriceTradingChan = make(chan bool)
							go w.priceTrading(md, "")
						} else {
							md.Chans.PriceTradingNextStartChan <- priceTradingVehicle{lastPrice, md.Data.MrktQuantity}
						}
					}
				}
			}
		}
		i = 0.0
	} //end of forloop
}

//PriceTrading ...
func (w WorkerAppService) priceTrading(md *App, toChange string) {
	var (
		tkr                                                   *api.Ticker
		ErrorMessage, ordrStatus                              string
		profitOrLost, floatHolder, floatHolder2, floatHolder3 float64
		lastPrice                                             float64
		profitPoint                                           float64
		profitOrLostStatus                                    string
		err                                                   error
		nsp                                                   priceTradingVehicle
	)
	sync := make(chan bool, 1)
	if md.Data.Side == "buy" {
		md.Data.PriceTradingStarted = "sell"
		defer func() {
			md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Ended the sell of %.8f %s asset bought at %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.MainQuantity, md.Data.SymbolCode, md.Data.MainStartPointSell, md.Data.DisableTransaction)
			md.Data.PriceTradingStarted = ""
		}()
		if !md.FromVersionUpdate {
			md.Data.MainStartPointSell = md.Data.AlternateData
			md.Data.MainQuantity = md.Data.MrktQuantity
			md.Data.NextStartPointPrice = md.Data.MainStartPointSell
			md.Data.HodlerQuantity = 0.0
			if !isClosed(w.profitPriceResetAChan) {
				md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Resetting profitPrice pendingA .... | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
				w.profitPriceResetAChan <- true
			}
			md.FromVersionUpdate = false
		}
		i := 0
		for {
			select {
			case <-md.Chans.StopSellPriceTradingChan:
				md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: App Shuting Down: Ended the sell of bought asset approptly !!!!! | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
				if !isClosed(w.profitPriceResetAChan) {
					md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: Resetting profitPrice pendingA .... | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
					w.profitPriceResetAChan <- true
				}
				return
			case nsp = <-md.Chans.PriceTradingNextStartChan:
				md.Data.NextStartPointPrice = nsp.LastPrice
				md.Data.MainQuantity += nsp.PriceTradingStartQuantity
				md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: md.Data.MainQuantity is updated to %.8f | nextStartPoint updated to %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.MainQuantity, nsp.LastPrice, md.Data.DisableTransaction)

				continue
			default:
				if strings.Contains(md.Data.DisableTransaction, "price") || strings.Contains(md.Data.DisableTransaction, "all trades") {
					time.Sleep(time.Second * 300)

					continue
				}
				profitPoint = md.Data.MainStartPointSell * md.Data.LeastProfitMargin
				for {
					tkr, err = w.API.GetTicker(md.Data.SymbolCode)
					if err != nil {
						ErrorMessage = fmt.Sprintf("%v", err)
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Unable to get Ticker | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
						if strings.Contains(ErrorMessage, "502 Bad Gateway") || strings.Contains(ErrorMessage, "503 Service Unavailable") || strings.Contains(ErrorMessage, "500 Internal Server") || strings.Contains(ErrorMessage, "Exchange temporary closed") {
							time.Sleep(time.Millisecond * time.Duration(20000+<-w.Rand800))
						} else {
							time.Sleep(time.Second * 20)
						}
						continue
					}
					break
				}
				lastPrice, _ = strconv.ParseFloat(tkr.Last, 64)
				if toChange == "NeverSold" {
					md.Chans.SetParamChan <- SetParam{"NeverSold", "no more aply"}
					md.Chans.SetParamChan <- SetParam{"MrktBuyPrice", lastPrice}
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: started to sell %.8f %s asset bought at %.8f | md.Data.NeverSold changed to %s | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.MainQuantity, md.Data.SymbolCode, md.Data.MainStartPointSell, md.Data.NeverSold, md.Data.DisableTransaction)
					md.Chans.SetParamChan <- SetParam{"", 0}
					toChange = ""
					i++
				} else if i == 0 {
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: started to sell %.8f %s asset bought at %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.MainQuantity, md.Data.SymbolCode, md.Data.MainStartPointSell, md.Data.DisableTransaction)
					i++
				}
				//PTSell Sell profit/investment
				if (math.Abs(md.Data.NextStartPointPrice-lastPrice) > profitPoint) && (md.Data.NextStartPointPrice < lastPrice) && (md.Data.MainStartPointSell < lastPrice) {
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Sell ProfitPoint crossed, now to try sell of bought asset at %.8f usd | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, lastPrice, md.Data.DisableTransaction)
					profitOrLost = ((lastPrice - md.Data.TickSize*2) - md.Data.MainStartPointSell) * md.Data.QuantityIncrement * (1.0 - md.Data.TakeLiquidityRate)
					if profitOrLost <= 0.0 && profitOrLost >= math.MaxFloat64 {
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: will make lost so repeat | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
						continue
					}
					ordrStatus, _, floatHolder3, err = w.placeOrder(md, md.Data.QuantityIncrement, lastPrice, "PriceTradingSell")
					if err != nil || !strings.Contains(ordrStatus, "filled") {
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Unable to place sell order: %.8f at %.8f status: %s | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
						if ordrStatus == "Insufficient funds" {
							floatHolder, err = w.jackpotOrderQuantity(md, lastPrice, md.Data.GoodBiz, "PriceTradingSell")
							if err == nil {
								profitOrLost = ((lastPrice - md.Data.TickSize*2) - md.Data.MainStartPointSell) * floatHolder * (1.0 - md.Data.TakeLiquidityRate)
								if profitOrLost <= 0.0 && profitOrLost >= math.MaxFloat64 {
									md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: will make lost so repeat | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
									continue
								}
								ordrStatus, _, floatHolder3, err = w.placeOrder(md, floatHolder, lastPrice, "PriceTradingSell")
								if err != nil || !strings.Contains(ordrStatus, "filled") {
									md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: SelfProfit Still Unable to place sell order: %.8f at %.8f status: %s | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
									time.Sleep(time.Second * 20)
									continue
								}
								lastPrice = lastPrice - md.Data.TickSize*2
								md.Chans.SetParamChan <- SetParam{"SuccessfulOrders", 1.0}
								md.Data.SoldQuantity += floatHolder3
								md.Data.MainQuantity -= floatHolder3
								md.Chans.SetParamChan <- SetParam{"", 0}
								md.Chans.SetParamChan <- SetParam{"", 0}
								md.Data.NextStartPointPrice = lastPrice
								md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Place order success: sell %.8f amount at %.8f %s| disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
								if profitOrLost > 0.0 && profitOrLost < math.MaxFloat64 {
									profitOrLostStatus = "Profit"
									md.Chans.SetParamChan <- SetParam{"TotalProfit", profitOrLost}
									md.Chans.SetParamChan <- SetParam{"MadeProfitOrders", 1.0}
									md.Chans.SetParamChan <- SetParam{"", 0}
									md.Chans.SetParamChan <- SetParam{"", 0}
									md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: made %s %.8f: Total Profit as at now is %.8f MainQuantity remaining %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, profitOrLostStatus, profitOrLost, md.Data.TotalProfit, md.Data.MainQuantity, md.Data.DisableTransaction)
									if md.Data.MainQuantity <= 0.0 {
										md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Resetting .... | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
										w.ResetApp(md, "all", sync)
										<-sync
										continue

									}
								}
							}
						}
						time.Sleep(time.Second * 20)
						continue
					}
					lastPrice = lastPrice - md.Data.TickSize*2
					md.Chans.SetParamChan <- SetParam{"SuccessfulOrders", 1.0}
					md.Data.SoldQuantity += floatHolder3
					md.Data.MainQuantity -= floatHolder3
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Data.NextStartPointPrice = lastPrice
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Place order success: sell %.8f amount at %.8f %s| disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
					if profitOrLost > 0.0 && profitOrLost < math.MaxFloat64 {
						profitOrLostStatus = "Profit"
						md.Chans.SetParamChan <- SetParam{"TotalProfit", profitOrLost}
						md.Chans.SetParamChan <- SetParam{"MadeProfitOrders", 1.0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: made %s %.8f: Total Profit as at now is %.8f MainQuantity remaining %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, profitOrLostStatus, profitOrLost, md.Data.TotalProfit, md.Data.MainQuantity, md.Data.DisableTransaction)
						if md.Data.MainQuantity <= 0.0 {
							md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Resetting .... | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
							w.ResetApp(md, "all", sync)
							<-sync
							continue
						}

					}
					//PTSell Negative Buy investment
				} else if (md.Data.MainStartPointSell >= md.Data.NextStartPointPrice) && (math.Abs(md.Data.NextStartPointPrice-lastPrice) > profitPoint) && (md.Data.NextStartPointPrice > lastPrice) {
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Negative Buy ProfitPoint crossed, now to try buy of more bought asset at %.8f usd | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, lastPrice, md.Data.DisableTransaction)
					profitOrLost = ((lastPrice + md.Data.TickSize*2) - md.Data.NextStartPointPrice) * md.Data.MainQuantity * (1.0 - md.Data.TakeLiquidityRate)
					ordrStatus, _, floatHolder3, err = w.placeOrder(md, md.Data.QuantityIncrement, lastPrice, "PriceTradingBuy")
					if err != nil || !strings.Contains(ordrStatus, "filled") {
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Negative Unable to place buy order: %.8f status: %s | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, ordrStatus, md.Data.DisableTransaction)
						time.Sleep(time.Second * 20)
						continue
					}
					lastPrice = lastPrice + md.Data.TickSize*2
					md.Chans.SetParamChan <- SetParam{"SuccessfulOrders", 1.0}
					profitOrLostStatus = "Lost"
					md.Chans.SetParamChan <- SetParam{"TotalLost", profitOrLost}
					md.Chans.SetParamChan <- SetParam{"MadeLostOrders", 1.0}
					md.Data.MainQuantity += floatHolder3
					md.Data.HodlerQuantity += floatHolder3
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
					floatHolder = lastPrice + profitPoint
					w.profitPriceOriginatorAChan <- ppPendingData{"", floatHolder}
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: profitPriceOriginator proccessed next profitPrice as %.8f ...  | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder, md.Data.DisableTransaction)
					md.Data.NextStartPointPrice = lastPrice
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Place order success: Negative buy %.8f amount at %.8f %s| disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s:  Negative made %s %.8f: Total Lost as at now is %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, profitOrLostStatus, profitOrLost, md.Data.TotalLost, md.Data.DisableTransaction)
					time.Sleep(time.Second * 2)
					continue
				}
				//PTSell Buy selfProfit
				if (md.Data.SoldQuantity > 0.0) && (math.Abs(md.Data.NextStartPointPrice-lastPrice) > profitPoint) && (md.Data.NextStartPointPrice > lastPrice) && (md.Data.MainStartPointSell < md.Data.NextStartPointPrice) {
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: SelfProfit Buy ProfitPoint crossed, now to try buy of again sold bought asset at %.8f usd | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, lastPrice, md.Data.DisableTransaction)
					if md.Data.Hodler == "No" || md.Data.Hodler == "no" || md.Data.Hodler == "NO" {
						profitOrLost = (md.Data.NextStartPointPrice - (lastPrice + md.Data.TickSize*2)) * md.Data.QuantityIncrement * (1.0 - md.Data.TakeLiquidityRate)
						if profitOrLost <= 0.0 && profitOrLost >= math.MaxFloat64 {
							md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: will make lost so repeat | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
							continue
						}
						ordrStatus, _, floatHolder3, err = w.placeOrder(md, md.Data.QuantityIncrement, lastPrice, "PriceTradingBuy")
						if err != nil || !strings.Contains(ordrStatus, "filled") {
							md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: SelfProfit Unable to place buy order: %.8f status: %s | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, ordrStatus, md.Data.DisableTransaction)
							time.Sleep(time.Second * 20)
							continue
						}
						lastPrice = lastPrice + md.Data.TickSize*2
						md.Chans.SetParamChan <- SetParam{"SuccessfulOrders", 1.0}
						md.Data.SoldQuantity -= floatHolder3
						md.Data.MainQuantity += floatHolder3
						if profitOrLost > 0.0 && profitOrLost < math.MaxFloat64 {
							profitOrLostStatus = "Profit"
							md.Chans.SetParamChan <- SetParam{"TotalProfit", profitOrLost}
							md.Chans.SetParamChan <- SetParam{"MadeProfitOrders", 1.0}
							md.Chans.SetParamChan <- SetParam{"", 0}
							md.Chans.SetParamChan <- SetParam{"", 0}
							md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: SelfProfit Place order success: buy %.8f amount at %.8f %s| disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
							md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: made %s %.8f: Total Profit as at now is %.8f SoldQuantity remaining %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, profitOrLostStatus, profitOrLost, md.Data.TotalProfit, md.Data.SoldQuantity, md.Data.DisableTransaction)
							if md.Data.SoldQuantity <= 0.0 {
								w.ResetApp(md, "all", sync)
								<-sync
								continue
							}
						}
						//for HODLERS
					} else if md.Data.MainStartPointSell > lastPrice {
						profitOrLost = (md.Data.NextStartPointPrice - (lastPrice + md.Data.TickSize*2)) * md.Data.SoldQuantity * (1.0 - md.Data.TakeLiquidityRate)
						if profitOrLost <= 0.0 && profitOrLost >= math.MaxFloat64 {
							md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: will make lost so repeat | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
							continue
						}
						ordrStatus, _, floatHolder3, err = w.placeOrder(md, md.Data.SoldQuantity, lastPrice, "PriceTradingBuy")
						if err != nil || !strings.Contains(ordrStatus, "filled") {
							md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: SelfProfit Unable to place buy order: %.8f status: %s | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, ordrStatus, md.Data.DisableTransaction)
							time.Sleep(time.Second * 20)
							continue
						}
						lastPrice = lastPrice + md.Data.TickSize*2
						md.Chans.SetParamChan <- SetParam{"SuccessfulOrders", 1.0}
						md.Data.SoldQuantity -= floatHolder3
						md.Data.MainQuantity += floatHolder3
						if profitOrLost > 0.0 && profitOrLost < math.MaxFloat64 {
							profitOrLostStatus = "Profit"
							md.Chans.SetParamChan <- SetParam{"TotalProfit", profitOrLost}
							md.Chans.SetParamChan <- SetParam{"MadeProfitOrders", 1.0}
							md.Chans.SetParamChan <- SetParam{"", 0}
							md.Chans.SetParamChan <- SetParam{"", 0}
							md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Reserved Holdling SelfProfit Place order success: buy %.8f amount at %.8f %s| disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
							md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: made %s %.8f: Total Profit as at now is %.8f SoldQuantity remaining %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, profitOrLostStatus, profitOrLost, md.Data.TotalProfit, md.Data.SoldQuantity, md.Data.DisableTransaction)
							if md.Data.SoldQuantity <= 0.0 {
								w.ResetApp(md, "all", sync)
								<-sync
								continue
							}
						}
					} else {
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Reversed Hodling continues, hodled %.8f amount | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.SoldQuantity, md.Data.DisableTransaction)
						time.Sleep(time.Second * 20)
						continue
					}
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: SelfProfit Place order success: buy %.8f amount at %.8f %s| disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
					md.Data.NextStartPointPrice = lastPrice
				}
				//PTSell profitPrice processed
				if (md.Data.HodlerQuantity > 0.0) && (md.Data.MainStartPointSell >= md.Data.NextStartPointPrice) {
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: profitPriceGen proccessing ProfitPrice ...  | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
					md.Chans.SetParamChan <- SetParam{"ProfitPrice", <-w.profitPriceGenAChan}
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: ProfitPrice proccessed from pendng slice as %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.ProfitPrice, md.Data.DisableTransaction)
				}
				//PTSell Negative Sell selfProfit
				if (md.Data.MainStartPointSell >= md.Data.NextStartPointPrice) && (md.Data.ProfitPrice < lastPrice) && (md.Data.HodlerQuantity > 0.0) {
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Negative Sell ProfitPoint crossed, now to try sell of bought asset at %.8f USD | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, lastPrice, md.Data.DisableTransaction)
					if md.Data.Hodler == "No" || md.Data.Hodler == "no" || md.Data.Hodler == "NO" {
						profitOrLost = ((lastPrice - md.Data.TickSize*2) - md.Data.NextStartPointPrice) * md.Data.QuantityIncrement * (1.0 - md.Data.TakeLiquidityRate)
						if profitOrLost <= 0.0 && profitOrLost >= math.MaxFloat64 {
							md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: will make lost so repeat | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
							continue
						}
						ordrStatus, _, floatHolder3, err = w.placeOrder(md, md.Data.QuantityIncrement, lastPrice, "PriceTradingSell")
						if err != nil || !strings.Contains(ordrStatus, "filled") {
							md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Negative Unable to place sell order: %.8f at %.8f status: %s | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
							if ordrStatus == "Insufficient funds" {
								floatHolder, err = w.jackpotOrderQuantity(md, lastPrice, md.Data.GoodBiz, "PriceTradingSell")
								if err == nil {
									profitOrLost = ((lastPrice - md.Data.TickSize*2) - md.Data.NextStartPointPrice) * floatHolder * (1.0 - md.Data.TakeLiquidityRate)
									if profitOrLost <= 0.0 && profitOrLost >= math.MaxFloat64 {
										md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: will make lost so repeat | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
										continue
									}
									ordrStatus, _, floatHolder3, err = w.placeOrder(md, floatHolder, lastPrice, "PriceTradingSell")
									if err != nil || !strings.Contains(ordrStatus, "filled") {
										md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: SelfProfit Still Unable to place sell order: %.8f at %.8f status: %s | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
										time.Sleep(time.Second * 20)
										continue
									}
									lastPrice = lastPrice - md.Data.TickSize*2
									md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: profitPriceOriginator proccessing pendingA resize ...  | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
									w.profitPriceOriginatorAChan <- ppPendingData{"resize", 0.0}
									md.Data.MainQuantity -= floatHolder3
									floatHolder2 = floatHolder3
									md.Data.HodlerQuantity -= floatHolder3
									md.Chans.SetParamChan <- SetParam{"SuccessfulOrders", 1.0}
									profitOrLostStatus = "Profit"
									md.Chans.SetParamChan <- SetParam{"TotalProfit", profitOrLost}
									md.Chans.SetParamChan <- SetParam{"MadeProfitOrders", 1.0}
									md.Chans.SetParamChan <- SetParam{"", 0}
									md.Chans.SetParamChan <- SetParam{"", 0}
									md.Data.NextStartPointPrice = lastPrice
									md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Place order success: Negative sell %.8f amount at %.8f %swith profitPrice of %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder2, lastPrice, ordrStatus, md.Data.ProfitPrice, md.Data.DisableTransaction)
									md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Negative made %s %.8f with, \"%s\" Hodling: Total Profit as at now is %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, profitOrLostStatus, profitOrLost, md.Data.Hodler, md.Data.TotalProfit, md.Data.DisableTransaction)
								}
							}
							time.Sleep(time.Second * 20)
							continue
						}
						lastPrice = lastPrice - md.Data.TickSize*2
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: profitPriceOriginator proccessing pendingA resize ...  | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
						w.profitPriceOriginatorAChan <- ppPendingData{"resize", 0.0}
						profitOrLost = (lastPrice - md.Data.NextStartPointPrice) * md.Data.MainQuantity * (1.0 - md.Data.TakeLiquidityRate)
						md.Data.MainQuantity -= floatHolder3
						floatHolder2 = floatHolder3
						md.Data.HodlerQuantity -= floatHolder3
						//for HODLERS
					} else if md.Data.MainStartPointSell < lastPrice {
						profitOrLost = ((lastPrice - md.Data.TickSize*2) - md.Data.NextStartPointPrice) * md.Data.HodlerQuantity * (1.0 - md.Data.TakeLiquidityRate)
						if profitOrLost <= 0.0 && profitOrLost >= math.MaxFloat64 {
							md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: will make lost so repeat | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
							continue
						}
						ordrStatus, _, floatHolder3, err = w.placeOrder(md, md.Data.HodlerQuantity, lastPrice, "PriceTradingSell")
						if err != nil || !strings.Contains(ordrStatus, "filled") {
							md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Negative Unable to place sell order: %.8f at %.8f status: %s | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
							time.Sleep(time.Second * 20)
							continue
						}
						lastPrice = lastPrice - md.Data.TickSize*2
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: profitPriceOriginator proccessing pendingA emptying ...  | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
						w.profitPriceOriginatorAChan <- ppPendingData{"resize", 9.9}

						md.Data.MainQuantity -= floatHolder3
						floatHolder2 = floatHolder3
						md.Data.HodlerQuantity -= floatHolder3
					} else {
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Negative: Hodling continues, hodled %.8f amount | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.HodlerQuantity, md.Data.DisableTransaction)
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: profitPriceOriginator proccessing pendingA resize ...  | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
						w.profitPriceOriginatorAChan <- ppPendingData{"resize", 0.0}
						time.Sleep(time.Second * 20)
						continue
					}
					md.Chans.SetParamChan <- SetParam{"SuccessfulOrders", 1.0}
					profitOrLostStatus = "Profit"
					md.Chans.SetParamChan <- SetParam{"TotalProfit", profitOrLost}
					md.Chans.SetParamChan <- SetParam{"MadeProfitOrders", 1.0}
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Data.NextStartPointPrice = lastPrice
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Place order success: Negative sell %.8f amount at %.8f %swith profitPrice of %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder2, lastPrice, ordrStatus, md.Data.ProfitPrice, md.Data.DisableTransaction)
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Negative made %s %.8f with, \"%s\" Hodling: Total Profit as at now is %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, profitOrLostStatus, profitOrLost, md.Data.Hodler, md.Data.TotalProfit, md.Data.DisableTransaction)
				}
				time.Sleep(time.Second * 2)
			}
		}
	} else {
		md.Data.PriceTradingStarted = "buy"
		defer func() {
			md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Ended the buy of %.8f %s asset sold at %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.MainQuantity, md.Data.SymbolCode, md.Data.MainStartPointBuy, md.Data.DisableTransaction)
			md.Data.PriceTradingStarted = ""
		}()
		if !md.FromVersionUpdate {
			md.Data.MainStartPointBuy = md.Data.AlternateData
			md.Data.MainQuantity = md.Data.MrktQuantity
			md.Chans.SetParamChan <- SetParam{"", 0}
			md.Chans.SetParamChan <- SetParam{"", 0}
			md.Data.NextStartPointPrice = md.Data.MainStartPointBuy
			if !isClosed(w.profitPriceResetBChan) {
				md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Resetting profitPrice pendingB .... | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
				w.profitPriceResetBChan <- true
			}
			md.FromVersionUpdate = false
		}
		md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: started to buy %.8f %s asset sold at %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.MainQuantity, md.Data.SymbolCode, md.Data.MainStartPointBuy, md.Data.DisableTransaction)
		for {
			select {
			case <-md.Chans.StopBuyPriceTradingChan:
				md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: App Shuting Down: Ended the buy of sold asset approptly !!!!! | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
				if !isClosed(w.profitPriceResetBChan) {
					md.Chans.MessageChan <- fmt.Sprintf("Market: %s: %s: Resetting profitPrice pending B.... | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
					w.profitPriceResetBChan <- true
				}
				return
			case nsp = <-md.Chans.PriceTradingNextStartChan:
				if md.Data.MainStartPointBuy >= nsp.LastPrice {
					md.Data.NextStartPointPrice = nsp.LastPrice
				}
				md.Data.MainQuantity += nsp.PriceTradingStartQuantity
				md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: md.Data.MainQuantity is updated to %.8f | nextStartPoint updated to %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.MainQuantity, nsp.LastPrice, md.Data.DisableTransaction)
				continue
			default:
				if strings.Contains(md.Data.DisableTransaction, "price") || strings.Contains(md.Data.DisableTransaction, "all trades") {
					time.Sleep(time.Second * 300)
					continue
				}
				profitPoint = md.Data.MainStartPointBuy * md.Data.LeastProfitMargin
				for {
					tkr, err = w.API.GetTicker(md.Data.SymbolCode)
					if err != nil {
						ErrorMessage = fmt.Sprintf("%v", err)
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Unable to get Ticker | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
						if strings.Contains(ErrorMessage, "502 Bad Gateway") || strings.Contains(ErrorMessage, "503 Service Unavailable") || strings.Contains(ErrorMessage, "500 Internal Server") || strings.Contains(ErrorMessage, "Exchange temporary closed") {
							time.Sleep(time.Millisecond * time.Duration(20000+<-w.Rand800))
						} else {
							time.Sleep(time.Second * 20)
						}
						continue
					}
					break
				}
				lastPrice, _ = strconv.ParseFloat(tkr.Last, 64)
				//PTBuy Buy profit/investment
				if (math.Abs(md.Data.NextStartPointPrice-lastPrice) > profitPoint && md.Data.NextStartPointPrice > lastPrice) && md.Data.MainStartPointBuy >= md.Data.NextStartPointPrice {
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Buy ProfitPoint crossed, now to try buy of sold asset at %.8f usd | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, lastPrice, md.Data.DisableTransaction)
					profitOrLost = (md.Data.MainStartPointBuy - ((lastPrice + md.Data.TickSize*2) + md.Data.TickSize*2)) * md.Data.QuantityIncrement * (1.0 - md.Data.TakeLiquidityRate)
					if profitOrLost <= 0.0 && profitOrLost >= math.MaxFloat64 {
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: will make lost so repeat | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
						continue
					}
					ordrStatus, _, floatHolder3, err = w.placeOrder(md, md.Data.QuantityIncrement, lastPrice, "PriceTradingBuy")
					if err != nil || !strings.Contains(ordrStatus, "filled") {
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Unable to place buy order: %.8f status: %s | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, ordrStatus, md.Data.DisableTransaction)
						time.Sleep(time.Second * 20)
						continue
					}
					lastPrice = lastPrice + md.Data.TickSize*2
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: Place order success: buy %.8f amount at %.8f %s| disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
					md.Chans.SetParamChan <- SetParam{"SuccessfulOrders", 1.0}
					md.Data.BoughtQuantity += floatHolder3
					md.Data.MainQuantity -= floatHolder3
					floatHolder = lastPrice + profitPoint
					w.profitPriceOriginatorBChan <- ppPendingData{"", floatHolder}
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: profitPriceOriginator proccessed next profitPrice as %.8f ...  | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder, md.Data.DisableTransaction)
					md.Data.NextStartPointPrice = lastPrice
					if profitOrLost > 0.0 && profitOrLost < math.MaxFloat64 {
						profitOrLostStatus = "Profit"
						md.Chans.SetParamChan <- SetParam{"TotalProfit", profitOrLost}
						md.Chans.SetParamChan <- SetParam{"MadeProfitOrders", 1.0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: made %s %.8f: Total Profit as at now is %.8f MainQuantity remaining %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, profitOrLostStatus, profitOrLost, md.Data.TotalProfit, md.Data.MainQuantity, md.Data.DisableTransaction)
						if md.Data.MainQuantity <= 0.0 {
							w.ResetApp(md, "all", sync)
							<-sync
							continue
						}
					}
				}
				//PTBuy profitPrice processed
				if (md.Data.BoughtQuantity > 0.0) && (md.Data.MainStartPointBuy >= md.Data.NextStartPointPrice) {
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: profitPriceGen proccessing ProfitPrice ...  | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
					md.Chans.SetParamChan <- SetParam{"ProfitPrice", <-w.profitPriceGenBChan}
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.SetParamChan <- SetParam{"", 0}
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: ProfitPrice proccessed from pendng slice as %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.ProfitPrice, md.Data.DisableTransaction)
				}
				//PTBuy Sell selfProfit
				if (md.Data.BoughtQuantity > 0.0) && (md.Data.ProfitPrice < lastPrice) && (md.Data.MainStartPointBuy >= md.Data.NextStartPointPrice) {
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: SelfProfit Sell ProfitPoint crossed, now to try sell of again bought sold asset %.8f usd | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, lastPrice, md.Data.DisableTransaction)
					profitOrLost = ((lastPrice - md.Data.TickSize*2) - md.Data.NextStartPointPrice) * md.Data.QuantityIncrement * (1.0 - md.Data.TakeLiquidityRate)
					if profitOrLost <= 0.0 && profitOrLost >= math.MaxFloat64 {
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: will make lost so repeat | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
						continue
					}
					ordrStatus, _, floatHolder3, err = w.placeOrder(md, md.Data.QuantityIncrement, lastPrice, "PriceTradingSell")
					if err != nil || !strings.Contains(ordrStatus, "filled") {
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: SelfProfit Unable to place sell order: %.8f at %.8f status: %s | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
						if ordrStatus == "Insufficient funds" {
							floatHolder, err = w.jackpotOrderQuantity(md, lastPrice, md.Data.GoodBiz, "PriceTradingSell")
							if err == nil {
								profitOrLost = ((lastPrice - md.Data.TickSize*2) - md.Data.NextStartPointPrice) * floatHolder * (1.0 - md.Data.TakeLiquidityRate)
								if profitOrLost <= 0.0 && profitOrLost >= math.MaxFloat64 {
									md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: will make lost so repeat | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
									continue
								}
								ordrStatus, _, floatHolder3, err = w.placeOrder(md, floatHolder, lastPrice, "PriceTradingSell")
								if err != nil || !strings.Contains(ordrStatus, "filled") {
									md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: SelfProfit Still Unable to lace sell order: %.8f status: %s | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, ordrStatus, md.Data.DisableTransaction)
									time.Sleep(time.Second * 20)
									continue
								}
								lastPrice = lastPrice - md.Data.TickSize*2
								md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: profitPriceOriginator proccessing pending resize ...  | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
								w.profitPriceOriginatorBChan <- ppPendingData{"resize", 0.0}
								md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: SelfProfit Place order success: sell %.8f amount at %.8f %s| disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
								md.Chans.SetParamChan <- SetParam{"SuccessfulOrders", 1.0}
								md.Data.BoughtQuantity -= floatHolder3
								md.Data.MainQuantity += floatHolder3
								if profitOrLost > 0.0 && profitOrLost < math.MaxFloat64 {
									profitOrLostStatus = "Profit"
									md.Chans.SetParamChan <- SetParam{"TotalProfit", profitOrLost}
									md.Chans.SetParamChan <- SetParam{"MadeProfitOrders", 1.0}
									md.Chans.SetParamChan <- SetParam{"", 0}
									md.Chans.SetParamChan <- SetParam{"", 0}
									md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: made %s %.8f: Total Profit as at now is %.8f BoughtQuantity remaining %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, profitOrLostStatus, profitOrLost, md.Data.TotalProfit, md.Data.BoughtQuantity, md.Data.DisableTransaction)
									if md.Data.BoughtQuantity <= 0.0 {
										w.ResetApp(md, "all", sync)
										<-sync
										continue
									}
								}
								md.Data.NextStartPointPrice = lastPrice
							}
						}
						time.Sleep(time.Second * 20)
						continue
					}
					lastPrice = lastPrice - md.Data.TickSize*2
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: profitPriceOriginator proccessing pending resize ...  | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.DisableTransaction)
					w.profitPriceOriginatorBChan <- ppPendingData{"resize", 0.0}
					md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: SelfProfit Place order success: sell %.8f amount at %.8f %s| disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, floatHolder3, lastPrice, ordrStatus, md.Data.DisableTransaction)
					md.Chans.SetParamChan <- SetParam{"SuccessfulOrders", 1.0}
					md.Data.BoughtQuantity -= floatHolder3
					md.Data.MainQuantity += floatHolder3
					if profitOrLost > 0.0 && profitOrLost < math.MaxFloat64 {
						profitOrLostStatus = "Profit"
						md.Chans.SetParamChan <- SetParam{"TotalProfit", profitOrLost}
						md.Chans.SetParamChan <- SetParam{"MadeProfitOrders", 1.0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.SetParamChan <- SetParam{"", 0}
						md.Chans.MessageChan <- fmt.Sprintf("PriceTrading: %s: %s: made %s %.8f: Total Profit as at now is %.8f BoughtQuantity remaining %.8f | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, profitOrLostStatus, profitOrLost, md.Data.TotalProfit, md.Data.BoughtQuantity, md.Data.DisableTransaction)
						if md.Data.BoughtQuantity <= 0.0 {
							w.ResetApp(md, "all", sync)
							<-sync
							continue
						}
					}
					md.Data.NextStartPointPrice = lastPrice
				}
			}
			time.Sleep(time.Second * 2)
		}
	}
}

//ProfitPriceAFunc ...
func (w WorkerAppService) ProfitPriceAFunc(md *model.AppData, profitPriceReset chan bool, profitPriceOriginator chan ppPendingData) <-chan float64 {
	profitPriceGen := make(chan float64)
	profitPriceHolder := profitPriceGen
	go func(profitPriceOriginator chan ppPendingData) {
		//log.Printf("ProfitPriceFunc: Goroutine started\n")
		defer log.Printf("ProfitPriceFunc: Goroutine Ended\n")
		lenght := 0
		var a bool
		var dat ppPendingData
		for {
			var last float64
			lenght = len(md.PendingA)
			if lenght > 0 {
				last = md.PendingA[lenght-1]
				profitPriceGen = profitPriceHolder
				//log.Printf("ProfitPriceFunc: processed profitPrice as: %.8f\n", last)
			}
			select {
			case a = <-profitPriceReset:
				if a {
					md.PendingA = []float64{}
					profitPriceGen = nil
					//log.Printf("ProfitPriceFunc: Resetted pending slice\n")
				} else {
					close(w.profitPriceResetAChan)
					return
				}
			case dat = <-profitPriceOriginator:
				if dat.action == "resize" {
					if lenght > 0 && dat.profitPrice != 9.9 {
						md.PendingA = md.PendingA[:lenght-1]
					} else if lenght > 0 {
						md.PendingA = []float64{}
					}
					profitPriceGen = nil
				} else {
					md.PendingA = append(md.PendingA, dat.profitPrice)
				}
			case profitPriceGen <- last:
				profitPriceGen = nil
			}
		}
	}(profitPriceOriginator)
	return profitPriceGen
}

//ProfitPriceBFunc ...
func (w WorkerAppService) ProfitPriceBFunc(md *model.AppData, profitPriceReset chan bool, profitPriceOriginator chan ppPendingData) <-chan float64 {
	profitPriceGen := make(chan float64)
	profitPriceHolder := profitPriceGen
	go func(profitPriceOriginator chan ppPendingData) {
		//log.Printf("ProfitPriceFunc: Goroutine started\n")
		defer log.Printf("ProfitPriceFunc: Goroutine Ended\n")
		lenght := 0
		var a bool
		var dat ppPendingData
		for {
			var last float64
			lenght = len(md.PendingB)
			if lenght > 0 {
				last = md.PendingB[lenght-1]
				profitPriceGen = profitPriceHolder
				//log.Printf("ProfitPriceFunc: processed profitPrice as: %.8f\n", last)
			}
			select {
			case a = <-profitPriceReset:
				if a {
					md.PendingB = []float64{}
					profitPriceGen = nil
					//log.Printf("ProfitPriceFunc: Resetted pending slice\n")
				} else {
					close(w.profitPriceResetBChan)
					return
				}
			case dat = <-profitPriceOriginator:
				if dat.action == "resize" {
					if lenght > 0 && dat.profitPrice != 9.9 {
						md.PendingB = md.PendingB[:lenght-1]
					} else if lenght > 0 {
						md.PendingB = []float64{}
					}
					profitPriceGen = nil
				} else {
					md.PendingB = append(md.PendingB, dat.profitPrice)
				}
			case profitPriceGen <- last:
				profitPriceGen = nil
			}
		}
	}(profitPriceOriginator)
	return profitPriceGen
}

//TrailProfitPointsFunc determines the factor used to calculate profit, gapDiff and gapPrice per symbol
func (w WorkerAppService) trailProfitPointsFunc(md *App) (float64, float64) {
	leastProfit := 0.035
	trailPoints := 53.0
	if strings.Contains("MRSUSD", md.Data.SymbolCode) {
		leastProfit = 0.055
		trailPoints = 70
	} else if strings.Contains("DCNUSD,ZECUSD", md.Data.SymbolCode) {
		leastProfit = 0.045
		trailPoints = 70
	} else if strings.Contains("MANAUSD,REPUSD,LTCUSD,BTCUSD", md.Data.SymbolCode) {
		leastProfit = 0.04
		trailPoints = 65
	} else if strings.Contains("XTZUSD,EOSUSD,QTUMUSD,BTGUSD,LSKUSD,ETHUSD", md.Data.SymbolCode) {
		leastProfit = 0.0375
		trailPoints = 62
	} else if strings.Contains("ZRXUSD,XMRUSD,ADAUSD,KMDUSD,TRXUSD,DASHUSD,NEOUSD,BCHUSD", md.Data.SymbolCode) {
		leastProfit = 0.035
		trailPoints = 58
	} else if strings.Contains("NANO,ETCUSD,BTTUSD,XRPUSDT,XLMUSD", md.Data.SymbolCode) {
		leastProfit = 0.03
		trailPoints = 53.0
	} else {
		log.Printf("trailProfitPointsFunc: %s not yet captured\n", md.Data.SymbolCode)
	}
	return trailPoints, leastProfit
}

//JackpotOrderQuantity is used to determine bulk order like the total balance as order
func (w WorkerAppService) jackpotOrderQuantity(md *App, lastPrice float64, goodbiz float64, caller string) (float64, error) {
	md.Chans.MessageChan <- fmt.Sprintf("JackpotOrderQuantity: %s: %s: to determine bulk %s order quantity | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, md.Data.Side, md.Data.DisableTransaction)
	var (
		order, baseCurrencyBalance, quoteCurrencyBalance, dividend, mid float64
		calOrder                                                        float64
		err                                                             error
		sym                                                             *api.Symbol
	)
	baseCurrencyBalance, quoteCurrencyBalance, err = w.GetTradingBalance(md)
	if err != nil {
		return 0.0, err
	}

	side := md.Data.Side
	if caller == "PriceTradingBuy" {
		side = "buy"
	} else if caller == "PriceTradingSell" {
		side = "sell"
	} else if caller == "AdaptAppSell" {
		side = "sell"
	}
	i := 50.0
	if side == "buy" {
		mid = quoteCurrencyBalance / (lastPrice * (1.0 + 0.003))
		if mid < md.Data.QuantityIncrement {
			order = md.Data.QuantityIncrement
			err = errors.Errorf("The trading balance was below QuantityIncremen")
		} else {
			for {
				calOrder = mid / i
				if calOrder >= goodbiz && calOrder >= md.Data.QuantityIncrement {
					dividend = math.Trunc(calOrder / md.Data.QuantityIncrement)
					order = md.Data.QuantityIncrement * dividend
					break
				}
				if i == 1.0 {
					dividend = math.Trunc(calOrder / md.Data.QuantityIncrement)
					order = md.Data.QuantityIncrement * dividend
					break
				}
				i--
			}
		}
	} else {
		if baseCurrencyBalance >= md.Data.QuantityIncrement {
			dividend = math.Trunc(baseCurrencyBalance / md.Data.QuantityIncrement)
			order = md.Data.QuantityIncrement * dividend
		} else {
			sym, err = w.API.GetSymbol(md.Data.SymbolCode)
			if err != nil {
				order = md.Data.QuantityIncrement
				err = errors.Errorf("The balance was below QuantityIncremen")
			} else if sqty, _ := strconv.ParseFloat(sym.QuantityIncrement, 64); baseCurrencyBalance >= sqty {
				dividend = math.Trunc(baseCurrencyBalance / sqty)
				order = sqty * dividend
			} else {
				order = md.Data.QuantityIncrement
				err = errors.Errorf("The balance was even below Original QuantityIncremen")
			}
		}
	}
	return order, err
}

//PlaceOrder ...
func (w WorkerAppService) placeOrder(md *App, quantity, lastPrice float64, caller string) (string, float64, float64, error) {
	var (
		price                       float64
		clientOrderID, ErrorMessage string
		err                         error
		i                           int
		after3by7Sec                <-chan time.Time
	)
	ordr := &api.Order{}
	side := md.Data.Side
	if caller == "PriceTradingBuy" {
		side = "buy"
	} else if caller == "PriceTradingSell" {
		side = "sell"
	}
	if side == "sell" {
		price = lastPrice - md.Data.TickSize*2
	} else {
		price = lastPrice + md.Data.TickSize*2
	}
	if side == "buy" && strings.Contains(md.Data.DisableTransaction, "buy") {
		time.Sleep(time.Second * 60)
		return "", price, quantity, errors.New("Buy Transaction Totally Disabled")
	} else if side == "sell" && strings.Contains(md.Data.DisableTransaction, "sell") {
		time.Sleep(time.Second * 60)
		return "", price, quantity, errors.New("Sell Transaction Totally Disabled")
	} else if strings.Contains(md.Data.DisableTransaction, "both") {
		time.Sleep(time.Second * 300)
		return "", price, quantity, errors.New("Both Sell and Buy Transactions Totally Disabled")
	}
	i = 0
	clientOrderID = <-w.UUIDChan
	after3by7Sec = time.After(time.Minute * 30)
	for {
		ordr, err = w.API.NewOrder(clientOrderID, md.Data.SymbolCode, side, fmt.Sprintf("%f", quantity), fmt.Sprintf("%f", price))
		if err != nil {
			ErrorMessage = fmt.Sprintf("%v", err)
			if strings.Contains(ErrorMessage, "500 Internal Server") || strings.Contains(ErrorMessage, "502 Bad Gateway") || strings.Contains(ErrorMessage, "503 Service Unavailable") || strings.Contains(ErrorMessage, "Exchange temporary closed") {
				md.Chans.MessageChan <- fmt.Sprintf("PlaceOrder: %s: %s: %s: Unable to place order due to  | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, side, md.Data.DisableTransaction)
				time.Sleep(time.Millisecond * 500)
				if i <= 5 {
					i++
					continue
				}
			}
			if strings.Contains(ErrorMessage, "Insufficient funds") {
				return "Insufficient funds", price, quantity, err
			}
			return "", price, quantity, err
		}
		if ordr.Status == "filled" {
			if ordr.Price != "" {
				price, _ = strconv.ParseFloat(ordr.Price, 64)
			}
			return "filled completely ", price, quantity, err
		} else if ordr.Status == "new" || ordr.Status == "partiallyFilled" {
			for {
				select {
				case <-md.Chans.CancelMyOrderChan:
					i = 0
					for {
						_, err = w.API.CancelOrder(clientOrderID)
						if err == nil {
							md.Chans.MessageChan <- fmt.Sprintf("PlaceOrder: %s: %s: %s Order is now \"canceled\" by owner, so was unsuccessful | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, side, md.Data.DisableTransaction)
							return "canceled", price, quantity, err
						}
						ErrorMessage = fmt.Sprintf("%v", err)
						if strings.Contains(ErrorMessage, "502 Bad Gateway") || strings.Contains(ErrorMessage, "503 Service Unavailable") || strings.Contains(ErrorMessage, "500 Internal Server") || strings.Contains(ErrorMessage, "Exchange temporary closed") {
							md.Chans.MessageChan <- fmt.Sprintf("PlaceOrder: %s: %s: Unable to cancel order due to %s\n", md.Data.SymbolCode, w.user.Username, ErrorMessage)
							time.Sleep(time.Millisecond * 500)
							if i <= 5 {
								i++
								continue
							}
						}
						return "", price, quantity, err
					}
				case <-after3by7Sec:
					for {
						//Getting Order ................
						ordr, err = w.API.GetOrder(clientOrderID, string(20000))
						if err != nil {
							ErrorMessage = fmt.Sprintf("%v", err)
							if strings.Contains(ErrorMessage, "502 Bad Gateway") || strings.Contains(ErrorMessage, "500 Internal Server") || strings.Contains(ErrorMessage, "503 Service Unavailable") {
								time.Sleep(time.Millisecond * 1000)
								continue
							} else if strings.Contains(ErrorMessage, "Order not found") {
								return "filled completely ", price, quantity, nil
							} else {
								md.Chans.MessageChan <- fmt.Sprintf("PlaceOrder: %s: %s: %s: Unable to get order status due to %v", md.Data.SymbolCode, w.user.Username, side, err)
								continue
							}
						}
						if ordr.Status == "filled" {
							if ordr.Price != "" {
								price, _ = strconv.ParseFloat(ordr.Price, 64)
							}
							return "filled completely ", price, quantity, err
						}
						if ordr.Status == "partiallyFilled" {
							time.Sleep(time.Millisecond * 500)
							continue
						}
						if ordr.Status == "new" && i < 10 {
							time.Sleep(time.Millisecond * 1000)
							i++
							continue
						}
						quantity, _ = strconv.ParseFloat(ordr.CumQuantity, 64)
						if quantity > 0.0 {
							md.Chans.MessageChan <- fmt.Sprintf("PlaceOrder: %s: %s: %s Order is filled partially | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, side, md.Data.DisableTransaction)
							return "filled partially s1", price, quantity, err
						}
						break
					}
					i = 0
					for {
						ordr, err = w.API.CancelOrder(clientOrderID)
						if err == nil {
							quantity, _ = strconv.ParseFloat(ordr.CumQuantity, 64)
							if quantity > 0.0 {
								md.Chans.MessageChan <- fmt.Sprintf("PlaceOrder: %s: %s: %s Order is filled partially | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, side, md.Data.DisableTransaction)
								return "filled partially s2", price, quantity, err
							}
							md.Chans.MessageChan <- fmt.Sprintf("PlaceOrder: %s: %s: %s Order is now \"canceled\" by overtime, so was unsuccessful | disabled: \"%s\"\n", md.Data.SymbolCode, w.user.Username, side, md.Data.DisableTransaction)
							return "canceled", price, quantity, err
						}
						ErrorMessage = fmt.Sprintf("%v", err)
						if strings.Contains(ErrorMessage, "502 Bad Gateway") || strings.Contains(ErrorMessage, "503 Service Unavailable") || strings.Contains(ErrorMessage, "500 Internal Server") || strings.Contains(ErrorMessage, "Exchange temporary closed") {
							md.Chans.MessageChan <- fmt.Sprintf("PlaceOrder: %s: %s: %s: Unable to cancel order", md.Data.SymbolCode, w.user.Username, side)
							time.Sleep(time.Millisecond * 500)
							if i <= 5 {
								i++
								continue
							}
						}
						time.Sleep(time.Millisecond * 500)
						return "", price, quantity, err
					}
				default:
					//Getting Order ................
					ordr, err = w.API.GetOrder(clientOrderID, string(20000))
					if err != nil {
						ErrorMessage = fmt.Sprintf("%v", err)
						if strings.Contains(ErrorMessage, "502 Bad Gateway") || strings.Contains(ErrorMessage, "500 Internal Server") || strings.Contains(ErrorMessage, "503 Service Unavailable") {
							md.Chans.MessageChan <- fmt.Sprintf("PlaceOrder: %s: %s: %s: Unable to get order status due to %v", md.Data.SymbolCode, w.user.Username, side, err)
							time.Sleep(time.Millisecond * 1000)
							continue
						} else if strings.Contains(ErrorMessage, "Order not found") {
							return "filled completely ", price, quantity, nil
						}
						md.Chans.MessageChan <- fmt.Sprintf("PlaceOrder: %s: %s: %s: Unable to get order status", md.Data.SymbolCode, w.user.Username, side)
						time.Sleep(time.Millisecond * 500)
						continue
					}
					if ordr.Status == "filled" {
						if ordr.Price != "" {
							price, _ = strconv.ParseFloat(ordr.Price, 64)
						}
						return "filled completely ", price, quantity, err
					}
					time.Sleep(time.Millisecond * 500)
					continue
				}
			}
		} else {
			md.Chans.MessageChan <- fmt.Sprintf("PlaceOrder: %s: %s: unhandled order status \"%s\" therefore will redo order\n", md.Data.SymbolCode, w.user.Username, ordr.Status)
		}
	}
}

//EmaData ...
type EmaData struct {
	Ema8  float64
	Ema4  float64
	Ema55 float64
	Ema15 float64
}

//GetEmaFunc ...
func (w WorkerAppService) GetEmaFunc(md *App) EmaData {
	var (
		Ema8, Ema4, Ema55, Ema15, closedPrice float64
		ema8, ema4, ema55, ema15              *ema.Ema
		candles                               *api.Candles
		candle                                api.Candle
		err                                   error
		ErrorMessage                          string
	)
	ema8 = ema.NewEma(alphaFromN(8))
	ema4 = ema.NewEma(alphaFromN(4))
	ema55 = ema.NewEma(alphaFromN(55))
	ema15 = ema.NewEma(alphaFromN(15))

	for {
		candles, err = w.API.GetCandles(md.Data.SymbolCode, w.Limit, w.Period2)
		if err != nil {
			ErrorMessage = fmt.Sprintf("%v", err)
			md.Chans.MessageChan <- fmt.Sprintf("GetEmaFunc: %s: Unable to get candles due to:\n", md.Data.SymbolCode)
			if strings.Contains(ErrorMessage, "502 Bad Gateway") || strings.Contains(ErrorMessage, "503 Service Unavailable") || strings.Contains(ErrorMessage, "500 Internal Server") || strings.Contains(ErrorMessage, "Exchange temporary closed") {
				time.Sleep(time.Millisecond * time.Duration(5000+<-w.Rand800))
			} else {
				time.Sleep(time.Second * 5)
			}
			continue
		}
		break
	}

	for _, candle = range *candles {
		closedPrice, _ = strconv.ParseFloat(candle.Close, 64)
		ema8.Step(closedPrice)
		ema4.Step(closedPrice)
		ema55.Step(closedPrice)
		ema15.Step(closedPrice)
	}
	Ema8 = ema8.Compute()
	Ema4 = ema4.Compute()
	Ema55 = ema55.Compute()
	Ema15 = ema15.Compute()
	return EmaData{Ema8, Ema4, Ema55, Ema15}
}
func randGen800() <-chan int {
	randNum := make(chan int, 100)
	go func() {
		for {
			randNum <- rand.Intn(800)
		}
	}()
	return randNum
}
func hatchID() string {
	var (
		u4      *hatchuid.UUID
		u4hatch []string
	)
	u4, _ = hatchuid.NewV4()
	u4hatch = strings.Split(u4.String(), "-")
	return u4hatch[0]
}
func genXid() string {
	var (
		id xid.ID
	)
	id = xid.New()
	return id.String()
}

//RealtimeUUID ...
func RealtimeUUID(DataChannel chan string) {
	for {
		var (
			uuidGenxID, uuidHatchID string
		)
		uuidGenxID = genXid()
		uuidHatchID = hatchID()
		DataChannel <- uuidGenxID + uuidHatchID
	}
}
func alphaFromN(N int) float64 {
	var (
		n float64
	)
	n = float64(N)
	return 2. / (n + 1.)
}
func trunc8(v float64) float64 {
	return math.Trunc(v*100000000) / 100000000
}
func trunc2(v float64) float64 {
	return math.Trunc(v*100) / 100
}
func trunc4(v float64) float64 {
	return math.Trunc(v*10000) / 10000
}
func trunc1(v float64) float64 {
	return math.Trunc(v*10) / 10
}
func isClosed(ch chan bool) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"myhitbtcv4/model"
	"myhitbtcv4/webClient"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/go-chi/chi"
	"golang.org/x/crypto/bcrypt"
)

var (
	indexTmpl          = webClient.NewAppTemplate("index.html")
	gettickerTmpl      = webClient.NewAppTemplate("getticker.html")
	getbalanceTmpl     = webClient.NewAppTemplate("getbalance.html")
	newtransactionTmpl = webClient.NewAppTemplate("newtransaction.html")
	signupTmpl         = webClient.NewAppTemplate("signup.html")
	loginTmpl          = webClient.NewAppTemplate("login.html")
	addappTmpl         = webClient.NewAppTemplate("addapp.html")
	deleteappTmpl      = webClient.NewAppTemplate("deleteapp.html")
	deleteallappTmpl   = webClient.NewAppTemplate("deleteallapp.html")
	editappTmpl        = webClient.NewAppTemplate("editapp.html")
	getapplistTmpl     = webClient.NewAppTemplate("getapplist.html")
	marginTmpl         = webClient.NewAppTemplate("margin.html")
)
var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize,
// 	WriteBufferSize: socketBufferSize}

//TradeHandler contains the chi mux and session and implements the ServeMux method
type TradeHandler struct {
	mux              *chi.Mux
	host             string
	sessionDBService SessionDBService
	uuidChan         chan string
	boltDbChans      model.ABDBChans
	chatHandler		 http.Handler
}

//NewTradeHandler returns a new instance of *TradeHandler
func NewTradeHandler(chatHdler http.Handler, host string, sessDBS SessionDBService, abdbchans model.ABDBChans, uuidch chan string) TradeHandler {
	h := TradeHandler{
		mux:         chi.NewRouter(),
		host:        host,
		uuidChan:    uuidch,
		boltDbChans: abdbchans,
		chatHandler: chatHdler,
	}
	
	h.sessionDBService = sessDBS
	h.sessionDBService.session.cachedUser = &model.User{}
	h.mux.Get("/signup", h.userSignUpHandler)
	h.mux.Post("/signup", h.userSignUpHandler)
	h.mux.Get("/login", h.userLoginHandler)
	h.mux.Post("/login", h.userLoginHandler)
	h.mux.Get("/addapp", h.userAddAppHandler)
	h.mux.Post("/addapp", h.userAddAppHandler)
	h.mux.Get("/editapp", h.userEditAppHandler)
	h.mux.Post("/editapp", h.userEditAppHandler)
	h.mux.Get("/getapplist", h.userGetAppListHandler)
	h.mux.Get("/feeds/ws", h.userFeedsHandler)
	h.mux.Get("/close", h.userCloseUserSocketHandler)
	// h.mux.Get("/shutdownapp", h.userShutdownAppHandler)
	// h.mux.Post("/shutdownapp", h.userShutdownAppHandler)
	// h.mux.Get("/shutdownallapp", h.userShutdownAllAppHandler)
	h.mux.Get("/deleteapp", h.userDeleteAppHandler)
	h.mux.Post("/deleteapp", h.userDeleteAppHandler)
	h.mux.Get("/deleteallapp", h.userDeleteAllAppHandler)
	h.mux.Get("/newverwrite", h.newVerWriteHandler)
	h.mux.Post("/mdupdate", h.mdUploadHandler)
	h.mux.Get("/messagedownload", h.userMessageDownloadHandler)
	h.mux.Get("/graphpoint", h.userGraphPointAppHandler)
	h.mux.Get("/newverread", h.newVerReadHandler)
	h.mux.Post("/deleteallapp", h.userDeleteAllAppHandler)
	h.mux.Get("/resetapp", h.userResetAppHandler)
	h.mux.Get("/message", h.userMessageAppHandler)
	h.mux.Get("/margin", h.userMarginAppHandler)
	h.mux.Post("/logout", h.userlogoutHandler)
	h.mux.Get("/", h.indexHandler)
	return h
}

//TradeHandler implements ServeHTTP method making it a Handler
func (h TradeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/webClient/asset/") {
		http.StripPrefix("/webClient/asset/", http.FileServer(http.Dir("./webClient/asset/"))).ServeHTTP(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/chat") {
		h.chatHandler.ServeHTTP(w, r)
	} else {
		h.mux.ServeHTTP(w, r)
	}
}
func (h TradeHandler) userMessageDownloadHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("Username")
	token := r.FormValue("Token")
	// is there a username?
	CallerChan := make(chan model.UserDbResp)
	h.sessionDBService.session.userBoltDBChans.GetDbByNameChan <- model.UserDbByNameData{username, nil, CallerChan}
	dbResp := <-CallerChan
	if dbResp.User == nil || dbResp.Err != nil || dbResp.User.Username != username {
		http.Error(w, "Username not Found", http.StatusForbidden)
		return
	}
	if dbResp.User.Token != token {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	readFile, err := os.OpenFile("./nohup.out", os.O_RDONLY, 0600)
	if err != nil {
		fmt.Printf("failed to open file: %v", err)
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var FileTextLines []string
	for fileScanner.Scan() {
		FileTextLines = append(FileTextLines, fileScanner.Text())
	}
	readFile.Close()
	err = json.NewEncoder(w).Encode(FileTextLines)
	if err != nil {
		log.Println(err)
	}
	return
}

//indexHandler delivers the Home page to the user
func (h TradeHandler) newVerWriteHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("Username")
	token := r.FormValue("Token")
	// is there a username?
	CallerChan := make(chan model.UserDbResp)
	h.sessionDBService.session.userBoltDBChans.GetDbByNameChan <- model.UserDbByNameData{username, nil, CallerChan}
	dbResp := <-CallerChan
	if dbResp.User == nil || dbResp.Err != nil || dbResp.User.Username != username {
		http.Error(w, "Username not Found", http.StatusForbidden)
		return
	}
	if dbResp.User.Token != token {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	AppCallerRespChan := make(chan AppDbResp)
	i := 0
	webJSONData := make([]*model.AppData, len(dbResp.User.ApIDs))
	for _, v := range dbResp.User.ApIDs {
		h.sessionDBService.session.appMemDBChans.GetDbChan <- AppDbServiceVehicle{v, nil, AppCallerRespChan}
		Res := <-AppCallerRespChan
		AppVeh := <-Res.App.Chans.MyChan
		AppVeh.RespChan <- true
		webJSONData[i] = AppVeh.App.Data
		i++
	}
	err := json.NewEncoder(w).Encode(webJSONData)
	if err != nil {
		log.Println(err)
	}
	return
}
func convertOld2New(WebJSON *model.AppDataOld) *model.AppData {
	Data := &model.AppData{
		ID: WebJSON.ID,
		//UsrID:                  WebJSON.UsrID,
		SessID:                 WebJSON.SessID,
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
		PendingA:               WebJSON.Pending,
		PendingB:               WebJSON.Pending,
		HeartbeatBuy:           WebJSON.HeartbeatBuy,
		MessageFilter:          WebJSON.MessageFilter,
		HeartbeatSell:          WebJSON.HeartbeatSell,
		NextMarketBuyPoint:     WebJSON.NextMarketBuyPoint,
		NextMarketSellPoint:    WebJSON.NextMarketSellPoint,
		DisableTransaction:     WebJSON.DisableTransaction,
		ProfitPriceUsed:        WebJSON.ProfitPriceUsed,
		ProfitPrice:            WebJSON.ProfitPrice,
	}
	return Data
}

//indexHandler delivers the Home page to the user
func (h TradeHandler) newVerReadHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := AlreadyLoggedIn(w, r, &h)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	//create url
	url := "http://localhost:" + r.FormValue("port") + "/newverwrite?Username=" + user.Username + ";Token=" + user.Token
	// Create Client
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	if Resp, err := client.Do(req); err == nil {
		if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
			var WebJSONData []*model.AppData
			Jdata := json.NewDecoder(Resp.Body)
			err := Jdata.Decode(&WebJSONData)
			if err != nil {
				msg := fmt.Sprintf("Get %s Error: %v", url, err)
				http.Error(w, msg, http.StatusInternalServerError)
				return
			}
			for i := 0; i < len(WebJSONData); i++ {
				//Check if symbol is already trading
				_, err := h.sessionDBService.session.appMemDBService.GetApp(user.ApIDs[WebJSONData[i].SymbolCode])
				if err != model.ErrAppNameEmpty && err != model.ErrAppNotFound {
					http.Error(w, "Symbol Already Trading ", http.StatusInternalServerError)
					return
				}
				//Iniialling worker service for this session
				md := &App{}
				//md.Data = convertOld2New(WebJSONData[i])
				md.Data = WebJSONData[i]
				log.Printf("For %s: For %s: pending = %v", md.Data.SymbolCode, user.Username, WebJSONData[i].PendingB)
				h.sessionDBService.session.workerAppService = NewWorkerAppService(md.Data, &h.sessionDBService.session, h.uuidChan)
				_, err = h.sessionDBService.session.workerAppService.API.GetSymbol(md.Data.SymbolCode)
				if err != nil {
					log.Printf("For %s: %v\n", md.Data.SymbolCode, err)
					continue
				}
				h, err = h.syncParams(user, md, "add")
				if err != nil {
					log.Printf("newVerReadHandler1 %v\n", err)
					return
				}
				appMarginSendingChan := make(chan model.MarginParam)
				md.FromVersionUpdate = true
				md.Chans.MyChan, err = h.sessionDBService.session.workerAppService.AutoTradeManager(md, appMarginSendingChan)
				if err != nil {
					log.Printf("newVerReadHandler2 %v\n", err)
					return
				}
				md.Chans.MParamChan = appMarginSendingChan
				h, err = h.syncParams(user, md, "addBolt")
				if err != nil {
					log.Printf("newVerReadHandler3 %v\n", err)
					return
				}
			}
			log.Printf("%d completed !!!!!!!", len(WebJSONData))
			http.Redirect(w, r, "/getapplist", http.StatusSeeOther)
			return
		}
	} else {
		msg := fmt.Sprintf("Could not connect: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	return
}

//indexHandler delivers the Home page to the user
func (h TradeHandler) mdUploadHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := AlreadyLoggedIn(w, r, &h)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var WebJSONData []*model.AppData
	Jdata := json.NewDecoder(r.Body)
	err := Jdata.Decode(&WebJSONData)
	if err != nil {
		msg := fmt.Sprintf("Error: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	var dat AppVehicle
	for i := 0; i < len(WebJSONData); i++ {
		//Check if symbol is already trading
		dataOfApp, err := h.sessionDBService.session.appMemDBService.GetApp(user.ApIDs[WebJSONData[i].SymbolCode])
		if err != model.ErrAppNameEmpty && err != model.ErrAppNotFound {
			//save the connection
			log.Printf("%s Already Exist", WebJSONData[i].SymbolCode)
			dat = <-dataOfApp.Chans.MyChan
			dat.App.Data.DisableTransaction = WebJSONData[i].DisableTransaction
			dat.App.Data.MessageFilter = WebJSONData[i].MessageFilter
			dat.App.Data.NeverBought = WebJSONData[i].NeverBought
			dat.App.Data.NeverSold = WebJSONData[i].NeverSold
			dat.App.Data.QuantityIncrement = WebJSONData[i].QuantityIncrement
			dat.App.Data.TickSize = WebJSONData[i].TickSize
			dat.App.Data.TakeLiquidityRate = WebJSONData[i].TakeLiquidityRate
			dat.App.Data.SuccessfulOrders = WebJSONData[i].SuccessfulOrders
			dat.App.Data.MadeProfitOrders = WebJSONData[i].MadeProfitOrders
			dat.App.Data.MadeLostOrders = WebJSONData[i].MadeLostOrders
			dat.App.Data.StopLostPoint = WebJSONData[i].StopLostPoint
			dat.App.Data.TrailPoints = WebJSONData[i].TrailPoints
			dat.App.Data.LeastProfitMargin = WebJSONData[i].LeastProfitMargin
			dat.App.Data.SureTradeFactor = WebJSONData[i].SureTradeFactor
			dat.App.Data.Hodler = WebJSONData[i].Hodler
			dat.App.Data.GoodBiz = WebJSONData[i].GoodBiz
			dat.App.Data.InstantProfit = WebJSONData[i].InstantProfit
			dat.App.Data.InstantLost = WebJSONData[i].InstantLost
			dat.App.Data.TotalProfit = WebJSONData[i].TotalProfit
			dat.App.Data.TotalLost = WebJSONData[i].TotalLost
			dat.App.Data.PendingA = WebJSONData[i].PendingA
			dat.App.Data.PendingB = WebJSONData[i].PendingB
			dat.RespChan <- true
			h, err = h.syncParams(user, dat.App, "updateBolt")
			if err != nil {
				log.Printf("%s AppUpload Error %v\n", dat.App.Data.SymbolCode, err)
				return
			}
			continue
		}
		//Initialling worker service for this session
		log.Printf("%s Is New", WebJSONData[i].SymbolCode)
		md := &App{}
		//md.Data = convertOld2New(WebJSONData[i])
		md.Data = WebJSONData[i]
		log.Printf("For %s: For %s: pending = %v", md.Data.SymbolCode, user.Username, WebJSONData[i].PendingB)
		h.sessionDBService.session.workerAppService = NewWorkerAppService(md.Data, &h.sessionDBService.session, h.uuidChan)
		_, err = h.sessionDBService.session.workerAppService.API.GetSymbol(md.Data.SymbolCode)
		if err != nil {
			log.Printf("For %s: %v\n", md.Data.SymbolCode, err)
			continue
		}
		h, err = h.syncParams(user, md, "add")
		if err != nil {
			log.Printf("newVerReadHandler1 %v\n", err)
			return
		}
		appMarginSendingChan := make(chan model.MarginParam)
		md.FromVersionUpdate = true
		md.Chans.MyChan, err = h.sessionDBService.session.workerAppService.AutoTradeManager(md, appMarginSendingChan)
		if err != nil {
			log.Printf("newVerReadHandler2 %v\n", err)
			return
		}
		md.Chans.MParamChan = appMarginSendingChan
		h, err = h.syncParams(user, md, "addBolt")
		if err != nil {
			log.Printf("newVerReadHandler3 %v\n", err)
			return
		}
	}
	log.Printf("%d completed !!!!!!!", len(WebJSONData))
	http.Redirect(w, r, "/getapplist", http.StatusSeeOther)
	return
}

//indexHandler delivers the Home page to the user
func (h TradeHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	usr, _ := AlreadyLoggedIn(w, r, &h)
	if err := indexTmpl.Execute(w, r, nil, usr, nil); err != nil {
		log.Printf("indexHandler1 %v\n", err)
		return
	} //Prints ticker to webpage
}
func (h TradeHandler) UserPowerUpHandler(uDBRCC chan chan *model.User, GetDbChan chan model.AppDataBoltVehicle) {
	log.Printf("UserPowerUpHandler started")
	var (
		appID model.AppID
		err   error
	)
	userDBRetrivalChan := make(chan *model.User)
	uDBRCC <- userDBRetrivalChan
	var memDBHolder []*model.User
	//Range the User DB to get each user
	for muser := range userDBRetrivalChan {
		memDBHolder = append(memDBHolder, muser)
	}
	for _, user := range memDBHolder {
		//Range the APIDs of the user to get each Symbol id
		h.sessionDBService.session.ID = user.SessID
		h.sessionDBService.session.Sockets = make(map[string]AppConn)
		var k string
		for k, appID = range user.ApIDs {
			//Get each App and set it running in memmory
			//Check if symbol is already trading
			md := &App{}
			CallerChan := make(chan model.AppDataResp)
			GetDbChan <- model.AppDataBoltVehicle{appID, nil, CallerChan}
			resp := <-CallerChan
			if resp.Err != nil {
				log.Printf("For %s: %v: %v: Error %v\n", user.Username, appID, md, resp.Err)
				delete(user.ApIDs, k)
				continue
			}
			md.Data = resp.AppData
			h.sessionDBService.session.workerAppService = NewWorkerAppService(md.Data, &h.sessionDBService.session, h.uuidChan)
			_, err = h.sessionDBService.session.workerAppService.API.GetSymbol(md.Data.SymbolCode)
			if err != nil {
				log.Printf("For %s: %s: Error %v\n", user.Username, md.Data.SymbolCode, err)
				continue
			}
			h, err = h.syncParams(user, md, "add")
			if err != nil {
				log.Printf("For %s: %s: Error %v\n", user.Username, md.Data.SymbolCode, err)
				continue
			}
			appMarginSendingChan := make(chan model.MarginParam)
			md.Chans.MyChan, err = h.sessionDBService.session.workerAppService.AutoTradeManager(md, appMarginSendingChan)
			if err != nil {
				log.Printf("For %s: %s: Error %v\n", user.Username, md.Data.SymbolCode, err)
				continue
			}
			md.Chans.MParamChan = appMarginSendingChan
			h, err = h.syncParams(user, md, "updateBolt")
			if err != nil {
				log.Printf("For %s: %s: Error %v\n", user.Username, md.Data.SymbolCode, err)
				continue
			}
			log.Printf("Revived %v for User %v", md.Data.SymbolCode, user.Username)
		}
		h, err = h.syncParams(user, nil, "add")
		if err != nil {
			log.Printf("For %s: Error %v\n", user.Username, err)
			continue
		}
	}
	return
}
func (h TradeHandler) userSignUpHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := AlreadyLoggedIn(w, r, &h); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var user model.User
	// process form submission
	if r.Method == http.MethodPost {
		// username taken?
		username := r.FormValue("username")
		CallerChan := make(chan model.UserDbResp)
		h.sessionDBService.session.userBoltDBChans.GetDbByNameChan <- model.UserDbByNameData{username, nil, CallerChan}
		dbResp := <-CallerChan
		if dbResp.User != nil && dbResp.Err == nil && dbResp.User.Username == username {
			http.Error(w, "Username already taken", http.StatusForbidden)
			return
		}
		// get form values
		user = model.User{
			Username: username,
			Email:    r.FormValue("email"),
		}
		// encrypt the password
		bc, err := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), bcrypt.MinCost)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		user.Password = string(bc)
		user.ApIDs = make(map[string]model.AppID)
		h.sessionDBService.session.Sockets = make(map[string]AppConn)
		//Promote session
		h.sessionDBService.session.ID = model.SessionID(<-h.uuidChan)
		//sync user and session and add them to their DBs
		h.sessionDBService.session.userBoltDBChans.AddDbChan <- model.UserDbData{0, &user, CallerChan}
		dbResp = <-CallerChan
		user.ID = dbResp.UserID
		h, err = h.syncParams(&user, nil, "add")
		if err != nil {
			log.Printf("userSignUpHandler1 %v\n", err)
			return
		}
		// create session
		h.sessionDBService.session.SetToken(w, r, user.Username, user.ID, "/", user.Level)
		return
	}
	if err := signupTmpl.Execute(w, r, nil, nil, nil); err != nil {
		log.Printf("userSignUpHandler2 %v\n", err)
		return
	}
	return
}
func (h TradeHandler) userLoginHandler(w http.ResponseWriter, r *http.Request) {
	var dbResp model.UserDbResp
	if _, ok := AlreadyLoggedIn(w, r, &h); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	// process form submission
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		// is there a username?
		CallerChan := make(chan model.UserDbResp)
		h.sessionDBService.session.userBoltDBChans.GetDbByNameChan <- model.UserDbByNameData{username, nil, CallerChan}
		dbResp = <-CallerChan
		if dbResp.User == nil || dbResp.Err != nil || dbResp.User.Username != username {
			http.Error(w, "Username not Found", http.StatusForbidden)
			return
		}
		// does the entered password match the stored password?
		err := bcrypt.CompareHashAndPassword([]byte(dbResp.User.Password), []byte(password))
		if err != nil {
			http.Error(w, "Username and/or password do not match", http.StatusForbidden)
			return
		}
		// get and create session
		//get session
		sCallerChan := make(chan SessionDbResp)
		h.sessionDBService.sessionDBChans.GetDbChan <- SessionDbData{dbResp.User.SessID, nil, sCallerChan}
		sDbResp := <-sCallerChan
		if sDbResp.Session == nil || sDbResp.Err != nil {
			http.Error(w, "Unable to creat Session please signUp", http.StatusForbidden)
			log.Printf("%v\n", sDbResp.Err)
			return
		}
		h.sessionDBService.session = *sDbResp.Session
		// create web session
		h.sessionDBService.session.SetToken(w, r, dbResp.User.Username, dbResp.User.ID, "/", dbResp.User.Level)
		return
	}
	if err := loginTmpl.Execute(w, r, nil, dbResp.User, nil); err != nil {
		log.Printf("userLoginHandler1 %v\n", err)
		return
	}
	return
}
func (h TradeHandler) userAddAppHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := AlreadyLoggedIn(w, r, &h)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	// process form submission
	if r.Method == http.MethodPost {
		// get form values and initialize App
		mdHighLevel := &App{
			Data: &model.AppData{
				PublicKey:  r.FormValue("publickey"),
				Secret:     r.FormValue("secret"),
				Host:       r.FormValue("platform"),
				SymbolCode: r.FormValue("symbol"),
			},
		}
		if mdHighLevel.Data.SymbolCode == "" {
			http.Error(w, "Symbol is Empty please enter a Trading Symbol ", http.StatusInternalServerError)
			return
		}
		if mdHighLevel.Data.Host == "" {
			http.Error(w, "Please ensure Host is provided and try again", http.StatusInternalServerError)
			return
		}
		if mdHighLevel.Data.PublicKey == "" {
			http.Error(w, "Please ensure PublicKey is provided and try again", http.StatusInternalServerError)
			return
		}
		if mdHighLevel.Data.Secret == "" {
			http.Error(w, "Please ensure Secret is provided and try again", http.StatusInternalServerError)
			return
		}
		symbols := strings.Split(mdHighLevel.Data.SymbolCode, ",")
		var symbol string
		// This Channel is use to store data direct to bolt database
		hboltDbChansAddDbChan := h.boltDbChans.AddDbChan
		//Further Initialize other fields of App
		for _, symbol = range symbols {
			hboltDbChansAddDbChan = h.boltDbChans.AddDbChan
			//Check if symbol is already trading
			_, err := h.sessionDBService.session.appMemDBService.GetApp(user.ApIDs[symbol])
			// if already exist ask the user whether it should be delete and return
			if err != model.ErrAppNameEmpty && err != model.ErrAppNotFound {
				dat := GetAppDataStringified{
					DeleteMessage: "Symbol Already Trading, Do you want to delete: ",
					SymbolCode:    symbol,
				}
				//presenting a delete form to the user to confirm deletion
				if err := deleteappTmpl.Execute(w, r, dat, user, nil); err != nil {
					log.Printf("userAddAppHandler1 %v\n", err)
				}
				return
			}
			// Creating initial md
			md := &App{
				Data: &model.AppData{
					PublicKey: mdHighLevel.Data.PublicKey,
					Secret:    mdHighLevel.Data.Secret,
					Host:      mdHighLevel.Data.Host,
				},
			}
			md.Data.SymbolCode = symbol
			md.Data.UsrID = user.ID
			RespChan := make(chan model.AppDataResp)
			//sending md to database and to generate the ID
			hboltDbChansAddDbChan <- model.AppDataBoltVehicle{0, md.Data, RespChan}
			hboltDbChansAddDbChan = nil
			apdaresp := <-RespChan
			if apdaresp.Err != nil {
				if apdaresp.Err == model.ErrAppExists {
					// If already existing then get it and have the ID
					h.boltDbChans.GetDbChan <- model.AppDataBoltVehicle{apdaresp.AppID, md.Data, RespChan}
					apdaresp = <-RespChan
					if apdaresp.Err != nil {
						log.Printf("%v\n", apdaresp.Err)
						http.Redirect(w, r, "/getapplist", http.StatusSeeOther)
						return
					}
					md.Data = apdaresp.AppData
					// to close and reopen any existing priceTrading with the new or revived version
					md.FromVersionUpdate = true
				} else {
					log.Printf("%v\n", apdaresp.Err)
					http.Redirect(w, r, "/getapplist", http.StatusSeeOther)
					return
				}
			}
			md.Data.ID = apdaresp.AppID
			h.sessionDBService.session.workerAppService = NewWorkerAppService(md.Data, &h.sessionDBService.session, h.uuidChan)
			_, err = h.sessionDBService.session.workerAppService.API.GetSymbol(md.Data.SymbolCode)
			if err != nil {
				log.Printf("userAddAppHandler2 %v\n", err)
				http.Error(w, md.Data.SymbolCode+" Symbol not found at Provider End", http.StatusInternalServerError)
				return
			}
			h, err = h.syncParams(user, md, "add")
			if err != nil {
				log.Printf("userAddAppHandler3 %v\n", err)
				return
			}
			appMarginSendingChan := make(chan model.MarginParam)
			md.Chans.MyChan, err = h.sessionDBService.session.workerAppService.AutoTradeManager(md, appMarginSendingChan)
			if err != nil {
				log.Printf("userAddAppHandler4 %v\n", err)
				return
			}
			md.Chans.MParamChan = appMarginSendingChan
			h, err = h.syncParams(user, md, "add")
			if err != nil {
				log.Printf("userAddAppHandler5 %v\n", err)
				return
			}
		}
		http.Redirect(w, r, "/getapplist", http.StatusSeeOther)
		return
	}
	if err := addappTmpl.Execute(w, r, nil, user, nil); err != nil {
		log.Printf("userAddAppHandler6 %v\n", err)
		return
	}
	return
}
func (h TradeHandler) userMarginAppHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := AlreadyLoggedIn(w, r, &h)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := MarginData{}
	mar := model.MarginDBVeh{
		MPChan:         nil,
		User:           user,
		MCalDBRespChan: make(chan model.MarginDBResp),
	}
	GrandMargin := 0.0
	Margins, err := h.sessionDBService.session.marginCalDBService.GetMargin(mar)
	if err != nil {
		log.Printf("userMarginAppHandler1 %v\n", err)
	}else{
		for k, v := range Margins {
			if k == user.ApIDs[v.SymbolCode] {
				data.Margins = append(data.Margins, Margin{
					SymbolCode:       v.SymbolCode,
					SuccessfulOrders: fmt.Sprintf("%.8f", v.SuccessfulOrders),
					MadeProfitOrders: fmt.Sprintf("%.8f", v.MadeProfitOrders),
					MadeLostOrders:   fmt.Sprintf("%.8f", v.MadeLostOrders),
					Value:            fmt.Sprintf("%.8f", v.Value),
				})
				GrandMargin += v.Value
			}
		}
	}
	data.GrandMargin = fmt.Sprintf("%-8f", GrandMargin)
	if err := marginTmpl.Execute(w, r, data, user, nil); err != nil {
		log.Printf("userMarginAppHandler1 %v\n", err)
		return
	}
	return
}

type MarginData struct {
	GrandMargin string
	Margins     []Margin
}
type Margin struct {
	SymbolCode       string
	SuccessfulOrders string
	MadeProfitOrders string
	MadeLostOrders   string
	Value            string
}

func (h TradeHandler) userResetAppHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := AlreadyLoggedIn(w, r, &h)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	sync := make(chan bool, 1)
	id := r.FormValue("appid")
	app, err := h.sessionDBService.session.appMemDBService.GetApp(user.ApIDs[id])
	if err != nil {
		return
	}
	h, err = h.syncParams(user, app, "updateBolt")
	if err != nil {
		log.Printf("userResetAppHandler1 %v\n", err)
		return
	}
	err = h.sessionDBService.session.workerAppService.ResetApp(app, "all", sync)
	<-sync
	if err != nil {
		log.Printf("%v", err)
		return
	}
	http.Redirect(w, r, "/getapplist", http.StatusSeeOther)
	return
}
func (h TradeHandler) userFeedsHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := AlreadyLoggedIn(w, r, &h)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	//Promote to user websocket if meant for websocket activity
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("userFeedsHandler1 %v\n", err)
		return
	}
	data := GetAppDataStringified{}
	err = conn.ReadJSON(&data)
	if err != nil {
		errStr := fmt.Sprintf("%v", err)
		log.Printf("Error: %v", err)
		http.Error(w, "Internal server error: "+errStr, http.StatusInternalServerError)
		conn.Close()
		return
	}
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				conn.Close()
				return
			}
		}
	}()
	dataOfApp, err := h.sessionDBService.session.appMemDBService.GetApp(user.ApIDs[data.SymbolCode])
	if err != nil {
		errStr := fmt.Sprintf("%v", err)
		http.Error(w, "Internal server error: "+errStr, http.StatusInternalServerError)
		return
	}
	//save the connection
	appSocketChan := make(chan bool, 10)
	h.sessionDBService.session.Sockets[data.SymbolCode] = AppConn{conn, appSocketChan}
	h.syncParams(user, nil, "update")
	var dat AppVehicle
	go func(user *model.User) {
		//defer log.Printf("For %s: Socket Ended\n", data.SymbolCode)
		for {
			select {
			case <-time.After(time.Second * 65):
				return
			case <-appSocketChan:
				return
			case dat = <-dataOfApp.Chans.MyChan:
				dat.RespChan <- true
				data.MessageFilter = dat.App.Data.MessageFilter
				data.PendingA = fmt.Sprintf("%v", dat.App.Data.PendingA)
				data.PendingB = fmt.Sprintf("%v", dat.App.Data.PendingB)
				data.Message = dat.App.Data.Message
				data.DisableTransaction = dat.App.Data.DisableTransaction
				data.SureTradeFactor = fmt.Sprintf("%.8f", dat.App.Data.SureTradeFactor)
				data.SuccessfulOrders = fmt.Sprintf("%.8f", dat.App.Data.SuccessfulOrders)
				data.MadeProfitOrders = fmt.Sprintf("%.8f", dat.App.Data.MadeProfitOrders)
				data.GoodBiz = fmt.Sprintf("%.8f", dat.App.Data.GoodBiz)
				data.LeastProfitMargin = fmt.Sprintf("%.8f", dat.App.Data.LeastProfitMargin)
				data.QuantityIncrement = fmt.Sprintf("%.8f", dat.App.Data.QuantityIncrement)
				data.NeverBought = dat.App.Data.NeverBought
				data.NeverSold = dat.App.Data.NeverSold
				data.StopLostPoint = fmt.Sprintf("%.8f", dat.App.Data.StopLostPoint)
				data.TrailPoints = fmt.Sprintf("%.8f", dat.App.Data.TrailPoints)
				data.InstantProfit = fmt.Sprintf("%.8f", dat.App.Data.InstantProfit)
				data.TotalProfit = fmt.Sprintf("%.8f", dat.App.Data.TotalProfit)
				data.SymbolCode = dat.App.Data.SymbolCode
				data.TotalLost = fmt.Sprintf("%.8f", dat.App.Data.TotalLost)
				data.InstantLost = fmt.Sprintf("%.8f", dat.App.Data.InstantLost)
				data.MadeLostOrders = fmt.Sprintf("%.8f", dat.App.Data.MadeLostOrders)
				data.Hodler = dat.App.Data.Hodler

				err = h.sessionDBService.session.websocketUserService.WriteToSocket(user, &data)
				if err != nil {
					//log.Printf("For %s: %v", dat.App.Data.SymbolCode, err)
					appSocketChan <- true
				}
			}
			time.Sleep(time.Second * 60)
		}
	}(user)
}
func (h TradeHandler) userCloseUserSocketHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := AlreadyLoggedIn(w, r, &h)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	id := r.FormValue("appid")
	//To do: get user fro user DB service since this will be authenticated agaisnt
	//above user in websocket service close Socket
	err := h.sessionDBService.session.websocketUserService.CloseSocket(user, id)
	if err != nil {
		log.Printf("userCloseUserSocketHandler1 %v\n", err)
	}
	http.Redirect(w, r, "/getapplist", http.StatusSeeOther)
	return
}

func (h TradeHandler) userEditAppHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := AlreadyLoggedIn(w, r, &h)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	idcode := r.FormValue("symbolcode")
	if idcode == "" {
		http.Error(w, "Symbol is Empty Unable to edit", http.StatusInternalServerError)
		return
	}
	dataOfApp, err := h.sessionDBService.session.appMemDBService.GetApp(user.ApIDs[idcode])
	if err != nil {
		errStr := fmt.Sprintf("%v", err)
		http.Error(w, "Internal server error: "+errStr, http.StatusInternalServerError)
		return
	}
	dat := <-dataOfApp.Chans.MyChan
	if r.Method != http.MethodPost {
		dat.RespChan <- true
	}
	data := GetAppDataStringified{}
	data.MessageFilter = dat.App.Data.MessageFilter
	data.Message = dat.App.Data.Message
	data.DisableTransaction = dat.App.Data.DisableTransaction
	data.SureTradeFactor = fmt.Sprintf("%.8f", dat.App.Data.SureTradeFactor)
	data.SuccessfulOrders = fmt.Sprintf("%.8f", dat.App.Data.SuccessfulOrders)
	data.MadeProfitOrders = fmt.Sprintf("%.8f", dat.App.Data.MadeProfitOrders)
	data.GoodBiz = fmt.Sprintf("%.8f", dat.App.Data.GoodBiz)
	data.LeastProfitMargin = fmt.Sprintf("%.8f", dat.App.Data.LeastProfitMargin)
	data.QuantityIncrement = fmt.Sprintf("%.8f", dat.App.Data.QuantityIncrement)
	data.NeverBought = dat.App.Data.NeverBought
	data.NeverSold = dat.App.Data.NeverSold
	data.StopLostPoint = fmt.Sprintf("%.8f", dat.App.Data.StopLostPoint)
	data.TrailPoints = fmt.Sprintf("%.8f", dat.App.Data.TrailPoints)
	data.InstantProfit = fmt.Sprintf("%.8f", dat.App.Data.InstantProfit)
	data.TotalProfit = fmt.Sprintf("%.8f", dat.App.Data.TotalProfit)
	data.Hodler = dat.App.Data.Hodler
	data.SymbolCode = dat.App.Data.SymbolCode
	data.Secret = dat.App.Data.Secret
	data.PublicKey = dat.App.Data.PublicKey
	// process form submission
	if r.Method == http.MethodPost {
		// get form values and initialize App,
		dat.App.Data.MessageFilter = r.FormValue("messagefilter")
		dat.App.Data.GoodBiz, _ = strconv.ParseFloat(r.FormValue("goodbiz"), 64)
		dat.App.Data.LeastProfitMargin, _ = strconv.ParseFloat(r.FormValue("leastprofitmargin"), 64)
		dat.App.Data.QuantityIncrement, _ = strconv.ParseFloat(r.FormValue("quantityincrement"), 64)
		dat.App.Data.StopLostPoint, _ = strconv.ParseFloat(r.FormValue("StopLostPoint"), 64)
		dat.App.Data.TrailPoints, _ = strconv.ParseFloat(r.FormValue("trailpoints"), 64)
		dat.App.Data.InstantProfit, _ = strconv.ParseFloat(r.FormValue("instantprofit"), 64)
		dat.App.Data.SureTradeFactor, _ = strconv.ParseFloat(r.FormValue("suretradefactor"), 64)
		dat.App.Data.Hodler = r.FormValue("hodler")
		dat.App.Data.DisableTransaction = r.FormValue("disabletransaction")
		dat.App.Data.Secret = r.FormValue("secret")
		dat.App.Data.PublicKey = r.FormValue("publickey")
		dat.App.Data.NeverBought = r.FormValue("neverbought")
		dat.App.Data.NeverSold = r.FormValue("neversold")
		dat.RespChan <- true
		h, _ = h.syncParams(user, nil, "update")
		h.sessionDBService.session.workerAppService.API.SesSion.Auth = []string{dat.App.Data.PublicKey, dat.App.Data.Secret}
		h, _ = h.syncParams(user, dat.App, "update")
		http.Redirect(w, r, "/getapplist", http.StatusSeeOther)
		return
	}

	if err := editappTmpl.Execute(w, r, data, user, nil); err != nil {
		log.Printf("userEditAppHandler1 %v\n", err)
	}
	return
}
func (h TradeHandler) userDeleteAllAppHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		app *App
	)
	user, ok := AlreadyLoggedIn(w, r, &h)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	id := r.FormValue("appid")
	Replaced := false
	if strings.Contains(id, "POST") {
		Replaced = true
	}
	h, err = h.syncParams(user, nil, "update")
	if err != nil {
		log.Printf("userDeleteAllAppHandler1 %v\n", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
	if Replaced {
		wg := sync.WaitGroup{}
		wg.Add(len(user.ApIDs))
		for k, v := range user.ApIDs {
			h, err = h.syncParams(user, nil, "update")
			if err != nil {
				log.Printf("userDeleteAllAppHandler5 %v\n", err)
				http.Error(w, "Server Error", http.StatusInternalServerError)
				return
			}
			app, err = h.sessionDBService.session.appMemDBService.GetApp(v)
			if err == model.ErrAppNotFound {
				_, err := h.sessionDBService.session.appBoltDBService.GetApp(v)
				if err != nil {
					//removing app id from user
					log.Printf("userDeleteAllAppHandler2.0 %v\n", err)
					continue
				}
				//Deleting app from Bolt DB
				err = h.sessionDBService.session.appBoltDBService.DeleteApp(v)
				if err != nil {
					log.Printf("userDeleteAppHandler2.1 %v\n", err)
					continue
				}
				//removing app id from user
				delete(user.ApIDs, k)
				continue
			} else if err != nil {
				log.Printf("userDeleteAllAppHandler2.2 %v\n", err)
				continue
			}
			//Closing down websocket
			go func() {
				_ = h.sessionDBService.session.websocketUserService.CloseSocket(user, app.Data.SymbolCode)
			}()
			//Deleting app chan from margin register
			log.Printf("working on %s \n", app.Data.SymbolCode)
			//Shutting down app finally
			go func() {
				err = h.sessionDBService.session.workerAppService.AppShutDown(app)
				if err != nil {
					log.Printf("userDeleteAllAppHandler3 %v\n", err)
					http.Error(w, "Server Error", http.StatusInternalServerError)
				}
				wg.Done()
			}()
			//Deleting app from Mem DB
			err = h.sessionDBService.session.appMemDBService.DeleteApp(app.Data.ID)
			if err != nil {
				log.Printf("userDeleteAllAppHandler4 %v\n", err)
				http.Error(w, "Server Error", http.StatusInternalServerError)
				return
			}
			//Deleting app from Bolt DB
			err = h.sessionDBService.session.appBoltDBService.DeleteApp(app.Data.ID)
			if err != nil {
				log.Printf("userDeleteAppHandler6 %v\n", err)
				http.Error(w, "Server Error", http.StatusInternalServerError)
				return
			}
			//removing app id from user
			delete(user.ApIDs, app.Data.SymbolCode)
			log.Printf("Completed %s deleting!!! \n", app.Data.SymbolCode)
		}
		wg.Wait()
		h, err = h.syncParams(user, nil, "update")
		if err != nil {
			log.Printf("userDeleteAllAppHandler5 %v\n", err)
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}
		log.Printf("Completed All deleting!!! \n")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := deleteallappTmpl.Execute(w, r, GetAppDataStringified{SymbolCode: "All", DeleteMessage: "Are you sure you want to delete"}, user, nil); err != nil {
		log.Printf("userDeleteAllAppHandler6 %v\n", err)
		return
	}
	return
}
func (h TradeHandler) userDeleteAppHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := AlreadyLoggedIn(w, r, &h)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	id := r.FormValue("appid")
	Replaced := false
	//The first request is a Get request to delete which presents a confirmation form for the second request which is a POST request
	if strings.Contains(id, "POST") {
		id = strings.Replace(id, "POST", "", 1)
		//This is a way to show that this is the second request with a POST method confirmaing going ahead to delete
		Replaced = true
	}
	app, err := h.sessionDBService.session.appMemDBService.GetApp(user.ApIDs[id])
	if err == model.ErrAppNotFound {
		_, err := h.sessionDBService.session.appBoltDBService.GetApp(user.ApIDs[id])
		if err != nil {
			//removing app id from user
			log.Printf("userDeleteAllAppHandler2.0 %v\n", err)
			return
		}
		//Deleting app from Bolt DB
		err = h.sessionDBService.session.appBoltDBService.DeleteApp(user.ApIDs[id])
		if err != nil {
			log.Printf("userDeleteAppHandler2.1 %v\n", err)
			return
		}
		//removing app id from user
		delete(user.ApIDs, id)
		h, err = h.syncParams(user, nil, "update")
		if err != nil {
			log.Printf("userDeleteAppHandler2 %v\n", err)
			return
		}
		return
	} else if err != nil {
		log.Printf("userDeleteAppHandler2.2 %v\n", err)
		return
	}
	h, err = h.syncParams(user, nil, "update")
	if err != nil {
		log.Printf("userDeleteAppHandler2 %v\n", err)
		return
	}
	dat := GetAppDataStringified{
		SymbolCode:    id,
		DeleteMessage: "Are you sure you want to delete",
	}
	if Replaced {
		//Closing down websocket
		go func() {
			_ = h.sessionDBService.session.websocketUserService.CloseSocket(user, id)
		}()
		//Shutting down app finally
		err = h.sessionDBService.session.workerAppService.AppShutDown(app)
		if err != nil {
			log.Printf("userDeleteAppHandler3 %v\n", err)
			return
		}
		// Deleting App from Mem DB
		err = h.sessionDBService.session.appMemDBService.DeleteApp(app.Data.ID)
		if err != nil {
			log.Printf("userDeleteAppHandler4 %v\n", err)
			return
		}
		// Deleting App from Bolt DB
		err = h.sessionDBService.session.appBoltDBService.DeleteApp(app.Data.ID)
		if err != nil {
			log.Printf("userDeleteAppHandler6 %v\n", err)
			return
		}
		delete(user.ApIDs, app.Data.SymbolCode)
		h, err = h.syncParams(user, nil, "update")
		if err != nil {
			log.Printf("userDeleteAppHandler5 %v\n", err)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	//Presents a confirmation form to the user's UI for the deleting which sends a POST form back to this handler for the deleting
	if err := deleteappTmpl.Execute(w, r, dat, user, nil); err != nil {
		log.Printf("userDeleteAppHandler6 %v\n", err)
		return
	}
	return
}
func (h TradeHandler) userMessageAppHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := AlreadyLoggedIn(w, r, &h)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.ServeFile(w, r, "nohup.out")
	return
}
func (h TradeHandler) userGraphPointAppHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, "./graph/graphPoint.json")
	return
}
func (h TradeHandler) userGetAppListHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := AlreadyLoggedIn(w, r, &h)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	// process form submission
	sm := WebData{
		Symbols: make([]string, len(user.ApIDs)),
	}
	i := 0
	for k := range user.ApIDs {
		sm.Symbols[i] = k
		i++
	}
	if err := getapplistTmpl.Execute(w, r, sm, user, nil); err != nil {
		log.Printf("userGetAppListHandler1 %v\n", err)
		return
	}
	return
}

//WebData ...
type WebData struct {
	Symbols []string
}

// logoutHandler deletes the cookie
func (h TradeHandler) userlogoutHandler(w http.ResponseWriter, r *http.Request) {
	h.sessionDBService.session.logout(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}
func (h TradeHandler) syncParams(user *model.User, md *App, action string) (TradeHandler, error) {
	//Sync all IDs of user, session and app
	var (
		err error
	)
	//Sync IDs
	if md != nil {
		//user - app ID sync
		md.Data.UsrID = user.ID
		user.ApIDs[md.Data.SymbolCode] = md.Data.ID
		//session - app ID sync
		md.Data.SessID = h.sessionDBService.session.ID
		h.sessionDBService.session.ApID = md.Data.ID
	}

	//session - user ID sync
	user.SessID = h.sessionDBService.session.ID
	h.sessionDBService.session.UsrID = user.ID
	//cache user in session
	h.sessionDBService.session.cachedUser = user
	//Iniialling webSocket feeder, userDB, worker and appDB services for this user and session
	//initiallize other services sessions
	h.sessionDBService.session.websocketUserService.session = &h.sessionDBService.session
	h.sessionDBService.session.workerAppService.session = &h.sessionDBService.session
	h.sessionDBService.session.workerAppService.user = user
	//initiallize DB services
	//initiallize app DB service
	h.sessionDBService.session.appMemDBService.session = &h.sessionDBService.session
	h.sessionDBService.session.appBoltDBService.session = &h.sessionDBService.session
	h.sessionDBService.session.marginCalDBService.session = &h.sessionDBService.session
	//initiallize user DB service
	h.sessionDBService.session.userBoltDBService.session = &h.sessionDBService.session
	//store user, session and app
	if action == "update" {
		err = h.sessionDBService.session.userBoltDBService.UpdateUser(user)
		if err != nil {
			log.Printf("syncParams1 %v", err)
			return h, err
		}
		err = h.sessionDBService.UpdateSession(&h.sessionDBService.session)
		if err != nil {
			log.Printf("syncParams2 %v", err)
			return h, err
		}
		if md != nil {
			err = h.sessionDBService.session.appMemDBService.UpdateApp(md)
			if err != nil {
				log.Printf("syncParams3 %v", err)
				return h, err
			}
		}
	} else if action == "updateBolt" {
		// User is not stored in memory it is already being stored and updated directly into the bolt database created at signup so you can only update
		err = h.sessionDBService.session.userBoltDBService.UpdateUser(user)
		if err != nil {
			log.Printf("syncParams1 %v", err)
			return h, err
		}
		// There is no bolt database for session it completely reside in memory, it is recreated at each login
		err = h.sessionDBService.UpdateSession(&h.sessionDBService.session)
		if err != nil {
			log.Printf("syncParams2 %v", err)
			return h, err
		}
		if md != nil {
			// md has both memory and bolt database storages, the app functions interact with its memory stored copy for faster access to md
			err = h.sessionDBService.session.appMemDBService.UpdateApp(md)
			if err != nil {
				log.Printf("syncParams3 %v", err)
				return h, err
			}
			err = h.sessionDBService.session.appBoltDBService.UpdateApp(md.Data)
			if err != nil {
				log.Printf("syncParams4 %v", err)
				return h, err
			} else {
				log.Printf("mdUpdate done successfully")
			}
		}
	} else if action == "addBolt" && md != nil {
		ndId, err := h.sessionDBService.session.appBoltDBService.AddApp(md.Data)
		if err != nil {
			log.Printf("syncParams6 %v", err)
			return h, err
		}
		//user - app ID sync
		md.Data.ID = ndId
		user.ApIDs[md.Data.SymbolCode] = md.Data.ID
		h.sessionDBService.session.ApID = md.Data.ID

		err = h.sessionDBService.session.appMemDBService.AddApp(md)
		if err != nil {
			log.Printf("syncParams6 %v", err)
			return h, err
		}
		//store user, session and app
		err = h.sessionDBService.session.userBoltDBService.UpdateUser(user)
		if err != nil {
			log.Printf("syncParams1 %v", err)
			return h, err
		}
		err = h.sessionDBService.AddSession(&h.sessionDBService.session)
		if err != nil {
			log.Printf("syncParams5 %v", err)
			return h, err
		}
	} else {
		//store user, session and app
		err = h.sessionDBService.session.userBoltDBService.UpdateUser(user)
		if err != nil {
			log.Printf("syncParams1 %v", err)
			return h, err
		}
		err = h.sessionDBService.AddSession(&h.sessionDBService.session)
		if err != nil {
			log.Printf("syncParams5 %v", err)
			return h, err
		}
		if md != nil {
			err = h.sessionDBService.session.appMemDBService.AddApp(md)
			if err != nil {
				log.Printf("syncParams6 %v", err)
				return h, err
			}
		}
	}
	return h, nil
}

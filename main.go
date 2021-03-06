package main

import (
	"myhitbtcv4/chat"
	"myhitbtcv4/graph"
	//"myhitbtcv4/memory"
	"myhitbtcv4/bolt"
	//"myhitbtcv4/twilio"
	"flag"
	"log"
	"myhitbtcv4/app"
	"myhitbtcv4/model"
	"os"
)

func initDB() (model.UDBChans, app.MDDBChans, app.SDBChans, app.WUSChans, model.ABDBChans, model.MCalDBChans) {
	return model.UDBChans{
			AddDbChan:       make(chan model.UserDbData),
			UpdateDbChan:    make(chan model.UserDbData),
			GetDbChan:       make(chan model.UserDbData),
			DeleteDbChan:    make(chan model.UserDbData),
			GetDbByNameChan: make(chan model.UserDbByNameData),
		}, app.MDDBChans{
			AddOrUpdateDbChan: make(chan app.AppDbServiceVehicle),
			GetDbChan:         make(chan app.AppDbServiceVehicle),
			DeleteDbChan:      make(chan app.AppDbServiceVehicle),
		}, app.SDBChans{
			AddOrUpdateDbChan: make(chan app.SessionDbData),
			GetDbChan:         make(chan app.SessionDbData),
			DeleteDbChan:      make(chan app.SessionDbData),
		}, app.WUSChans{
			WriteSocketChan: make(chan app.WebsocketUSData),
			ReadSocketChan:  make(chan app.WebsocketUSData),
			CloseSocketChan: make(chan app.WebsocketUSData),
		}, model.ABDBChans{
			AddDbChan:    make(chan model.AppDataBoltVehicle),
			UpdateDbChan: make(chan model.AppDataBoltVehicle),
			GetDbChan:    make(chan model.AppDataBoltVehicle),
			DeleteDbChan: make(chan model.AppDataBoltVehicle),
		}, model.MCalDBChans{
			GetMarginDBChan: make(chan model.MarginDBVeh), //handler uses ths to get the margin DB
		}
}

//BTGUSD,BTCUSD,ETHUSD,BCHUSD,KMDUSD,XTZUSD,LSKUSD,NEOUSD,EOSUSD,XRPUSDT,ETCUSD,
//browse the app with: http://139.162.35.162:35259/
//BTGUSD,BTCUSD,ETHUSD,BCHUSD,KMDUSD,XTZUSD,LSKUSD,NEOUSD,EOSUSD,XRPUSDT,ETCUSD,TRXUSD,ZECUSD,XMRUSD,XLMUSD,XEMUSD,LTCUSD,ADAUSD
func main() {
	var (
		host string
	)
	//Get the port number from the environmental variable
	addr := os.Getenv("PORT2")
	//Because the handler is served in a goroutine, so the main function process will complete its process
	//and close while the goroutine is stil working. we must hang the main function process after the call
	//to Open() by waiting on a signal using the combination of 'sigs' os.signal channel and 'done' bool channel.
	flag.StringVar(&host, "host", "https://api.hitbtc.com", "Host is the base url for HitBTC")
	// tw := twilio.NewTwilioAPI("+" + "2349086790286")
	//Starting the getTicker goroutine

	UserBoltDBChans, AppMemDBChans, SessionMemDBChans, WebsocketUserChans, AppDataBoltDBChans, MarginCalDBChans := initDB()
	memAppDataChanChan := make(chan chan *model.AppData)
	uDBRCC := make(chan chan *model.User)
	// UserBoltServiceFunc proveds boltDB user storege services and as access speed is not required for user read/write there is no memory data storage for user
	go bolt.UserBoltDBServiceFunc(UserBoltDBChans, uDBRCC)
	//AppDataBoltServiceFunc provides in boltDB app storage service. Direct app read/write with boltDB was prevented for faster access achieved by memory read/write  
	go bolt.AppDataBoltDBServiceFunc(AppDataBoltDBChans, memAppDataChanChan)
	//AppDBServiceFunc privides in memory app storage for faster operation  
	go app.AppMemDBServiceFunc(AppMemDBChans, memAppDataChanChan)
	go app.SessionMemDBServiceFunc(SessionMemDBChans)
	go app.WebsocketUserServiceFunc(WebsocketUserChans)
	wGraphServ := graph.NewWorkerGraphService()
	go wGraphServ.GraphPointGen()
	go wGraphServ.GraphPopulation()
	UUIDChan := make(chan string)
	go app.RealtimeUUID(UUIDChan)
	sessDBS := app.NewSession(UserBoltDBChans, AppDataBoltDBChans,AppMemDBChans, WebsocketUserChans, MarginCalDBChans, SessionMemDBChans, sess)
	chatHdler := chat.NewChatHandler(sessDBS)
	h := app.NewTradeHandler(chatHdler, host, sessDBS, AppDataBoltDBChans, UUIDChan) //Passes the session to initialize a new instance of appHandler
	server := app.NewServer(addr, h)
	h.UserPowerUpHandler(uDBRCC, AppDataBoltDBChans.GetDbChan)
	//Start the webserver
	if err := server.Open(); err != nil {
		log.Fatalf("Unable to Open Server for listen and serve: %v", err)
	}
}

package main

import (
	//"myhitbtcv4/memory"
	"myhitbtcv4/bolt"
	//"myhitbtcv4/twilio"
	"flag"
	"log"
	"myhitbtcv4/app"
	"myhitbtcv4/model"
	"os"
)

func initDB() (model.UDBChans, app.MDDBChans, app.SDBChans, app.WUSChans, model.ABDBChans) {
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

	UserDBChans, AppDBChans, SessionDBChans, WebsocketUserChans, AppDataBoltChans := initDB()
	memAppDataChanChan := make(chan chan *model.AppData)
	uDBRCC := make(chan chan *model.User)
	go bolt.UserBoltServiceFunc(UserDBChans, uDBRCC)
	go bolt.AppDataBoltServiceFunc(AppDataBoltChans, memAppDataChanChan)
	//go memory.UserDBServiceFunc(UserDBChans)
	go app.AppDBServiceFunc(AppDBChans, memAppDataChanChan)
	go app.SessionDBServiceFunc(SessionDBChans)
	go app.WebsocketUserServiceFunc(WebsocketUserChans)
	UUIDChan := make(chan string)
	go app.RealtimeUUID(UUIDChan)
	sendMChan := make(chan chan app.MarginDB)
	registerMChan := make(chan app.MarginDBVeh)
	go app.MarginCal(registerMChan, sendMChan)
	h := app.NewTradeHandler(host, UserDBChans, AppDBChans, SessionDBChans, WebsocketUserChans, AppDataBoltChans, UUIDChan, sendMChan, registerMChan) //Passes the session to initialize a new instance of appHandler
	server := app.NewServer(addr, h)
	h.UserPowerUpHandler(uDBRCC, AppDataBoltChans.GetDbChan)
	//Start the webserver
	if err := server.Open(); err != nil {
		log.Fatalf("Unable to Open Server for listen and serve: %v", err)
	}
}

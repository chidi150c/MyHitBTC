package app

import (
	//"log"
	"myhitbtcv4/model"
)

//MarginDBVeh is the struct used by apps to reguster a chan for their sending of margin update
type MarginDBVeh struct { //This is the
	ID  model.AppID
	MChan       chan MarginVeh
	AddOrDelete string
}
type marginParam struct {
	SymbolCode  string
	SuccessfulOrders float64
	MadeProfitOrders float64
	MadeLostOrders   float64
	Value            float64
}
type AppMarginData struct {
	GrandMargin float64
	Margins     MarginDB
}

//MarginDB DB holds margin across app board
type MarginDB map[model.AppID]marginParam

func MarginCal(marginRegisterChan chan MarginDBVeh, SendMeMarginDBChan chan chan MarginDB) {
	marginDb := make(MarginDB) //
	marginRegister := make(map[model.AppID]MarginDBVeh)
	for {
		select {
		case dat := <-marginRegisterChan:
			if dat.AddOrDelete == "add" {
				marginRegister[dat.ID] = dat
			} else {
				delete(marginRegister, dat.ID)
				delete(marginDb, dat.ID)
			}
		case dbChan := <-SendMeMarginDBChan:
			for _, dat := range marginRegister {
				//log.Printf("waiting to get margin from %s ....", dat.SymbolCode)
				appMargin := <-dat.MChan
				marginDb[appMargin.ID] = appMargin.Margin
			}
			//log.Printf("waitng for handler to collect the margin DB....")
			dbChan <- marginDb
		}
	}
}

package model

type MCalDBChans struct {
	GetMarginDBChan    chan MarginDBVeh //handler uses ths to get the margin DB
}
type MarginParam struct {
	AppID            AppID
	SymbolCode       string
	SuccessfulOrders float64
	MadeProfitOrders float64
	MadeLostOrders   float64
	Value            float64
}
//MarginDBVeh is the struct used by apps to reguster a chan for their sending of margin update
type MarginDBVeh struct { //This is the
	ID             AppID
	MPChan         chan MarginParam
	User           *User
	MCalDBRespChan chan MarginDBResp
}
//MarginDBResp is the struct used by apps to reguster a chan for their sending of margin update
type MarginDBResp struct { //This is the
	ID       UserID
	MarginDB map[AppID]MarginParam //MarginDB DB holds margin across app board for a user
	Err      error
}
type MarginCalDBServicer interface{
	GetMargin(mar MarginDBVeh) (map[AppID]MarginParam, error)
}

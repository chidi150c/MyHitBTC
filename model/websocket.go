package model

import (
	"github.com/gorilla/websocket"
)

//User is the person using the site
type AppConn struct{
	Conn *websocket.Conn
	CloseAppSocketChan chan bool
}
//GetAppDataStringified ..
type GetAppDataStringified struct {
	Secret             string `json:"secret"`
	PublicKey          string `json:"publickey"`
	SymbolCode         string `json:"symbolcode"`
	SuccessfulOrders   string `json:"successfulorders"`
	MadeProfitOrders   string `json:"madeprofitorders"`
	TotalProfit        string `json:"totalprofit"`
	InstantProfit      string `json:"instantprofit"`
	Message            string `json:"message"`
	GoodBiz            string `json:"goodbiz"`
	LeastProfitMargin  string `json:"leastprofitmargin"`
	DisableTransaction string `json:"disabletransaction"`
	QuantityIncrement  string `json:"quantityincrement"`
	MessageFilter      string `json:"messagefilter"`
	NeverBought        string `json:"neverbought"`
	PendingA           string `json:"pendinga"`
	PendingB           string `json:"pendingb"`
	NeverSold          string `json:"neversold"`
	StopLostPoint      string `json:"stoplostpoint"`
	TrailPoints        string `json:"trailpoints"`
	SureTradeFactor    string `json:"suretradefactor"`
	TotalLost          string `json:"totallost"`
	InstantLost        string `json:"instantlost"`
	MadeLostOrders     string `json:"madelostorders"`
	Hodler             string `json:"hodler"`
	DeleteMessage      string `json:"deleteMessage"`
}
type WebsocketUSData struct{
	socket *websocket.Conn
	data *GetAppDataStringified
	RespChan chan error
}
type WUSChans struct {
	WriteSocketChan chan WebsocketUSData
	ReadSocketChan chan WebsocketUSData
	CloseSocketChan chan WebsocketUSData
}
type WebsocketUserServicer interface{
	WriteToSocket(user *User, dat *GetAppDataStringified) error
	CloseSocket(user *User, id string) error
}

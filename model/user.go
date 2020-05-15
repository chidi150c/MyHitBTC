package model

import (
	"github.com/gorilla/websocket"
)

type UserID uint64


//User is the person using the site
type AppConn struct{
	Conn *websocket.Conn
	CloseAppSocketChan chan bool
}
type User struct {
	ID            UserID
	SessID        SessionID
	ApIDs         map[string]AppID
	Username      string
	Password      string
	Firstname     string
	Lastname      string
	//channel       string
	Email         string
	ImageURL      string
	Token         string
	Url           string
	//Authenticated bool
	Expiry        int64
	//Role          string
	//Amount        string
	//Bank          string
	//BuyStatus     string
	//SellStatus    string
	//WalletAddress string
	Host          string
	Level         string
}

type UDBChans struct {
	AddDbChan chan UserDbData
	UpdateDbChan chan UserDbData
	GetDbChan         chan UserDbData
	GetDbByNameChan         chan UserDbByNameData
	DeleteDbChan      chan UserDbData
}

type UserDbByNameData struct {
	Username     string
	User       *User
	CallerChan chan UserDbResp
}
type UserDbData struct {
	UserID     UserID
	User       *User
	CallerChan chan UserDbResp
}

type UserDbResp struct {
	UserID UserID
	User   *User
	Err    error
}

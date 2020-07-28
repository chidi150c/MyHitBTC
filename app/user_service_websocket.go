package app

import (
	"errors"
	"myhitbtcv4/model"
	//"log"
)

func WebsocketUserServiceFunc(WebsocketUSChans WUSChans){
	for{
		select{
		case dat := <-WebsocketUSChans.WriteSocketChan:
			if dat.socket != nil{
				dat.RespChan <-dat.socket.WriteJSON(dat.data)
			}else{
				dat.RespChan <-errors.New("nil socket")
			}
		case dat := <-WebsocketUSChans.CloseSocketChan:
			dat.RespChan <-dat.socket.Close()
		}
	}
}
type WebsocketUserService struct {
	session *Session
}
func (w *WebsocketUserService) WriteToSocket(user *model.User, dat *GetAppDataStringified) error{
	//w.session.Authenticate()
	RespChan  :=  make(chan error)
	w.session.websocketUSChans.WriteSocketChan <-WebsocketUSData{w.session.Sockets[dat.SymbolCode].Conn, dat, RespChan}
	if err := <-RespChan; err != nil {
		//log.Printf("Could not Write to user socket: %v", err)
		return err
	}
	return nil
}
func (w *WebsocketUserService) CloseSocket(user *model.User, id string) error{
	user, err := w.session.Authenticate()
	if err != nil {
		return err
	}
	w.session.Sockets[id].CloseAppSocketChan <- true
	RespChan  :=  make(chan error)
	w.session.websocketUSChans.CloseSocketChan <-WebsocketUSData{w.session.Sockets[id].Conn, nil, RespChan}
	if err := <-RespChan; err != nil {
		//log.Printf("Could not Close user socket: %v", err)
		return err
	}
	delete(w.session.Sockets, id)
	return nil
}




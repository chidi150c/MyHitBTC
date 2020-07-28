package model

type MDDBChans struct {
	AddOrUpdateDbChan chan AppDbServiceVehicle
	GetDbChan         chan AppDbServiceVehicle
	DeleteDbChan      chan AppDbServiceVehicle
}
type AppDbServiceVehicle struct {
	AppID      AppID
	App        *App
	CallerChan chan AppDbResp
}
type AppDbResp struct {
	AppID AppID
	App   *App
	Err   error
}
type AppMemDBServicer interface{
	AddApp(App *App) error
	GetApp(id AppID) (*App, error)
	UpdateApp(App *App) error
	DeleteApp(id AppID) error
}
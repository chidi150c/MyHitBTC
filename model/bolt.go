
package model

type ABDBChans struct {
	AddDbChan    chan AppDataBoltVehicle
	UpdateDbChan chan AppDataBoltVehicle
	GetDbChan    chan AppDataBoltVehicle
	DeleteDbChan chan AppDataBoltVehicle
}
type AppDataBoltVehicle struct {
	AppID      AppID
	AppData    *AppData
	CallerChan chan AppDataResp
}
type AppDataResp struct {
	AppID   AppID
	AppData *AppData
	Err     error
}

type AppBoltDBServicer interface{
	AddApp(md *AppData) (AppID, error)
	GetApp(id AppID) (*AppData, error)
	UpdateApp(md *AppData) error
	DeleteApp(id AppID) error
}

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
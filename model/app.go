package model

const (
	publicKey = "edbe31ad902e7fa6f228e04ab8d0367c"
	secret    = "350ad6ceb288fe79dddf1cf7c7e04005"
)

type AppID uint64

//App is the main model that holds information about market trading which is shared or routed arround many methods for communicaton
type AppData struct {
	ID                 AppID
	UsrID              UserID
	SessID             SessionID
	PublicKey          string
	Secret             string
	Host               string
	SymbolCode         string
	Side               string
	DisableTransaction string
	MessageFilter      string
	MrktQuantity       float64
	MrktBuyPrice        float64
	MrktSellPrice      float64
	NeverBought        string
	NeverSold          string
	QuantityIncrement  float64
	Message            string
	TickSize           float64
	TakeLiquidityRate  float64
	HeartbeatBuy       string
	HeartbeatSell      string
	SuccessfulOrders   float64
	MadeProfitOrders   float64
	MadeLostOrders     float64
	StopLostPoint          float64
	BaseCurrency           string
	QuoteCurrency          string
	TrailPoints            float64
	LeastProfitMargin      float64
	SpinOutReason          string
	SureTradeFactor        float64
	Hodler                 string
	GoodBiz                float64
	AlternateData          float64
	InstantProfit          float64
	InstantLost            float64
	TotalProfit            float64
	TotalLost              float64
	PriceTradingStarted    string
	NextMarketBuyPoint     float64
	NextMarketSellPoint    float64
	MainStartPointSell     float64
	SoldQuantity           float64
	BoughtQuantity         float64
	MainStartPointBuy      float64
	MainQuantity           float64
	NextStartPointNegPrice float64
	NextStartPointPrice    float64
	ProfitPointFactor      float64
	HodlerQuantity         float64
	ProfitPriceUsed        string
	ProfitPrice            float64
	PendingA               []float64
	PendingB               []float64
}
type AppDataOld struct {
	ID                     AppID
	UsrID                  string
	SessID                 SessionID
	PublicKey              string
	Secret                 string
	Host                   string
	SymbolCode             string
	Side                   string
	DisableTransaction     string
	MessageFilter          string
	MrktQuantity           float64
	MrktBuyPrice            float64
	MrktSellPrice          float64
	NeverBought            string
	NeverSold              string
	QuantityIncrement      float64
	Message                string
	TickSize               float64
	TakeLiquidityRate      float64
	HeartbeatBuy           string
	HeartbeatSell          string
	SuccessfulOrders       float64
	MadeProfitOrders       float64
	MadeLostOrders         float64
	StopLostPoint          float64
	BaseCurrency           string
	QuoteCurrency          string
	TrailPoints            float64
	LeastProfitMargin      float64
	SpinOutReason          string
	SureTradeFactor        float64
	Hodler                 string
	GoodBiz                float64
	AlternateData          float64
	InstantProfit          float64
	InstantLost            float64
	TotalProfit            float64
	TotalLost              float64
	PriceTradingStarted    string
	NextMarketBuyPoint     float64
	NextMarketSellPoint    float64
	MainStartPointSell     float64
	SoldQuantity           float64
	BoughtQuantity         float64
	MainStartPointBuy      float64
	MainQuantity           float64
	NextStartPointNegPrice float64
	NextStartPointPrice    float64
	ProfitPointFactor      float64
	HodlerQuantity         float64
	ProfitPriceUsed        string
	ProfitPrice            float64
	Pending                []float64
}

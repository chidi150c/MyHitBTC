package api

type Symbols []Symbol
type Symbol struct {
	ID                   string `json:"id"`
	BaseCurrency         string `json:"baseCurrency"`
	QuoteCurrency        string `json:"quoteCurrency"`
	QuantityIncrement    string `json:"quantityIncrement"`
	TickSize             string `json:"tickSize"`
	TakeLiquidityRate    string `json:"takeLiquidityRate"`
	ProvideLiquidityRate string `json:"provideLiquidityRate"`
	FeeCurrency          string `json:"feeCurrency"`
	PreStartPrice        string `json:"-"`
}

type Balances []Balance
type Balance struct {
	Currency  string `json:"currency"`
	Available string `json:"available"`
	Reserved  string `json:"reserved"`
}

type Coinmarketcaps []Coinmarketcap
type Coinmarketcap struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Symbol           string `json:"symbol"`
	Rank             string `json:"rank"`
	PriceUsd         string `json:"price_usd"`
	PriceBtc         string `json:"price_btc"`
	Volume24hUsd     string `json:"24h_volume_usd"`
	MarketCapUsd     string `json:"market_cap_usd"`
	AvailableSupply  string `json:"available_supply"`
	TotalSupply      string `json:"total_supply"`
	PercentChange1h  string `json:"percent_change_1h"`
	PercentChange24h string `json:"percent_change_24h"`
	PercentChange7d  string `json:"percent_change_7d"`
	LastUpdated      string `json:"last_updated"`
}

type Currency struct {
	//Currency code
	Id       string `json:"id"`
	FullName string `json:"fullName"`
	//True for cryptocurrencies, false for fiat, ICO and others.
	Crypto bool `json:"crypto"`
	//True if cryptocurrency support generate adress or paymentId for deposits
	PayinEnabled bool `json:"payinEnabled"`
	//True if cryptocurrency requred use paymentId for deposits
	PayinPaymentId bool `json:"payinPaymentId"`
	//Confirmations count for cryptocurrency deposits
	PayinConfirmations int  `json:"payinConfirmations"`
	PayoutEnabled      bool `json:"payoutEnabled"`
	//True if cryptocurrency allow use paymentId for withdraw
	PayoutIsPaymentId bool `json:"payoutIsPaymentId"`
	TransferEnabled   bool `json:"transferEnabled"`
}

type Ticker struct {
	Symbol string `json:"symbol"`
	//Best ASK.
	Ask string `json:"ask"`
	//Best BID.
	Bid string `json:"bid"`
	//Last trade price
	Last string `json:"last"`
	//Min trade price of the last 24 hours.
	Low string `json:"low"`
	//Max trade price of the last 24 hours.
	High string `json:"high"`
	//Trade price 24 hours ago.
	Open string `json:"open"`
	//Trading volume in commoduty currency of the last 24 hours.
	Volume string `json:"volume"`
	//Trading volume in currency of the last 24 hours.
	VolumeQuoute string `json:"volumeQuoute"`
	//Actual timestamp.
	Timestamp string `json:"timestamp"`
}

type PublicTrade struct {
	Id       int    `json:"id"`
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
	//Enum : [sell, buy]
	Side      string `json:"side"`
	Timestamp string `json:"timestamp"`
}

type Label struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}
type OrderBooks []OrderBook
type OrderBook struct {
	//example:OrderedMap { "price": "0.012285", "size": "6.754" }
	Ask []Label `json:"ask"`
	//example:OrderedMap { "price": "0.012106", "size": "43.167" }
	Bid       []Label `json:"bid"`
	Timestamp string  `json:"timestamp"`
}

type TradingFee struct {
	TakeLiquidityRate    string `json:"takeLiquidityRate"`
	ProvideLiquidityRate string `json:"provideLiquidityRate"`
}

type Order struct {
	Id            int `json:"id"`
	ClientOrderId string `json:"clientOrderId"`
	Symbol        string `json:"symbol"`
	Side          string `json:"side"`
	Status        string `json:"status"`
	Type          string `json:"type"`
	TimeInForce   string `json:"timeInForce"`
	Quantity      string `json:"quantity"`
	Price         string `json:"price"`
	CumQuantity   string `json:"cumQuantity"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type RespAsMap []map[string]interface{}

//[map[symbol:BCHBTC side:buy quantity:0.001 cumQuantity:0.000 createdAt:2018-02-15T20:08:09.982Z id:17421017068 clientOrderId:ba2uh7m45ib0l65mu740cb2802fe timeInForce:GTC price:0.135745 updatedAt:2018-02-15T20:08:09.982Z status:new type:limit]]

type TradeReport struct {
	Id        string `json:"id"`
	Quantity  string `json:"quantity"`
	Price     string `json:"price"`
	Fee       string `json:"free"`
	Timestamp string `json:"timestamp"`
}

type TransferOk struct {
	Id string `json:"id"`
}

type Trade struct {
	Id            string `json:"id"`
	ClientOrderId string `json:"clientOrderId"`
	OrderId       int64  `json:"orderId"`
	Symbol        string `json:"symbol"`
	Side          string `json:"side"`
	Quantity      string `json:"quantity"`
	Fee           string `json:"fee"`
	Price         string `json:"price"`
	Timestamp     string `json:"timestamp"`
}

type Transaction struct {
	Id         string `json:"id"`
	Index      string `json:"index"`
	Currency   string `json:"currency"`
	Amount     string `json:"amount"`
	Fee        string `json:"fee"`
	NetworkFee string `json:"networkFee"`
	Address    string `json:"address"`
	PaymentId  string `json:"paymentId"`
	Hash       string `json:"hash"`
	Status     string `json:"status"`
	Type       string `json:"type"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type Address struct {
	Address   string `json:"address"`
	PaymentId string `json:"paymentId"`
}

type WithdrawConfirm struct {
	Result bool `json:"result"`
}

type Candles []Candle
type Candle struct {
	Timestamp   string `json:"timestamp"`
	Open        string `json:"open"`
	Close       string `json:"close"`
	Min         string `json:"min"`
	Max         string `json:"max"`
	Volume      string `json:"volume"`
	VolumeQuote string `json:"volumeQuote"`
}

type mainError struct {
	//- 500 - 504 - 503 - 2001 - 1001 - 1002 - 2001 - 10001
	Code        int32  `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

type AppError struct {
	Error mainError `json:"error"`
}

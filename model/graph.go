package model


// CoinmarketcapData ...
type CoinmarketcapData struct{
	Status CMarketStatus 
	Data CMarketDataSlice 
}



// CMarketStatus ...
type CMarketStatus struct{
	Timestamp string `json:"timestamp"`
	ErrorCode float64 `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Elapsed float64 `json:"elapsed"`
	CreditCount float64 `json:"credit_count"`
	Notice string `json:"notice"`
}
// CMNgn ...
type CMNgn struct{
	Price float64 `json:"price"`
	Volume24h float64 `json:"volume_24h"`
	PercentChange1h float64 `json:"percent_change_1h"`
	PercentChange24h float64 `json:"percent_change_24h"`
	PercentChange7d float64 `json:"percent_change_7d"`
	MarketCap float64 `json:"market_cap"`
	LastUpdated string `json:"last_updated"`
}
// CMQuote ...
type CMQuote struct{
	NGN CMNgn
}
// CMarketDataStruct ...
type CMarketDataStruct struct{
	ID	float64 `json:"id"`
	Name 	string `json:"name"`
	Symbol	string `json:"symbol"`
	Slug 	string `json:"slug"`
	NumMarketPairs	float64 `json:"num_market_pairs"`
	DateAdded string `json:"date_added"`
	Tags []string `json:"tags"`
	MaxAupply float64 `json:"max_supply"`
	CirculatingAupply float64 `json:"circulating_supply"`
	TotalAupply float64 `json:"total_supply"`
	Platform string `json:"platform"`
	CmcRank float64 `json:"cmc_rank"`
	LastUpdated string `json:"last_updated"`
	Quote CMQuote
}
// CMarketDataSlice ...
type CMarketDataSlice []CMarketDataStruct




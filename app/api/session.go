package api

import (
	//"fmt"
	"encoding/json"
	//"log"
	"net/http"
	netUrl "net/url"
	"strings"

	"github.com/pkg/errors"
)

type Simulate struct {
	PlacedOrder          float64
	LastPrice            float64
	QuoteCurrencyBalance float64
	BaseCurrencyBalance  float64
}
type Session struct {
	Auth                 []string
	BuyOrderChan         chan Simulate
	SellOrderChan        chan Simulate
	SellBalanceCheckChan chan Simulate
	BuyBalanceCheckChan  chan Simulate
}

func NewSession() *Session {
	return &Session{
		Auth:                 []string{},
		BuyOrderChan:         make(chan Simulate),
		SellOrderChan:        make(chan Simulate),
		SellBalanceCheckChan: make(chan Simulate, 1),
		BuyBalanceCheckChan:  make(chan Simulate, 1),
	}
}

func (s *Session) getSymbol(base, endpoint string) (*Symbol, error) {
	// Create Client
	client := &http.Client{}
	url := base + endpoint
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var symbol Symbol
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		Jdata := json.NewDecoder(Resp.Body)
		if endpoint == "" {
			return nil, errors.New("Get " + url + " Error: symbol empty please use the getSymbols method")
		} else {
			err := Jdata.Decode(&symbol)
			if err != nil {
				return nil, errors.Wrap(err, "Get "+url+" Error")
			}
			return &symbol, nil
		}
	}
	var edata AppError
	Edata := json.NewDecoder(Resp.Body)
	err = Edata.Decode(&edata)
	//fmt.Println(Resp.StatusCode)
	return nil, errors.New("Get " + url + " Error: " + Resp.Status + " App Error Message: " + edata.Error.Message + " App Error Description: " + edata.Error.Description)

}

func (s *Session) getTicker(base, endpoint string) (*Ticker, error) {
	// Create Client
	client := &http.Client{}
	url := base + endpoint
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var ticker Ticker
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		Jdata := json.NewDecoder(Resp.Body)
		if endpoint == "" {
			return nil, errors.New("Get " + url + " Error: ticker empty please use the getTickers method")
		} else {
			err := Jdata.Decode(&ticker)
			if err != nil {
				return nil, errors.Wrap(err, "Get "+url+" Error")
			}
			return &ticker, nil
		}
	}
	var edata AppError
	Edata := json.NewDecoder(Resp.Body)
	err = Edata.Decode(&edata)
	//fmt.Println(Resp.StatusCode)
	return nil, errors.New("Get " + url + " Error: " + Resp.Status + " App Error Message: " + edata.Error.Message + " App Error Description: " + edata.Error.Description)

}

func (s *Session) getSymbols(base string) (Symbols, error) {
	// Create Client
	client := &http.Client{}
	url := base
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		var symbols Symbols
		Jdata := json.NewDecoder(Resp.Body)
		err := Jdata.Decode(&symbols)
		if err != nil {
			return nil, errors.Wrap(err, "Get "+url+" Error")
		}
		return symbols, nil
	}
	var edata AppError
	Edata := json.NewDecoder(Resp.Body)
	err = Edata.Decode(&edata)
	//fmt.Println(Resp.StatusCode)
	return nil, errors.New("Get " + url + " Error: " + Resp.Status + " App Error Message: " + edata.Error.Message + " App Error Description: " + edata.Error.Description)
}

func (s *Session) getOrderBook(base, endpoint string) (*OrderBook, error) {
	// Create Client
	client := &http.Client{}
	url := base + endpoint + "?limit=1"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var orderBook OrderBook
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		Jdata := json.NewDecoder(Resp.Body)
		if endpoint == "" {
			return nil, errors.New("Get " + url + " Error: symbol empty please use the OrderBooks method")
		} else {
			err := Jdata.Decode(&orderBook)
			if err != nil {
				return nil, errors.Wrap(err, "Get "+url+" Error")
			}
		}
		return &orderBook, nil
	}
	var edata AppError
	Edata := json.NewDecoder(Resp.Body)
	err = Edata.Decode(&edata)
	//fmt.Println(Resp.StatusCode)
	return nil, errors.New("Get " + url + " Error: " + Resp.Status + " App Error Message: " + edata.Error.Message + " App Error Description: " + edata.Error.Description)

}

func (s *Session) getOrderBooks(base, endpoint string) (OrderBooks, error) {
	// Create Client
	client := &http.Client{}
	url := base + endpoint
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		var orderBooks OrderBooks
		var orderBook OrderBook
		Jdata := json.NewDecoder(Resp.Body)
		if endpoint == "" {
			err := Jdata.Decode(&orderBooks)
			if err != nil {
				return nil, errors.Wrap(err, "Get "+url+" Error")
			}
		} else {
			err := Jdata.Decode(&orderBook)
			if err != nil {
				return nil, errors.Wrap(err, "Get "+url+" Error")
			}
			orderBooks = append(orderBooks, orderBook)
		}
		return orderBooks, nil
	}
	var edata AppError
	Edata := json.NewDecoder(Resp.Body)
	err = Edata.Decode(&edata)
	//fmt.Println(Resp.StatusCode)
	return nil, errors.New("Get " + url + " Error: " + Resp.Status + " App Error Message: " + edata.Error.Message + " App Error Description: " + edata.Error.Description)

}

func (s *Session) getAddress(base, endpoint string) (*Address, error) {
	// Create Client
	client := &http.Client{}
	url := base + endpoint
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(s.Auth[0], s.Auth[1])
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		//var addresss Addresss
		var address Address
		Jdata := json.NewDecoder(Resp.Body)
		err := Jdata.Decode(&address)
		if err != nil {
			return nil, errors.Wrap(err, "Get "+url+" Error")
		}
		return &address, nil
	}
	//fmt.Println(Resp.StatusCode)
	return nil, errors.New("Get " + url + " Error: " + Resp.Status)
}

func (s *Session) getBalances(url string) (Balances, error) {
	// Create Client
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(s.Auth[0], s.Auth[1])
	//req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		var data Balances
		Jdata := json.NewDecoder(Resp.Body)
		err := Jdata.Decode(&data)
		if err != nil {
			return nil, errors.Wrap(err, "Get "+url+" Error")
		}
		return data, nil
	}
	var edata AppError
	Edata := json.NewDecoder(Resp.Body)
	err = Edata.Decode(&edata)
	//fmt.Println(Resp.StatusCode)
	return nil, errors.New("Get " + url + " Error: " + Resp.Status + " App Error Message: " + edata.Error.Message + " App Error Description: " + edata.Error.Description)

}

func (s *Session) getCoinMarketData() (Coinmarketcaps, error) {
	bitResp, err := http.Get("https://api.coinmarketcap.com/v1/ticker")
	if err != nil {
		return nil, err
	}
	bitdecoder := json.NewDecoder(bitResp.Body)
	var data Coinmarketcaps
	err = bitdecoder.Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, err
}

func (s *Session) postTransfer(url string, Msg ...string) (*TransferOk, error) {
	var MsgData = netUrl.Values{}
	MsgData.Set("currency", Msg[0])
	MsgData.Set("amount", Msg[1])
	MsgData.Set("type", Msg[2])
	msgDataReader := *strings.NewReader(MsgData.Encode())
	// Create Client
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, &msgDataReader)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(s.Auth[0], s.Auth[1])
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var data TransferOk
	var edata AppError
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		Jdata := json.NewDecoder(Resp.Body)
		err := Jdata.Decode(&data)
		if err != nil {
			return nil, errors.Wrap(err, "Get "+url+" Error")
		}
		return &data, nil
	}
	Edata := json.NewDecoder(Resp.Body)
	err = Edata.Decode(&edata)
	//fmt.Println(Resp.StatusCode)
	return &data, errors.New("Get " + url + " Error: " + Resp.Status + "App Error Message: " + edata.Error.Message + "App Error Description: " + edata.Error.Description)
}
func (s *Session) postNewOrder(url string, dat Order) (*Order, error) {
	var MsgData = netUrl.Values{}
	MsgData.Set("clientOrderId", dat.ClientOrderId)
	MsgData.Set("symbol", dat.Symbol)
	MsgData.Set("side", dat.Side)
	MsgData.Set("timeInForce", "GTC")
	MsgData.Set("quantity", dat.Quantity)
	MsgData.Set("price", dat.Price)
	MsgData.Set("type", dat.Type)
	msgDataReader := *strings.NewReader(MsgData.Encode())
	// Create Client
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, &msgDataReader)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(s.Auth[0], s.Auth[1])
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var data Order
	var edata AppError
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		Jdata := json.NewDecoder(Resp.Body)
		err := Jdata.Decode(&data)
		if err != nil {
			return nil, errors.Wrap(err, "Get "+url+" Error")
		}
		return &data, nil
	}
	Edata := json.NewDecoder(Resp.Body)
	err = Edata.Decode(&edata)
	return &data, errors.New("Get " + url + " Error: " + Resp.Status + " App Error Message: " + edata.Error.Message + " App Error Description: " + edata.Error.Description)
}
func (s *Session) putNewOrder(url string, dat Order) (*Order, error) {
	var MsgData = netUrl.Values{}
	MsgData.Set("symbol", dat.Symbol)
	MsgData.Set("side", dat.Side)
	MsgData.Set("timeInForce", "GTC")
	MsgData.Set("quantity", dat.Quantity)
	MsgData.Set("price", dat.Price)
	MsgData.Set("type", dat.Type)
	msgDataReader := *strings.NewReader(MsgData.Encode())
	// Create Client
	client := &http.Client{}
	req, _ := http.NewRequest("PUT", url, &msgDataReader)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(s.Auth[0], s.Auth[1])
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var data Order
	var edata AppError
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		Jdata := json.NewDecoder(Resp.Body)
		err := Jdata.Decode(&data)
		if err != nil {
			return nil, errors.Wrap(err, "Get "+url+" Error")
		}
		return &data, nil
	}
	Edata := json.NewDecoder(Resp.Body)
	err = Edata.Decode(&edata)
	//fmt.Println(Resp.StatusCode)
	return &data, errors.New("Get " + url + " Error: " + Resp.Status + " App Error Message: " + edata.Error.Message + " App Error Description: " + edata.Error.Description)
}

func (s *Session) getOrder(url string) (*Order, error) {
	// Create Client
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(s.Auth[0], s.Auth[1])
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var order3 interface{}
	var order2 RespAsMap
	order2 = make(RespAsMap, 1)
	order2[0] = make(map[string]interface{})
	var order Order
	var edata AppError
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		Jdata := json.NewDecoder(Resp.Body)
		err := Jdata.Decode(&order3)
		if err != nil {
			return nil, errors.Wrap(err, "Get "+url+" Error")
		}
		switch order3.(type) {
		case RespAsMap:
			order2 = order3.(RespAsMap)
			order = Order{
				Id:            order2[0]["id"].(int),
				ClientOrderId: order2[0]["clientOrderId"].(string),
				Symbol:        order2[0]["symbol"].(string),
				Side:          order2[0]["side"].(string),
				Status:        order2[0]["status"].(string),
				Type:          order2[0]["type"].(string),
				TimeInForce:   order2[0]["timeInForce"].(string),
				Quantity:      order2[0]["quantity"].(string),
				Price:         order2[0]["price"].(string),
				CumQuantity:   order2[0]["cumQuantity"].(string),
				CreatedAt:     order2[0]["createdAt"].(string),
				UpdatedAt:     order2[0]["updatedAt"].(string),
			}
		case Order:
			order = order3.(Order)
		}
		return &order, nil
	}
	Edata := json.NewDecoder(Resp.Body)
	err = Edata.Decode(&edata)
	//fmt.Println(Resp.StatusCode)
	return nil, errors.New("Get " + url + " Error: " + Resp.Status + " App Error Message: " + edata.Error.Message + " App Error Description: " + edata.Error.Description)
}

func (s *Session) delete(url string) (*Order, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(s.Auth[0], s.Auth[1])
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var order3 interface{}
	var order2 RespAsMap
	order2 = make(RespAsMap, 1)
	order2[0] = make(map[string]interface{})
	var order Order
	var edata AppError
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		Jdata := json.NewDecoder(Resp.Body)
		err := Jdata.Decode(&order3)
		if err != nil {
			return nil, errors.Wrap(err, "Get "+url+" Error")
		}
		//log.Println(order3)
		switch order3.(type) {
		case RespAsMap:
			order2 = order3.(RespAsMap)
			order = Order{
				Id:            order2[0]["id"].(int),
				ClientOrderId: order2[0]["clientOrderId"].(string),
				Symbol:        order2[0]["symbol"].(string),
				Side:          order2[0]["side"].(string),
				Status:        order2[0]["status"].(string),
				Type:          order2[0]["type"].(string),
				TimeInForce:   order2[0]["timeInForce"].(string),
				Quantity:      order2[0]["quantity"].(string),
				Price:         order2[0]["price"].(string),
				CumQuantity:   order2[0]["cumQuantity"].(string),
				CreatedAt:     order2[0]["createdAt"].(string),
				UpdatedAt:     order2[0]["updatedAt"].(string),
			}
		case Order:
			order = order3.(Order)
		}

		return &order, nil
	}
	Edata := json.NewDecoder(Resp.Body)
	err = Edata.Decode(&edata)
	//fmt.Println(Resp.StatusCode)
	return nil, errors.New("Get " + url + " Error: " + Resp.Status + " App Error Message: " + edata.Error.Message + " App Error Description: " + edata.Error.Description)
}

func (s *Session) postWithdraw(url string, dat Transaction) (*Transaction, error) {
	var MsgData = netUrl.Values{}
	MsgData.Set("currency", dat.Currency)
	MsgData.Set("amount", dat.Amount)
	MsgData.Set("address", dat.Address)
	MsgData.Set("networkFee", dat.NetworkFee)
	msgDataReader := *strings.NewReader(MsgData.Encode())
	// Create Client
	client := &http.Client{}
	req, _ := http.NewRequest("PUT", url, &msgDataReader)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(s.Auth[0], s.Auth[1])
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var data Transaction
	var edata AppError
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		Jdata := json.NewDecoder(Resp.Body)
		err := Jdata.Decode(&data)
		if err != nil {
			return nil, errors.Wrap(err, "Get "+url+" Error")
		}
		return &data, nil
	}
	Edata := json.NewDecoder(Resp.Body)
	err = Edata.Decode(&edata)
	//fmt.Println(Resp.StatusCode)
	return &data, errors.New("Get " + url + " Error: " + Resp.Status + " App Error Message: " + edata.Error.Message + " App Error Description: " + edata.Error.Description)

}

func (s *Session) getTransaction(url string) (*Transaction, error) {
	// Create Client
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(s.Auth[0], s.Auth[1])
	//req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var data Transaction
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		Jdata := json.NewDecoder(Resp.Body)
		err := Jdata.Decode(&data)
		if err != nil {
			return nil, errors.Wrap(err, "Get "+url+" Error")
		}
		return &data, nil
	}
	var edata AppError
	Edata := json.NewDecoder(Resp.Body)
	err = Edata.Decode(&edata)
	//fmt.Println(Resp.StatusCode)
	return &data, errors.New("Get " + url + " Error: " + Resp.Status + " App Error Message: " + edata.Error.Message + " App Error Description: " + edata.Error.Description)

}

func (s *Session) getCandles(url string) (*Candles, error) {
	// Create Client
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	//req.SetBasicAuth(s.Auth[0], s.Auth[1])
	//req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	Resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var data Candles
	if Resp.StatusCode >= 200 && Resp.StatusCode < 300 {
		Jdata := json.NewDecoder(Resp.Body)
		err := Jdata.Decode(&data)
		if err != nil {
			return nil, errors.Wrap(err, "Get "+url+" Error")
		}
		return &data, nil
	}
	var edata AppError
	Edata := json.NewDecoder(Resp.Body)
	err = Edata.Decode(&edata)
	//fmt.Println(Resp.StatusCode)
	return &data, errors.New("Get " + url + " Error: " + Resp.Status + " App Error Message: " + edata.Error.Message + " App Error Description: " + edata.Error.Description)

}

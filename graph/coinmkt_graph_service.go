package graph

import (
	"strconv"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"myhitbtcv4/model"
	"net/url"
)

//GraphService performs graph operations and services
type WorkerGraphService struct {
	CloseDown chan bool
	GraphPoint model.CoinmarketcapData
	GraphPointChan chan model.CoinmarketcapData
}
func NewWorkerGraphService()*WorkerGraphService{
	return&WorkerGraphService{
		CloseDown: make(chan bool),
		GraphPoint: model.CoinmarketcapData{
			Data: make(model.CMarketDataSlice, 1),
		},
		GraphPointChan: make(chan model.CoinmarketcapData),
	}
}

//ReceiveGraphPointFromAPI reieves graph API from provider and sends it into the app
func (g *WorkerGraphService) GraphPointGen() {
	for {
		select {
		case <-g.CloseDown:
			return
		default:
			client := &http.Client{}
			req, err := http.NewRequest("GET","https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest", nil)
			if err != nil {
			log.Print(err)
			os.Exit(1)
			}			
			q := url.Values{}
		//   q.Add("symbol", "BTC")
			q.Add("convert", "NGN")
			q.Add("limit", "1")
		//   q.Add("convert", "USD")			
			req.Header.Set("Accepts", "application/json")
			req.Header.Add("X-CMC_PRO_API_KEY", "11034a6e-ce2c-462e-8366-5c9d0c940c09")
			req.Header.Add("Accept-Encoding", "deflate") //, gzip
			req.URL.RawQuery = q.Encode()			
			resp, err := client.Do(req);
			if err != nil {
				fmt.Println("Error sending request to server: %v", err)
				time.Sleep(time.Second * 180)
				continue
			}
			bitdecoder := json.NewDecoder(resp.Body)
			err = bitdecoder.Decode(&g.GraphPoint)
			if err != nil {
				log.Fatalf("could no json decode into GraphPoint from bitResp.Body Error is %v", err)
			}
			t := time.Now().Unix() * 1000
			g.GraphPoint.Status.Timestamp = strconv.FormatInt(t, 10)
			// time.Parse("2006-01-02T15:04:05.000-0700", g.GraphPoint.Status.Timestamp)
			// sstr := strings.Split(g.GraphPoint.Status.Timestamp, ".")
			// sstr[0] = strings.Replace(sstr[0], "-", ",", -1)
			// sstr[0] = strings.Replace(sstr[0], ":", ",", -1)
			// g.GraphPoint.Status.Timestamp = strings.Replace(sstr[0], "T", ",", -1)
			// g.GraphPointChan = nil
			g.GraphPointChan <- g.GraphPoint
			time.Sleep(time.Second * 2 * 60)
		}
	}
}

//CreateGraph4Coinmkt the graph out of the channel of graph api
func (g *WorkerGraphService) GraphPopulation() {
	headfootFs, err := os.OpenFile("./graph/graphPoint.json", os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer headfootFs.Close()

	var bodydata = []string{}

	//get statistics to know the size of file with .Size() method later
	hfStat, e := headfootFs.Stat()
	if e != nil {
		log.Fatal(e)
	}

	var buf = make([]byte, hfStat.Size())
	n, _ := headfootFs.Read(buf)
	if n != 0 {
		b1 := bytes.Replace(buf, []byte("[["), []byte("["), 1)
		b2 := bytes.Replace(b1, []byte("]]"), []byte("]"), 1)
		b3 := bytes.Split(b2, []byte(",\n"))

		fmt.Println()
		for _, v := range b3 {
			bodydata = append(bodydata, string(v))
		}
	}
	for {
		for point := range g.GraphPointChan {
			pointData := fmt.Sprintf("%.2f", point.Data[0].Quote.NGN.Price) //already converted to NGN
			bodydata = append(bodydata, fmt.Sprintf("[%s,%s]", point.Status.Timestamp, pointData))
			bodydatamodified := strings.Join(bodydata, ",\n")
			headNfoot := fmt.Sprintf("[%s]", bodydatamodified)
			_, err = headfootFs.WriteAt([]byte(headNfoot), 0)
		}
	}

}

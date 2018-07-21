package netdania

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"./mini"

	xj "github.com/basgys/goxml2json"
)

type Datafeed struct {
	DataInfo QuoteStrap `json:"datafeed"`
}

type QuoteStrap struct {
	QuotesInfo []Quote `json:"quote"`
}

type Quote struct {
	Symbol string `json:"-f25"`
	Bid    string `json:"-f10"`
	Ask    string `json:"-f11"`
}

func GetRateMapFromNetdaniaSources(source string, c chan map[string]map[string]float64) {
	var url string
	if source == "netdania_fxa" {
		//fmt.Println("Source is netdania_fxa")
		url = "http://balancer.netdania.com/StreamingServer/StreamingServer?xml=price&group=moneynetint&pass=y12mo148not&source=netdania_fxa"
		url += "&symbols=USDAUD|USDEUR|USDAED|USDBRL|USDCAD|USDCHF|USDCOP|USDCZK|USDDKK|USDHKD|USDHRK|USDILS|USDJPY|USDPLN|USDRUB|USDSEK|USDTHB|USDZAR|USDGBP"
	} else {
		//fmt.Println("Source is ebs")
		url = "http://balancer.netdania.com/StreamingServer/StreamingServer?xml=price&group=moneynetint&pass=y12mo148not&source=ebs"
		url += "&symbols=USDCNY|USDEGP|USDINR|USDVND|USDXAF"
	}
	url += "&fields=10|11|14|15|25|4|2|3|1&time=dd HH:mm:ss&tzone=GMT"

	req, _ := http.NewRequest("GET", url, nil)
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	xml := strings.NewReader(string(body))
	jsonConverted, err := xj.Convert(xml)
	if err != nil {
		panic("That's embarrassing...")
	}

	readBuf, _ := ioutil.ReadAll(jsonConverted)

	unmarshalledJson := &Datafeed{}
	err = json.Unmarshal(readBuf, unmarshalledJson)

	mini.Check(err)

	newRatesMap := make(map[string]map[string]float64)

	for _, elem := range unmarshalledJson.DataInfo.QuotesInfo {
		if elem.Symbol != "N/A" {

			s := strings.Split(elem.Symbol, "/")
			splitSymbol := s[1]

			newRatesMap[splitSymbol] = make(map[string]float64)

			askRate, err := strconv.ParseFloat(elem.Ask, 64)
			mini.Check(err)
			newAskRate := mini.ToFixed(askRate, 4)

			bidRate, err := strconv.ParseFloat(elem.Bid, 64)
			mini.Check(err)
			newBidRate := mini.ToFixed(bidRate, 4)

			avgRate := (askRate + bidRate) / 2
			newAvgRate := mini.ToFixed(avgRate, 4)

			newRatesMap[splitSymbol]["rate"] = newAvgRate
			newRatesMap[splitSymbol]["rate_ask"] = newAskRate
			newRatesMap[splitSymbol]["rate_bid"] = newBidRate
		}

	}

	c <- newRatesMap
}

func GetRateMapFromNetdania(c chan map[string]map[string]float64) {
	ch := make(chan map[string]map[string]float64)
	go GetRateMapFromNetdaniaSources("ebs", ch)
	go GetRateMapFromNetdaniaSources("netdania_fxa", ch)
	x, y := <-ch, <-ch

	for key, value := range x {
		y[key] = value
	}

	c <- y
}

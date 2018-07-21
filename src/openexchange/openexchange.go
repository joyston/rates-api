package openexchange

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"functions"
	"readconfig"
)

type Result struct {
	ResultMap map[string]string
	Err       error
}

func GetOEdata() <-chan Result {
	c := make(chan Result)
	newRatesMap := make(map[string]string)

	retrievedConfig := readconfig.GetConfigInfo()

	go func() {

		// FOR PRODUCTION
		oeUrl := "https://openexchangerates.org/api/latest.json?app_id=***********"

		timeout := time.Duration(retrievedConfig.Server.ServerTimeout * time.Second)
		client := http.Client{
			Timeout: timeout,
		}

		requestTime := time.Now()
		res, errRequest := client.Get(oeUrl)

		responseCode := res.StatusCode

		errFlag := functions.CheckError(errRequest)
		if errFlag == 1 {
			functions.SourceLogs(requestTime.String(), "OpenExchange", responseCode, errRequest.Error())
			c <- Result{ResultMap: nil, Err: errRequest}
			return
		}
		// res, doErr := http.DefaultClient.Do(req)
		// errDoFlag := functions.CheckError(doErr)
		// if errDoFlag == 1 {
		// 	c <- Result{result: nil, err: doErr}
		// 	return
		// }
		defer res.Body.Close()
		body, errRead := ioutil.ReadAll(res.Body)
		errBodyFlag := functions.CheckError(errRead)
		if errBodyFlag == 1 {
			c <- Result{ResultMap: nil, Err: errRead}
			return
		}
		functions.SourceLogs(requestTime.String(), "OpenExchange", responseCode, string(body))
		// COMMENT TILL HERE DURING TESTING

		//FOR TESTING -> Read from a file
		// f, errOpen := os.Open("../bouncer/tmp/openExchangeData")
		// if errOpen != nil {
		// 	panic(errOpen)
		// }
		// body, _ := ioutil.ReadAll(f)
		//COMMENT TILL HERE FOR PRODUCTION

		var objmap map[string]*json.RawMessage
		errUnmarshal := json.Unmarshal(body, &objmap)
		errUnmarshalFlag := functions.CheckError(errUnmarshal)
		if errUnmarshalFlag == 1 {
			c <- Result{ResultMap: nil, Err: errUnmarshal}
			return
		}

		var ratesMap map[string]float64
		errrates := json.Unmarshal(*objmap["rates"], &ratesMap)
		errratesFlag := functions.CheckError(errrates)
		if errratesFlag == 1 {
			c <- Result{ResultMap: nil, Err: errrates}
			return
		}

		for key, value := range ratesMap {
			valuePrecision := "%." + strconv.Itoa(retrievedConfig.Miscellaneous.Precision) + "f"
			stringvalue := fmt.Sprintf(valuePrecision, value)
			newRatesMap[key] = stringvalue
		}
		c <- Result{ResultMap: newRatesMap, Err: nil}
	}()

	return c
}

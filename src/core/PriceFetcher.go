package core

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/garry-sharp/assessment/src/logger"
)

// json structs
type Response struct {
	Result map[string]struct {
		C []string `json:"c"`
	} `json:"result"`
	Error []interface{} `json:"error"`
}

func GetPrice(asset string, quote string) (float32, error) {
	resp, err := http.Get("https://api.kraken.com/0/public/Ticker?pair=" + asset + quote)
	if err != nil {
		logger.Log(err.Error(), logger.Fatal, logger.Price)
	} else {
		var response Response
		body, _ := ioutil.ReadAll(resp.Body)
		decodeError := json.Unmarshal(body, &response)
		if decodeError != nil {
			logger.Log(decodeError.Error(), logger.Fatal, logger.Price)
		} else {
			if len(response.Error) != 0 {
				return 0, errors.New("Unable to load price")
			} else {
				for _, v := range response.Result {
					price, err := strconv.ParseFloat(v.C[0], 32)
					if err != nil {
						logger.Log("Cannot parse price into float", logger.Panic, logger.Price)
						return 0, errors.New("Cannot parse price into float")
					} else {
						return float32(price), nil
					}
				}
			}
		}
	}

	return 0, nil
}

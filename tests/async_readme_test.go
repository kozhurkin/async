package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kozhurkin/async"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"testing"
	"time"
)

func TestReadmeSlice(t *testing.T) {
	symbols := []string{"BTC", "ETH", "BNB", "DOGE"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// concurrency = 0 means that all tasks will be executed at the same time in parallel
	concurrency := 0
	results, err := async.Slice(ctx, concurrency, symbols, func(i int, ticker string) (float64, error) {
		resp, err := http.Get("https://api.binance.com/api/v3/ticker/price?symbol=" + ticker + "USDT")
		if err != nil {
			return 0, err
		}
		var info struct {
			Price float64 `json:"price,string"`
		}
		err = json.NewDecoder(resp.Body).Decode(&info)
		return info.Price, err
	})

	fmt.Println(results, err)
	// [64950 3340.74 556.5 0.17076] <nil>
}

func TestReadmeSliceMapped(t *testing.T) {

	videos := []string{
		"XqZsoesa55w",
		"kJQP7kiw5Fk",
		"RgKAFK5djSk",
		"JGwWNGJdvx8",
		"ThisIsError",
		"9bZkp7q19f0",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// concurrency = 2 means that no more than 2 tasks can be performed at a time
	concurrency := 2
	responses, err := async.SliceMapped(ctx, concurrency, videos, func(i int, vid string) (int, error) {
		resp, err := http.Get("https://www.youtube.com/watch?v=" + vid)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return 0, err
			}
			if res := regexp.MustCompile(`viewCount":"(\d+)`).FindSubmatch(bodyBytes); res != nil {
				return strconv.Atoi(string(res[1]))
			}
		}
		return 0, fmt.Errorf(`can't parse "%v" views`, vid)
	})

	fmt.Println(responses)
	fmt.Println(err)
	// map[9bZkp7q19f0:0 JGwWNGJdvx8:0 RgKAFK5djSk:6211818831 ThisIsError:0 XqZsoesa55w:14277740491 kJQP7kiw5Fk:8404577810]
	// can't parse "ThisIsError" views
}

func TestReadmeFuncs(t *testing.T) {
	var binancePrices []struct {
		Symbol string  `json:"symbol"`
		Price  float64 `json:"lastPrice,string"`
		Volume float64 `json:"volume,string"`
	}
	var huobiPrices struct {
		Data []struct {
			Symbol string  `json:"symbol"`
			Price  float64 `json:"close"`
			Volume float64 `json:"vol"`
		} `json:"data"`
	}

	concurrency := 0
	_, err := async.Funcs(context.Background(), concurrency, func() (interface{}, error) {
		resp, err := http.Get("https://api.binance.com/api/v3/ticker/24hr")
		if err == nil {
			err = json.NewDecoder(resp.Body).Decode(&binancePrices)
		}
		return binancePrices, err
	}, func() (interface{}, error) {
		resp, err := http.Get("https://api.huobi.pro/market/tickers")
		if err == nil {
			err = json.NewDecoder(resp.Body).Decode(&huobiPrices)
		}
		return binancePrices, err
	})

	fmt.Println(binancePrices[0:3])
	fmt.Println(huobiPrices.Data[0:3])
	fmt.Println(err)
	// [{ETHBTC 0.05135 23536.0039} {LTCBTC 0.001364 98422.194} {BNBBTC 0.008554 23751.836}]
	// [{sylousdt 0.003456 475985.8738069513} {zigusdt 0.09386 57274.5040727801} {walletusdt 0.014629 502630.9344168179}]
	// <nil>
}

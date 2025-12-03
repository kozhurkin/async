# async

A helper library for asynchronous tasks in Go, powered by generics.

[![async status](https://github.com/kozhurkin/async/actions/workflows/tests.yml/badge.svg)](https://github.com/kozhurkin/async/actions)

* `AsyncToArray()`
* `AsyncToMap()`
* `AsyncFuncs()`


---

## Features

- Simple and intuitive API
- Supports task concurrency limits
- Context-aware with timeout and cancellation support
- Works with slices, maps, and standalone functions
- Requires Go 1.18+ (generics)

---

## Installation

```bash
go get github.com/kozhurkin/async
```

Usage
-----

#### async.AsyncToArray()
``` golang
import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/kozhurkin/async"
)

func main() {
    coins := []string{"BTC", "ETH", "BNB", "DOGE"}

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // concurrency = 0 means unlimited parallelism
    concurrency := 0

    results, err := async.AsyncToArray(ctx, concurrency, coins, func(ctx context.Context, i int, coin string) (float64, error) {
        resp, err := http.Get("https://api.binance.com/api/v3/ticker/price?symbol=" + coin + "USDT")
        if err != nil {
            return 0, err
        }
        defer resp.Body.Close()

        var info struct {
            Price float64 `json:"price,string"`
        }
        err = json.NewDecoder(resp.Body).Decode(&info)
        return info.Price, err
    })

    fmt.Println(results, err)
    // Output: [64950 3340.74 556.5 0.17076] <nil>
}
```

#### async.AsyncToMap()
``` golang
import (
    "context"
    "fmt"
    "io"
    "net/http"
    "regexp"
    "strconv"
    "time"

    "github.com/kozhurkin/async"
)

func main() {
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

    concurrency := 2

    responses, err := async.AsyncToMap(ctx, concurrency, videos, func(ctx context.Context, i int, vid string) (int, error) {
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
    // Output:
    // map[9bZkp7q19f0:0 JGwWNGJdvx8:0 RgKAFK5djSk:6211818831 ThisIsError:0 XqZsoesa55w:14277740491 kJQP7kiw5Fk:8404577810]
    // can't parse "ThisIsError" views
}
```

#### async.AsyncFuncs()
``` golang
import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/kozhurkin/async"
)

func main() {
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

    concurrency := 2

    results, err := async.AsyncFuncs(
        context.Background(),
        concurrency,
        func(ctx context.Context) (interface{}, error) {
            resp, err := http.Get("https://api.binance.com/api/v3/ticker/24hr")
            if err == nil {
                defer resp.Body.Close()
                err = json.NewDecoder(resp.Body).Decode(&binancePrices)
            }
            return binancePrices, err
        },
        func(ctx context.Context) (interface{}, error) {
            resp, err := http.Get("https://api.huobi.pro/market/tickers")
            if err == nil {
                defer resp.Body.Close()
                err = json.NewDecoder(resp.Body).Decode(&huobiPrices)
            }
            return huobiPrices, err
        },
    )

    fmt.Println(binancePrices[:3])
    fmt.Println(huobiPrices.Data[:3])
    fmt.Println(err)
    // Output:
    // [{ETHBTC 0.05135 23536.0039} {LTCBTC 0.001364 98422.194} {BNBBTC 0.008554 23751.836}]
    // [{sylousdt 0.003456 475985.87} {zigusdt 0.09386 57274.50} {walletusdt 0.014629 502630.93}]
    // <nil>
}
```

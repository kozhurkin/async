# async

helper for asynchronous work in golang powered by generics.

[![async status](https://github.com/kozhurkin/async/actions/workflows/test.yml/badge.svg)](https://github.com/kozhurkin/async/actions)

* `AsyncToArray()`
* `AsyncToMap()`

Installing
----------

	go get github.com/kozhurkin/async

Usage
-----

#### AsyncToArray()
``` golang
import github.com/kozhurkin/async

func main() {
    tickers := []string{"BTC", "ETH", "BNB", "DOGE"}

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // concurrency = 0 means that all tasks will be executed at the same time in parallel
    concurrency := 0
    results, err := async.AsyncToArray(ctx, tickers, func(i int, ticker string) (float64, error) {
        resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%vUSDT", ticker))
        if err != nil {
            return 0, err
        }
        var info struct {
            Symbol string  `json:"symbol"`
            Price  float64 `json:"price,string"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
            return 0, err
        }
        return info.Price, nil
    }, concurrency)

    fmt.Println(results, err)
    // [64950 3340.74 556.5 0.17076] <nil>
}
```
#### AsyncToMap()

``` golang
import github.com/kozhurkin/async

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

    // concurrency = 0 means that all tasks will be executed at the same time
    concurrency := 2
    responses, err := async.AsyncToMap(ctx, videos, func(i int, vid string) (int, error) {
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
            if res := regexp.MustCompile(`viewCount":"\d+`).Find(bodyBytes); res != nil {
                return strconv.Atoi(string(res[12:]))
            }
        }
        return 0, fmt.Errorf(`can't parse "%v" views`, vid)
    }, concurrency)

    fmt.Println(responses)
    fmt.Println(err)
    // map[9bZkp7q19f0:0 JGwWNGJdvx8:0 RgKAFK5djSk:6211818831 ThisIsError:0 XqZsoesa55w:14277740491 kJQP7kiw5Fk:8404577810]
    // can't parse "ThisIsError" views
}
```
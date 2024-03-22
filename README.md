# async

Helper for asynchronous work in golang powered by generics

* `AsyncToArray()`
* `AsyncToMap()`
* `NewPipers()`
* `NewPip()`



Installing
----------

	go get github.com/kozhurkin/async

Usage
-----

#### AsyncToArray()
``` golang
urls := []string{
    "https://www.youtube.com",
    "https://www.instagram.com",
    "https://www.tiktok.com",
}
ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5 * time.Second))
defer cancel()

// concurrency = 2 means that no more than 2 tasks can be performed at a time
concurrency := 2
responses, err := async.AsyncToArray(ctx, urls, func(i int, url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    return string(body), nil
}, concurrency)

if err != nil {
    log.Fatalln(err)
}

youtube, instagram, tiktok := responses[0], responses[1], responses[2]
```
#### AsyncToMap()

``` golang
videos := []string{
    "XqZsoesa55w",
    "kJQP7kiw5Fk",
    "RgKAFK5djSk",
}
ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
defer cancel()

// concurrency = 0 means that all tasks will be executed at the same time
concurrency := 0
responses, err := async.AsyncToMap(ctx, videos, func(i int, vid string) (int, error) {
    views, err := youtube.GetViews(vid)
    if err != nil {
        return 0, err
    }
    return views, nil
}, concurrency)

if err != nil {
    log.Fatalln(err)
}

fmt.Println(responses)
// map[XqZsoesa55w:11e9 kJQP7kiw5Fk:8e9 RgKAFK5djSk:5.7e9]
```
#### NewPipers()
``` golang
import github.com/kozhurkin/async/pipers

func main() {

    pp := pipers.NewPipers(
        func() (interface{}, error) { return loadAds() },
        func() (interface{}, error) { return loadUser(userId) },
        func() (interface{}, error) { return loadSite(siteId) },
    )

    err := pp.Run().FirstError()

    if err != nil {
        panic(err)
    }

    res := pp.Results()

    ads := res[1].([]Ad)
    user := res[2].(User)
    site := res[3].(Site)
}
```

``` golang
import github.com/kozhurkin/async/pipers

// usage of helper .Ref() and method .Resolve()
func main() {

    var ads []*Ad
    var user *User
    var site *Site

    pp := pipers.NewPipers(
        pipers.Ref(&ads, func() ([]*Ad, error) { return loadAds() }),
        pipers.Ref(&user, func() (*User, error) { return loadUser(userId) }),
        pipers.Ref(&site, func() (*Site, error) { return loadSite(siteId) }),
    )

    results, _ := pp.Run().Resolve()

    fmt.Printf("results: %T\t%v\n", results, results)   // results: []interface {} [[0xc000224010] 0xc0002a00f0 0xc000224020]
    fmt.Printf("ads:     %T\t%v\n", ads, ads)           // ads:     []*tests.Ad    [0xc000224010]
    fmt.Printf("user:    %T\t%v\n", user, user)         // user:    *tests.User    &{Dima}
    fmt.Printf("site:    %T\t%v\n", site, site)         // site:    *tests.Site    &{site.com}
}
```

#### NewPip()
``` golang
import github.com/kozhurkin/async/pip

func main() {
    ts := time.Now()
    pa := pip.NewPip(func() int {
        <-time.After(2 * time.Second) // working 2 seconds
        return 2024
    })
    pb := pip.NewPip(func() string {
        <-time.After(3 * time.Second) // working 3 seconds
        return "Happy New Year!"
    })
    a, b := <-pa, <-pb // parallel execution
    fmt.Println(time.Now().Sub(ts)) // execution time 3 seconds (not 5)
}
```

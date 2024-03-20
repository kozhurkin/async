# async

Helper for asynchronous work in golang powered by generics

* `AsyncToArray()`
* `AsyncToMap()`
* `NewPipers()`
* `Pip()`



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
ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5 * time.Second))
defer cancel()

// concurrency = 0 means that all tasks will be executed at the same time
concurrency := 0
responses, err := async.AsyncToArray(ctx, videos, func(i int, vid string) (int, error) {
    views, err := youtube.GetViews(vid)
    if err != nil {
        return nil, err
    }
    return views, nil
}, 0) 

if err != nil {
    log.Fatalln(err)
}

fmt.Println(responses)
// map[XqZsoesa55w:11e9 kJQP7kiw5Fk:8e9 RgKAFK5djSk:5.7e9]
```
#### NewPipers()
```golang
pp := async.NewPipers(
    func() (interface{}, error) { return loadAds() },
    func() (interface{}, error) { return loadUser(userId) },
    func() (interface{}, error) { return loadPlatform(platformId) }, 
)

err := pp.Run().FirstError()

if err != nil {
    panic(err)
}

results := pp.Results()

fmt.Println(err, results)
// <nil> [[] 0xc000180008 0x10247db80]

ads := res.Shift().([]Ad)
user := res.Shift().(User)
platform := res.Shift().(Platform)

fmt.Println(ads, user, platform)
// [] 0xc000180008 0x10247db80
```

```golang
// usage of helper .Ref() and method .Resolve() 

var ads []*Ad
var user *User
var platform *Platform

pp := async.NewPipers(
    async.Ref(&ads, func() ([]*Ad, error) {
        return loadAds()
    }),
    async.Ref(&user, func() (*User, error) {
        return loadUser(userId)
    }),
    async.Ref(&platform, func() (*Platform, error) {
        return loadPlatform(platformId)
    }),
)

res, err := pp.Run().Resolve()

fmt.Println(res, err)
// [[0xc00005e040] 0xc000114008 0xc000096740] <nil>

fmt.Println(ads, user, platform)
// [0xc00005e040] &{11051991} &{site.com}

```

#### Pip()
``` golang
ts := time.Now()
pa := async.Pip(func() int {
    <-time.After(2 * time.Second) // working 2 seconds
    return 2024
})
pb := async.Pip(func() string {
    <-time.After(3 * time.Second) // working 3 seconds
    return "Happy New Year!"
})
a, b := <-pa, <-pb // parallel execution
fmt.Println(time.Now().Sub(ts)) // execution time 3 seconds (not 5)
```

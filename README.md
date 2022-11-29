# async

Helper for asynchronous work in golang powered by generics

* `AsyncToArray()`
* `AsyncToMap()`
* `Pipeline()`



Installing
----------

	go get github.com/kozhurkin/async

Usage
-----

#### Pipeline()
``` golang
    ts := time.Now()
    pa := async.Pipeline(func() int {
        <-time.After(2 * time.Second) // working 2 seconds
        return 2023
    })
    pb := async.Pipeline(func() string {
        <-time.After(3 * time.Second) // working 3 seconds
        return "Happy New Year!"
    })
    a, b := <-pa, <-pb // parallel execution
    fmt.Println(time.Now().Sub(ts)) // execution time 3.0 seconds
```
#### AsyncToArray()
``` golang
    urls := []string{
        "https://www.youtube.com",
        "https://www.instagram.com",
        "https://www.tiktok.com",
    }
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5 * time.Second))
    defer cancel()
    
    concurrency := 2
    responses, err := async.AsyncToArray(ctx, urls, func(url string) (string, error) {
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
    
    responses, err := async.AsyncToArray(ctx, videos, func(vid string) (int, error) {
        views, err := youtube.GetViews(vid)
        if err != nil {
            return nil, err
        }
        return views, nil
    }, 0) // all requests at the same time
    
    if err != nil {
        log.Fatalln(err)
    }
    
    fmt.Println(responses)
    // map[XqZsoesa55w:11e9 kJQP7kiw5Fk:8e9 RgKAFK5djSk:5.7e9]

```

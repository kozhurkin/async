# async

helper for asynchronous work in golang powered by generics.

* `AsyncToArray()`
* `AsyncToMap()`

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
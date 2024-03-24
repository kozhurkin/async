package tests

import (
	"context"
	"fmt"
	"github.com/kozhurkin/pipers"
	"math/rand"
	"testing"
	"time"
)

func TestPipersMain(t *testing.T) {
	ts := time.Now()
	ps := pipers.FromFuncs(
		func() (int, error) { <-time.After(1000 * time.Millisecond); return 1, nil },
		func() (int, error) { <-time.After(2000 * time.Millisecond); return 2, nil },
		func() (int, error) { <-time.After(1500 * time.Millisecond); return 3, nil },
		func() (int, error) { <-time.After(2500 * time.Millisecond); return 4, nil },
		func() (int, error) { <-time.After(3000 * time.Millisecond); return 5, nil },
	)

	ctx := context.Background()
	timeout := time.Duration(rand.Intn(500)) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	fmt.Println(timeout)

	ps.Concurrency(2).Context(ctx)

	err, res := ps.FirstError(), ps.Results()

	fmt.Println(res, err, time.Since(ts))

	<-time.After(3 * time.Second)

	return
}

type Ad struct {
	Name string
}
type Site struct {
	Url string
}
type User struct {
	Name string
}

func loadAds() ([]*Ad, error) {
	return []*Ad{{"Aviasales"}}, nil
}
func loadUser(id int) (*User, error) {
	return &User{"Dima"}, nil
}
func loadSite(id int) (*Site, error) {
	return &Site{"site.com"}, nil
}

var userId = 1
var siteId = 1

func TestPipersReadme(t *testing.T) {

	var ads []*Ad
	var user *User
	var site *Site

	ps := pipers.FromFuncs(
		pipers.Ref(&ads, func() ([]*Ad, error) { return loadAds() }),
		pipers.Ref(&user, func() (*User, error) { return loadUser(userId) }),
		pipers.Ref(&site, func() (*Site, error) { return loadSite(siteId) }),
	)

	results, _ := ps.Resolve()

	fmt.Printf("results: %T\t%v\n", results, results)
	fmt.Printf("ads:     %T\t%v\n", ads, ads)
	fmt.Printf("user:    %T\t%v\n", user, user)
	fmt.Printf("site:    %T\t%v\n", site, site)

	<-time.After(time.Second)
}

package tests

import (
	"context"
	"errors"
	"fmt"
	"github.com/kozhurkin/async/pipers"
	"math/rand"
	"testing"
	"time"
)

func TestPipers(t *testing.T) {
	ts := time.Now()
	pp := pipers.NewPipers(
		func() (int, error) { <-time.After(10 * time.Millisecond); return 1, nil },
		func() (int, error) { <-time.After(20 * time.Millisecond); return 2, errors.New("surprise") },
		func() (int, error) { <-time.After(15 * time.Millisecond); return 3, nil },
		func() (int, error) { <-time.After(25 * time.Millisecond); return 4, errors.New("surprise 2") },
		func() (int, error) { <-time.After(30 * time.Millisecond); return 5, nil },
	)

	ctx := context.Background()
	timeout := time.Duration(rand.Intn(99)) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err, res := pp.Run().FirstErrorContext(ctx), pp.Results()

	fmt.Println(timeout, res, err, time.Since(ts))

	<-time.After(time.Second)

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
func loadPlatform(id int) (*Site, error) {
	return &Site{"site.com"}, nil
}
func TestPipersReadme(t *testing.T) {

	var userId = 1
	var platformId = 1

	var ads []*Ad
	var user *User
	var site *Site

	pp := pipers.NewPipers(
		pipers.Ref(&ads, func() ([]*Ad, error) {
			return loadAds()
		}),
		pipers.Ref(&user, func() (*User, error) {
			return loadUser(userId)
		}),
		pipers.Ref(&site, func() (*Site, error) {
			return loadPlatform(platformId)
		}),
	)

	results, _ := pp.Run().Resolve()

	fmt.Printf("results: %T\t%v\n", results, results)
	fmt.Printf("ads:     %T\t%v\n", ads, ads)
	fmt.Printf("user:    %T\t%v\n", user, user)
	fmt.Printf("site:    %T\t%v\n", site, site)

	<-time.After(time.Second)

	return
}

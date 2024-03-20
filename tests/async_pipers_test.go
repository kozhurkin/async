package tests

import (
	"context"
	"errors"
	"fmt"
	"github.com/kozhurkin/async"
	"math/rand"
	"testing"
	"time"
)

func TestPipers(t *testing.T) {
	ts := time.Now()
	pp := async.NewPipers(
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
type Platform struct {
	Url string
}
type User struct {
	Id int
}

func loadAds() ([]*Ad, error) {
	return []*Ad{{"aviasales"}, {"skyeng"}}, nil
}
func loadUser(id int) (*User, error) {
	return &User{11051991}, nil
}
func loadPlatform(id int) (*Platform, error) {
	return &Platform{"https://logr.info"}, nil
}
func TestPipersReadme(t *testing.T) {

	var userId = 1
	var platformId = 1

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
	fmt.Println(ads, user, platform)

	//ads := res.Shift().([]*Ad)
	//user := res.Shift().(*User)
	//platform := res.Shift().(Platform)

	//ads, user, platform := res.Shift().([]*Ad), res.Shift().(*User), res.Shift().(*Platform)

	//fmt.Println(ads, platform, user, user.Id)

	//fmt.Println(results.Get())

	<-time.After(time.Second)

	return
}

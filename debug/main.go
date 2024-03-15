package main

import (
	"context"
	"fmt"
	"github.com/kozhurkin/async"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {

	wp := GetWorkerPool(5)

	for i := 1; i <= 20; i++ {
		job := Job{ID: i}
		wp.AddJob(job)
	}

	wp.Wait()
	return
	//
	//t := time.Now()
	//fmt.Println("start")
	//
	//p2 := async.Promise(func() (int, error) {
	//	<-time.After(2 * time.Second)
	//	return 2, nil
	//})
	//p2.Start()
	//
	//p3 := async.Promise(func() (int, error) {
	//	<-time.After(3 * time.Second)
	//	return 3, nil
	//})
	//p3.Start()
	//
	//r2, r3 := <-p2.Out, <-p3.Out
	//
	//fmt.Println("___", r2, r3, time.Now().Sub(t).Seconds())

	//fmt.Println("go")
	//wg := new(errgroup.Group)
	//
	//wg.Go(func() (err error) {
	//	<-time.After(5 * time.Second)
	//	return errors.New("test")
	//})
	//wg.Go(func() (err error) {
	//	return errors.New("test")
	//})
	//err := wg.Wait()
	//fmt.Println(err)
	//return
	//fmt.Println(time.Now())
	//pa := async.Pipeline(func() time.Time { return <-time.After(2 * time.Second) })
	//pb := async.Pipeline(func() time.Time { return <-time.After(3 * time.Second) })
	//a, b := <-pa, <-pb
	//fmt.Println(time.Now(), []time.Time{a, b})
	//a2, b2 := <-pa, <-pb
	//fmt.Println(time.Now(), []time.Time{a2, b2})
	//return
	rand.Seed(time.Now().UnixNano())
	async.SetDebug(1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// handle SIGINT (control+c)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		select {
		case <-c:
			fmt.Println("\nmain: interrupt received. cancelling context.")
			cancel()
		}
	}()

	concurrency := 3
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
OUT:
	for i := 1; i <= 1; i++ {
		res, err := async.AsyncSemaphore(ctx, data, func(i int, k int) (int, error) {
			rnd := rand.Intn(1000)
			<-time.After(time.Duration(rnd) * time.Millisecond)
			if rand.Intn(len(data)) == 0 {
				return k, fmt.Errorf("unknown error (%v)", k)
			}
			return k, nil
		}, concurrency)
		fmt.Printf("[%v] RESULT: %v %v\n", i, res, err)
		select {
		case <-ctx.Done():
			fmt.Println("BREAAAAAK")
			break OUT
		default:
			continue
		}
	}
	<-time.After(1 * time.Second)
}

var once sync.Once

func GetWorkerPool(maxWorkers int) (singleton *WorkerPool) {
	once.Do(func() {
		workers := make([]*Worker, maxWorkers)
		workerPool := make(chan chan Job, maxWorkers)
		jobQueue := make(chan Job, maxWorkers)

		pool := &WorkerPool{
			workers:    workers,
			jobQueue:   jobQueue,
			workerPool: workerPool,
		}

		for i := 0; i < maxWorkers; i++ {
			worker := NewWorker(pool.workerPool, i)
			worker.Start(&pool.wg)
			workers[i] = worker
		}

		go pool.dispatch()

		singleton = pool
	})

	return singleton
}

func (wp *WorkerPool) dispatch() {
	for job := range wp.jobQueue {
		fmt.Println("-- job =", job)
		fmt.Println("--- wp.workerPool =", wp.workerPool)
		workerJobQueue := <-wp.workerPool
		fmt.Println("---- workerJobQueue =", workerJobQueue)
		workerJobQueue <- job
	}
}

func (wp *WorkerPool) AddJob(job Job) {
	wp.wg.Add(1)
	wp.jobQueue <- job
}

func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

type Job struct {
	ID int
}

type Worker struct {
	ID         int
	WorkerPool chan chan Job
	JobChannel chan Job
	quit       chan bool
}

func NewWorker(workerPool chan chan Job, id int) *Worker {
	return &Worker{
		ID:         id,
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
	}
}

func (w *Worker) Start(wg *sync.WaitGroup) {
	go func() {
		for {
			w.WorkerPool <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				fmt.Printf("Worker %d recv job %d\n", w.ID, job.ID)
				wg.Done()
			case <-w.quit:
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

type WorkerPool struct {
	workers    []*Worker
	jobQueue   chan Job
	workerPool chan chan Job
	wg         sync.WaitGroup
}

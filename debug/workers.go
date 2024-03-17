package main

import (
	"fmt"
	"sync"
)

func main() {

	wp := GetWorkerPool(5)

	for i := 1; i <= 20; i++ {
		job := Job{ID: i}
		wp.AddJob(job)
	}

	wp.Wait()
	return
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

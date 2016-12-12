package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pborman/uuid"
)

// payload type(job type)
type PayloadType string

type fnType func(payload PayloadType)

// Job
type Job struct {
	ID      string
	fn      func(payload PayloadType)
	Payload PayloadType
}

type Worker struct {
	ID       int32
	jobQueue chan Job
	idle     bool
	die      bool
	wait     bool
	wg       *sync.WaitGroup
}

func (w *Worker) Run() {
	go func() {
		for {
			if w.die {
				return
			}
			job := <-w.jobQueue
			w.idle = false
			job.fn(job.Payload)
			w.idle = true
			if w.wait {
				w.wg.Done()
			}
		}
	}()
}

type GoroutinePool struct {
	maxWorkers  int32
	doneJobs    int32
	jobQueue    chan Job
	jobQueueLen int32
	workers     []*Worker
	wg          *sync.WaitGroup
	wait        bool
	feedback    bool
}

// new GoroutinePool
func NewGoroutinePool(maxWorkers int32, jobQueueLen int32, feedback bool) *GoroutinePool {
	jobQueue := make(chan Job, jobQueueLen)
	workers := make([]*Worker, jobQueueLen)
	var wg sync.WaitGroup

	pool := &GoroutinePool{
		jobQueueLen: jobQueueLen,
		jobQueue:    jobQueue,
		wait:        true,
		feedback:    feedback,
		wg:          &wg,
		workers:     workers,
		doneJobs:    0,
		maxWorkers:  maxWorkers,
	}
	// start worker
	pool.Start()
	// start monitor
	pool.Monitor()

	return pool
}

func (pool *GoroutinePool) AddJob(fn func(payload PayloadType), payload PayloadType) {
	ID := uuid.NewUUID()
	job := Job{
		ID:      ID.String(),
		fn:      fn,
		Payload: payload,
	}
	if pool.wait {
		pool.wg.Add(1)
	}
	pool.doneJobs++
	pool.jobQueue <- job
}

func (pool *GoroutinePool) Monitor() {

	// work speed
	var lastDone int32
	lastDone = pool.jobQueueLen
	var speed int32
	interval := time.NewTicker(10 * time.Second)

	// real-time speed
	var realtimeLastDone int32
	realtimeLastDone = pool.jobQueueLen
	var realtimeSpeed int32

	// max average speed
	var maxAverageSpeed int32
	maxAverageSpeed = 0

	// interval monintor
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})

	// time cost
	tStart := time.Now()
	var costDuration time.Duration
	go func() {
		for {
			select {
			case <-ticker.C:
				tCurrent := time.Now()
				costDuration = tCurrent.Sub(tStart)

				// real-time speed
				tmpRealtimeSpeed := (pool.doneJobs - realtimeLastDone)

				realtimeSpeed = tmpRealtimeSpeed
				realtimeLastDone = pool.doneJobs

				fmt.Printf("\r Start at: %s, time cost: %s, average speed: %d, read-time speed: %d, current workers: %d, done jobs: %d     ", tStart.Format("15:04:05.000"), costDuration.String(), speed, realtimeSpeed, pool.maxWorkers, pool.doneJobs)

			case <-interval.C:
				// feedback mechanism!
				// average speed
				tmpSpeed := (pool.doneJobs - lastDone) / 10

				if tmpSpeed > maxAverageSpeed {
					maxAverageSpeed = tmpSpeed
				}

				if pool.feedback {
					var feedbackMaxWorkers int32
					feedbackMaxWorkers = pool.maxWorkers + tmpSpeed

					if speed > tmpSpeed {
						feedbackMaxWorkers = pool.maxWorkers - speed/2
					}

					if feedbackMaxWorkers < maxAverageSpeed {
						feedbackMaxWorkers = maxAverageSpeed
					}
					if feedbackMaxWorkers > pool.jobQueueLen {
						feedbackMaxWorkers = pool.jobQueueLen
					}
					// feedback worker numbers

					fmt.Printf("\r Start at: %s, time cost: %s, average speed: %d, read-time speed: %d, current workers: %d, done jobs: %d     ", tStart.Format("15:04:05.000"), costDuration.String(), speed, realtimeSpeed, pool.maxWorkers, pool.doneJobs)
					pool.feedbackWorkers(feedbackMaxWorkers)
				}

				speed = tmpSpeed
				lastDone = pool.doneJobs

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (pool *GoroutinePool) MapRun(fn func(payload PayloadType), payloads []PayloadType) {
	// maybe reflect
	// http://blog.burntsushi.net/type-parametric-functions-golang/
	for _, payload := range payloads {
		pool.AddJob(fn, payload)
	}
}

func (pool *GoroutinePool) MapRunChan(fn func(payload PayloadType), fnfetch func() PayloadType) {
	for {
		payload := fnfetch()
		pool.AddJob(fn, payload)
	}
}

func (pool *GoroutinePool) Start() {
	for i := int32(0); i < pool.maxWorkers; i++ {
		worker := &Worker{
			ID:       int32(i),
			jobQueue: pool.jobQueue,
			wait:     pool.wait,
			idle:     false,
			die:      false,
			wg:       pool.wg,
		}
		worker.Run()
		pool.workers[i] = worker
	}
}

func (pool *GoroutinePool) feedbackWorkers(feedbackMaxWorkers int32) {
	fmt.Println(feedbackMaxWorkers, pool.maxWorkers)
	if feedbackMaxWorkers > pool.maxWorkers {
		for i := pool.maxWorkers; i < feedbackMaxWorkers; i++ {
			worker := &Worker{
				ID:       int32(i),
				jobQueue: pool.jobQueue,
				wait:     pool.wait,
				idle:     false,
				die:      false,
				wg:       pool.wg,
			}
			worker.Run()
			pool.workers[i] = worker
		}
	} else {
		for i := feedbackMaxWorkers; i < pool.maxWorkers; i++ {
			pool.workers[i].die = true
		}
	}

	pool.maxWorkers = feedbackMaxWorkers
}

func (pool *GoroutinePool) Wait() {
	pool.wg.Wait()
}

func CustomLogger(fileName string) *log.Logger {
	// f, err := os.OpenFile(fileName, os.O_APPEND | os.O_CREATE | os.O_RDWR, 0666)
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}
	logger = log.New(f, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	return logger
}

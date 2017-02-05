package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pborman/uuid"
)

// default average speed interval
var SpeedInterval time.Duration = 10

// payload type(job type)
type PayloadType string

type fnType func(payload PayloadType)

// --------------------------------------- //

// Job
type Job struct {
	ID      string                    //job ID
	fn      func(payload PayloadType) //job function
	Payload PayloadType               //job payload
}

// Worker
type Worker struct {
	ID            int32    //job ID
	jobCacheQueue chan Job //job channl
	die           bool
	wait          bool //wait for done
	wg            *sync.WaitGroup
}

func (w *Worker) Run() {
	go func() {
		for {
			// shutdown
			if w.die {
				return
			}
			job := <-w.jobCacheQueue
			job.fn(job.Payload)
			// wait for done
			if w.wait {
				w.wg.Done()
			}
		}
	}()
}

type GoroutinePool struct {
	maxWorkers       int32 //max workers
	doneJobs         int32
	jobCacheQueue    chan Job
	jobCacheQueueLen int32
	workers          []*Worker
	wg               *sync.WaitGroup
	wait             bool
	feedback         bool
}

// new GoroutinePool
func NewGoroutinePool(maxWorkers int32, jobCacheQueueLen int32, feedback bool) *GoroutinePool {
	jobCacheQueue := make(chan Job, jobCacheQueueLen)
	workers := make([]*Worker, jobCacheQueueLen)

	if maxWorkers > jobCacheQueueLen {
		panic("maxWorkers must <= jobCacheQueueLen")
	}

	var wg sync.WaitGroup

	pool := &GoroutinePool{
		maxWorkers:       maxWorkers,
		jobCacheQueueLen: jobCacheQueueLen,
		jobCacheQueue:    jobCacheQueue,
		wait:             true,
		feedback:         feedback,
		wg:               &wg,
		workers:          workers,
		doneJobs:         0,
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
	// bad
	pool.doneJobs++
	pool.jobCacheQueue <- job
}

func (pool *GoroutinePool) Monitor() {
	var lastDone int32
	// bad
	lastDone = pool.jobCacheQueueLen
	var speed int32
	interval := time.NewTicker(SpeedInterval * time.Second)

	// real-time speed
	var realtimeLastDone int32
	// bad
	realtimeLastDone = pool.jobCacheQueueLen
	var realtimeSpeed int32

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

			// feedback mechanism!
			case <-interval.C:
				// average speed
				tmpSpeed := (pool.doneJobs - lastDone) / int32(SpeedInterval)

				if pool.feedback {
					var feedbackMaxWorkers int32
					feedbackMaxWorkers = pool.maxWorkers + (tmpSpeed-speed)*2

					if tmpSpeed < speed {
						feedbackMaxWorkers = pool.maxWorkers - (tmpSpeed - speed)
					}

					if feedbackMaxWorkers < pool.maxWorkers {
						feedbackMaxWorkers = pool.maxWorkers
					}
					if feedbackMaxWorkers > pool.jobCacheQueueLen {
						feedbackMaxWorkers = pool.jobCacheQueueLen
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
			ID:            int32(i),
			jobCacheQueue: pool.jobCacheQueue,
			wait:          pool.wait,
			die:           false,
			wg:            pool.wg,
		}
		worker.Run()
		pool.workers[i] = worker
	}
}

func (pool *GoroutinePool) feedbackWorkers(feedbackMaxWorkers int32) {
	if feedbackMaxWorkers > pool.maxWorkers {
		for i := pool.maxWorkers; i < feedbackMaxWorkers; i++ {
			worker := &Worker{
				ID:            int32(i),
				jobCacheQueue: pool.jobCacheQueue,
				wait:          pool.wait,
				die:           false,
				wg:            pool.wg,
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

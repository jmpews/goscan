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
	JobQueue chan Job
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
			job := <-w.JobQueue
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
	LoadJobNum  int32
	JobQueue    chan Job
	jobQueueLen int32
	workers     []*Worker
	wg          *sync.WaitGroup
	wait        bool
	feedback    bool
}

// new GoroutinePool
func NewGoroutinePool(initWorkers int32, jobQueueLen int32, feedback bool) *GoroutinePool {
	// if over `jobQueueLen`, will block
	JobChannel := make(chan Job, jobQueueLen)
	workers := make([]*Worker, jobQueueLen)
	var wg sync.WaitGroup

	pool := &GoroutinePool{
		jobQueueLen: jobQueueLen,
		JobQueue:    JobChannel,
		wait:        true,
		feedback:    feedback,
		wg:          &wg,
		workers:     workers,
		LoadJobNum:  0,
		maxWorkers:  initWorkers,
	}
	// start worker
	pool.Start()
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
	pool.LoadJobNum++
	pool.JobQueue <- job
}

func (pool *GoroutinePool) Monitor() {

	// work speed
	var lastDone int32
	lastDone = pool.jobQueueLen
	var speed int32
	interval := time.NewTicker(10 * time.Second)

	// real-time
	var realLastDone int32
	realLastDone = pool.jobQueueLen
	var realSpeed int32

	// max average speed
	var maxAverageSpeed int32
	maxAverageSpeed = 0

	// interval monintor
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})

	tStart := time.Now()
	go func() {
		for {
			select {
			case <-ticker.C:
				tCurrent := time.Now()
				costDuration := tCurrent.Sub(tStart)

				// real-speed
				realTmpSpeed := (pool.LoadJobNum - realLastDone)

				realSpeed = realTmpSpeed
				realLastDone = pool.LoadJobNum

				fmt.Printf("\r Start at: %s, Time over: %s, Worker Speed: %d, Real Speed: %d, Current Workers: %d, Load Job: %d", tStart.Format("15:04:05.000"), costDuration.String(), speed, realSpeed, pool.maxWorkers, pool.LoadJobNum)

			// feedback mechanism!
			case <-interval.C:
				// average speed
				tmpSpeed := (pool.LoadJobNum - lastDone) / 10

				if tmpSpeed > maxAverageSpeed {
					maxAverageSpeed = tmpSpeed
				}

				if pool.feedback {
					var reMaxWorkers int32
					reMaxWorkers = pool.maxWorkers + tmpSpeed

					if speed > tmpSpeed {
						reMaxWorkers = pool.maxWorkers - speed/2
					}

					if reMaxWorkers < maxAverageSpeed {
						reMaxWorkers = maxAverageSpeed
					}
					if reMaxWorkers > pool.jobQueueLen {
						reMaxWorkers = pool.jobQueueLen
					}
					// feedback and restart worker numbers
					pool.reStart(reMaxWorkers)
				}

				speed = tmpSpeed
				lastDone = pool.LoadJobNum

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
			JobQueue: pool.JobQueue,
			wait:     pool.wait,
			idle:     false,
			die:      false,
			wg:       pool.wg,
		}
		worker.Run()
		pool.workers[i] = worker
	}
}

func (pool *GoroutinePool) reStart(reMaxWorkers int32) {
	if reMaxWorkers > pool.jobQueueLen {
		return
	}
	if reMaxWorkers > pool.maxWorkers {
		for i := pool.maxWorkers; i < reMaxWorkers; i++ {
			worker := &Worker{
				ID:       int32(i),
				JobQueue: pool.JobQueue,
				wait:     pool.wait,
				idle:     false,
				die:      false,
				wg:       pool.wg,
			}
			worker.Run()
			pool.workers[i] = worker
		}
	} else {
		for i := reMaxWorkers; i < pool.maxWorkers; i++ {
			pool.workers[i].die = true
		}
	}

	pool.maxWorkers = reMaxWorkers
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

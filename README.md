# Golang Scan Framework

golang的扫描框架, 支持协程池和自动调节协程个数. **在30min内扫描391W的ULR(根据带宽和配置改变, 和Zmap不同, Zmap是无连接状态扫描)**

golang scanner framework, with goroutines pool and automatically adjusting the scanning speed.

**Scan 391W wordpress sites in 30min.**

---

## Features
* goroutines pool
* workers feedback mechanism
* monitor status

---

## Usage
before run, set the `maxWorkers`, `jobQueueLen` and `feedback mechanism`

```
// if you set a fixed number of goroutine, set feedback-mechanism `false` and initWorkers == jobQueueLen`
// Example: pool = NewGoroutinePool(1000, 1000, false)
// if you want feedback-mechanism, set `feedback = true`, initWorkers and jobQueueLen
// Example: pool := NewGoroutinePool(1000, 20000, true)

// 1000 initWorkers and 20000 jobQueueLen, with feedback mechanism
pool := NewGoroutinePool(1000, 20000, true)
```
config golang env

```
# build go env
> ./env.sh

```
start scanner

```
> go run scanner.go pool.go

# or
# > go build scanner.go pool.go
```

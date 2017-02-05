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

#### before run

* set `ulimit -n 100000`
* set [PayloadType type](https://github.com/jmpews/goscan/blob/master/pool.go#L16) (maybe i will add `relfect`)
* set [`maxWorkers`, `jobQueueLen` and `feedback mechanism`](https://github.com/jmpews/goscan/blob/master/scanner.go#L26)

```
// if you set a fixed number of goroutine, set feedback-mechanism `false` and maxWorkers == jobCacheQueueLen`
// Example: pool = NewGoroutinePool(100, 100, false)

// if you use feedback-mechanism, set `feedback = true`, maxWorkers and jobCacheQueueLen
// Example: pool := NewGoroutinePool(100, 1000, true)

// 1000 maxWorkers and 20000 jobCacheQueueLen, with feedback mechanism
pool := NewGoroutinePool(1000, 20000, true)
```

#### config golang env

```
sudo locale-gen zh_CN.UTF-8
sudo locale-gen zh_CN.UTF-8 en_US.UTF-8

wget https://storage.googleapis.com/golang/go1.7.3.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.7.3.linux-amd64.tar.gz

echo 'PATH=$PATH:/usr/local/go/bin' >> ~/.profile
echo 'GOPATH=/tmp' >> ~/.profile
. ~/.profile

go get github.com/pborman/uuid
```
#### start scanner

```
> go run scanner.go pool.go

# or
# > go build scanner.go pool.go
```

#### build in docker

```
$ git clone https://github.com/jmpews/goscan.git
$ cd goscan/
$ docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.7 go get github.com/pborman/uuid && go build -v scanner.go pool.go
```

#### cross compilation

```
GOOS=linux GOARCH=amd64 go build -v scanner.go pool.go
```

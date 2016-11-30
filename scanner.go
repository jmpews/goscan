package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var logger *log.Logger

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	logger = CustomLogger("run.log")

	// initWorkers, jobQueueLen(maxWorkers), feedback mechanism
	pool := NewGoroutinePool(1000, 20000, false)

	urlFile := "./wordpress.txt"
	fd, err := os.Open(urlFile)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		var domain string
		s := strings.Split(scanner.Text(), ",")
		domain, _ = s[0], s[1]
		pool.AddJob(fetchURL, PayloadType(domain))
	}
	pool.Wait()
}

func fetchURL(targetURL PayloadType) {
	// set timeout
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(25 * time.Second)
				// set timeout of connect
				c, err := net.DialTimeout(netw, addr, time.Second*20)
				if err != nil {
					logger.Println(err)
					return nil, err
				}
				// set timeout of send, write
				c.SetDeadline(deadline)
				return c, nil
			},
			// prevents re-use
			DisableKeepAlives: true,
		},
	}
	requestURL := "http://" + string(targetURL)
	// requestURL := "http://" + "baidu.com" + "/index.php"
	parseRequestURL, _ := url.Parse(requestURL)
	extraParams := url.Values{
		"spiderZz": {"Zz:0.6.1"},
	}
	parseRequestURL.RawQuery = extraParams.Encode()
	requestURL = parseRequestURL.String()
	req, err := http.NewRequest("GET", requestURL, nil)

	// set headers
	var Header map[string][]string
	Header = make(map[string][]string)
	Header["User-Agent"] = []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.98 Safari/537.36"}
	Header["Connection"] = []string{"keep-alive"}
	Header["Accept-Encoding"] = []string{"gzip, deflate"}
	Header["Accept"] = []string{"*/*"}
	Header["Accept-Encoding"] = []string{"gzip, deflate"}
	req.Header = Header

	// close indicates
	req.Close = true

	resp, err := client.Do(req)

	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		logger.Println(err)
		return
	}

	// save target-url to file
	if absolutePath, ok := checkVul(resp); ok {
		resultFile := "./result.txt"
		outFd, err := os.OpenFile(resultFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			outFd, err = os.Create(resultFile)
		}
		defer outFd.Close()
		outWriter := bufio.NewWriter(outFd)
		outWriter.WriteString(string(targetURL) + "," + absolutePath + "\n")
		outWriter.Flush()
		return
	}
	return
}

func checkVul(resp *http.Response) (string, bool) {
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Println(err)
		return "", false
	}
	r, _ := regexp.Compile(`vul_function\(\) in <b>(.+?)</b>`)

	regxResult := r.FindStringSubmatch(string(respData))
	if len(regxResult) == 2 {
		return regxResult[1], true
	}
	return "-", false
}

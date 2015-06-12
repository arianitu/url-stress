package main

import (
	"net/http"
	"time"
	"sync"
	"flag"
	"math"
	"fmt"
	"runtime"
	"os"
)

func Worker(url string, in <-chan int, sink chan<- int64) {
	for _ = range in {
		n := time.Now()
		resp, err := http.Get(url)
		r := time.Since(n)
		if err != nil {
			fmt.Println(err);
			os.Exit(1)
		}
		resp.Body.Close()
		sink <- r.Nanoseconds()
	}
	
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	requests := flag.Int("requests", 50, "The total number of requests to send out.")
	rps := flag.Int("rps", 0, "The number of requests per second. If this is set to 0, it will send as many as possible.")
	workers := flag.Int("workers", 0, "The number of workers. By default, it's set to number of CPUs.")
	url := flag.String("url", "", "The url to stress. Must have http/https in the url (required).")
	fout := flag.String("fout", "", "Path to file to print data in the format: request_number, latency(ms)\\n. This is useful to see how latency goes up over time on a graph")
	
	flag.Parse()
	if *url == "" {
		fmt.Println("URL is required! \n")
		flag.PrintDefaults()
		return
	}
	
	if *workers == 0 {
		*workers = runtime.NumCPU()
	}

	in := make(chan int)
	sink := make(chan int64)
	for i := 0; i < *workers; i++ {
		go Worker(*url, in, sink)
	}
	
	wg := &sync.WaitGroup{}
	wg.Add(*requests)

	var file *os.File
	var err error
	if (*fout != "") {
		file, err = os.Create(*fout)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	}
	
	var worst int64 = 0
	var best int64 = math.MaxInt64
	var totalTime int64 = 0
	go func() {
		count := 1
		for t := range sink {
			totalTime += t
			if t > worst {
				worst = t
			}
			if t < best {
				best = t
			}
			if file != nil {
				file.WriteString(fmt.Sprintf("%v,%v\n", count, t / int64(time.Millisecond)))
			}
			count++
			wg.Done()
		}
	}()
	now := time.Now()

	var sleep int
	if *rps > 0 {
		sleep = 1000 / *rps
	} else {
		sleep = 0
	}
	if *rps == 0 {
		fmt.Printf("Hitting URL %v with %v workers and %v requests as fast as I can \n", *url, *workers, *requests)
	} else {
		fmt.Printf("Hitting URL %v with %v workers, %v requests and %v rps \n", *url, *workers, *requests, *rps)
	}
	fmt.Println()
	
	for i := 0; i < *requests; i++ {
		in <- 1
		time.Sleep(time.Duration(sleep) * time.Millisecond)
	}
	wg.Wait()
	actualRps := float64(*requests) / time.Since(now).Seconds()
	avgDuration := time.Duration(totalTime/int64(*requests)) * time.Nanosecond
	bestDuration := time.Duration(best) * time.Nanosecond
	worstDuration := time.Duration(worst) * time.Nanosecond

	fmt.Printf("Rps: %.2f Avg: %v Worst: %v Best: %v \n", actualRps, avgDuration, worstDuration, bestDuration)

}

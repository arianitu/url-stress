package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Worker(url string, echo int, in <-chan int, sink chan<- int64) {
	for _ = range in {
		n := time.Now()
		resp, err := http.Get(url)
		r := time.Since(n)
		checkError(err)

		if echo == 1 {
			body, err := ioutil.ReadAll(resp.Body)
			checkError(err)

			fmt.Println(resp.Status)
			fmt.Printf("%s\n", body)
		}
		if resp.StatusCode != http.StatusOK {
			sink <- -1
			resp.Body.Close()
			continue
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
	echo := flag.Int("echo", 0, "Echo the body of the HTTP get response and the status code")

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
		go Worker(*url, *echo, in, sink)
	}

	wg := &sync.WaitGroup{}
	var file *os.File
	var err error
	if *fout != "" {
		file, err = os.Create(*fout)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	}

	var worst int64 = 0
	var best int64 = math.MaxInt64
	var totalTime int64 = 0
	errors := 0
	go func() {
		count := 1
		for t := range sink {
			if t < 0 {
				errors += 1
				wg.Done()
				continue
			}
			totalTime += t
			if t > worst {
				worst = t
			}
			if t < best {
				best = t
			}
			if file != nil {
				file.WriteString(fmt.Sprintf("%v,%v\n", count, t/int64(time.Millisecond)))
			}
			count++
			wg.Done()
		}
	}()
	now := time.Now()

	var sleep time.Duration
	if *rps > 0 {
		sleep = time.Duration(1e9/ *rps) * time.Nanosecond
	}
	
	if *rps == 0 {
		fmt.Printf("Hitting URL %v with %v workers and %v requests as fast as I can \n", *url, *workers, *requests)
	} else {
		fmt.Printf("Hitting URL %v with %v workers, %v requests and %v rps \n", *url, *workers, *requests, *rps)
	}
	fmt.Println()
	for i := 0; i < *requests; i++ {
		wg.Add(1)
		in <- 1
		if *rps > 0 {
			time.Sleep(sleep)
		}
	}
	wg.Wait()
	actualRps := float64(*requests-errors) / time.Since(now).Seconds()
	avgDuration := time.Duration(totalTime/int64(*requests)) * time.Nanosecond
	bestDuration := time.Duration(best) * time.Nanosecond
	worstDuration := time.Duration(worst) * time.Nanosecond

	fmt.Printf("Rps: %.2f Avg: %v Worst: %v Best: %v Errors: %.2f%%\n", actualRps, avgDuration, worstDuration, bestDuration, (float64(errors)/float64(*requests))*100.0)
	if errors > 0 {
		fmt.Println("You had errors in some of your requests, use the -echo option to find what those errors were. Requests with errors are not counted towards rps, average, worst and best.")
	}
}

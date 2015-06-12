# url-stress
Stress a URL by specifying the requests per second, the number of requests and how many to run in parallel.

Usage of ./url-stress:
```
  -fout="": Path to file to print data in the format: request_number, latency(ms)\n. 
  This is useful to see how latency goes up over time on a graph
  -requests=50: The total number of requests to send out.
  -rps=0: The number of requests per second. If this is set to 0, it will send as many as possible.
  -url="": The url to stress. Must have http/https in the url (required).
  -workers=0: The number of workers. By default, it's set to number of CPUs.
```

Example:

```./url-stress -url="http://www.google.ca" -requests=100 -rps=10```

Out:

```
Hitting URL http://www.google.ca with 8 workers, 100 requests and 10 rps 

Rps: 9.46 Avg: 131.805151ms Worst: 1.086656536s Best: 68.174066ms    
```

If you use the -fout="data.out" option, you can graph that data at: https://www.meta-chart.com/scatter-plot

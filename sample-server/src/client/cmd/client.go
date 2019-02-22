package main

import (
	"bytes"
	"common"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	timeInterval = flag.Duration("interval", 100, "timeinterval for send data to server")
	serverURL    = flag.String("url", "http://127.0.0.1:8081/request", "default url for server")
	threadCount  = flag.Int("count", 1, "default thread for access server")
	accessMethod = flag.String("method", "post", "default access method mode(post/get)")
)

type HttpClient struct {
	wg           *sync.WaitGroup
	stop         chan struct{}
	threadCount  int
	url          string
	timeInterval time.Duration
	method       string
}

func NewHttpClient(count int, timeInterval time.Duration, method, url string) *HttpClient {
	httpClient := &HttpClient{
		wg:           &sync.WaitGroup{},
		stop:         make(chan struct{}, count),
		url:          url,
		timeInterval: timeInterval,
		threadCount:  count,
		method:       strings.ToUpper(method),
	}
	fmt.Printf("...stop %d workers...\n", count)
	return httpClient
}

func (httpClient *HttpClient) initWorker(id int) {

	defer httpClient.wg.Done()
	fmt.Println("...start worker", id, "...")
	client := http.Client{}
	ticker := time.NewTicker(*timeInterval * time.Millisecond)
	var accessMethod string
	switch httpClient.method {
	case http.MethodGet:
		accessMethod = http.MethodGet
		break
	case http.MethodPost:
		accessMethod = http.MethodPost
		break
	}
	defer ticker.Stop()
	for {
		select {
		case <-httpClient.stop:
			fmt.Printf("...worker %d stoped\n", id)
			return
		case <-ticker.C:
			b, err := common.RequestToString(id)
			if err != nil {
				log.Error("RequestToString failed:", err)
				continue
			}
			reader := bytes.NewReader(b)
			if reader == nil {
				log.Error("NewReader failed:")
				continue
			}

			req, err := http.NewRequest(accessMethod, *serverURL, reader)
			if err != nil {
				log.Error("NewRequest failed:", err)
				continue
			}
			resp, err := client.Do(req)
			if err != nil {
				log.Error("client.Do failed:", err)
				return
			}
			req.Body.Close()
			b, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error("ReadAll failed:", err)
				return
			}
			resp.Body.Close()
			fmt.Printf("client got message:%s,status code:%s\n", string(b), resp.Status)
		}
	}
}
func (httpClient *HttpClient) Run() {
	httpClient.wg.Add(httpClient.threadCount)
	for i := 0; i < *threadCount; i++ {
		go httpClient.initWorker(i)
	}
}

func (httpClient *HttpClient) Close() {
	defer httpClient.wg.Wait()
	for i := 0; i < httpClient.threadCount; i++ {
		httpClient.stop <- struct{}{}
	}
	fmt.Printf("...stop %d worker success...", httpClient.threadCount)
}
func main() {
	flag.Parse()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	httpClient := NewHttpClient(*threadCount, *timeInterval, *accessMethod, *serverURL)
	httpClient.Run()
	defer httpClient.Close()
	for {
		select {
		case <-sig:
			fmt.Println("..got stop signal..")
			return
		}
	}
}

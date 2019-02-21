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
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	timeInterval = flag.Duration("interval", 100, "timeinterval for send data to server")
	serverURL    = flag.String("url", "http://127.0.0.1:8081/welcome", "default url for server")
	threadCount  = flag.Int("count", 1, "default thread for access server")
)

type HttpClient struct {
	wg           *sync.WaitGroup
	Stop         chan struct{}
	threadCount  int
	url          string
	timeInterval time.Duration
}

func NewHttpClient(count int, timeInterval time.Duration, url string) *HttpClient {
	httpClient := &HttpClient{
		wg:           &sync.WaitGroup{},
		Stop:         make(chan struct{}, count),
		url:          url,
		timeInterval: timeInterval,
	}
	httpClient.wg.Add(count)
	fmt.Printf("...stop %d workers...\n", count)
	return httpClient
}

func (httpClient *HttpClient) initWorker(id int) {

	defer httpClient.wg.Done()
	fmt.Println("...start worker", id, "...")
	client := http.Client{}
	ticker := time.NewTicker(*timeInterval * time.Millisecond)
	defer ticker.Stop()
MainLoop:
	for {
		select {
		case <-httpClient.Stop:
			log.Info("worker", id, " exited")
			break MainLoop
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
			req, err := http.NewRequest(http.MethodPost, *serverURL, reader)
			if err != nil {
				log.Error("NewRequest failed:", err)
				continue
			}
			resp, err := client.Do(req)
			if err != nil {
				log.Error("client.Do failed:", err)
				break MainLoop
			}
			b, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error("ReadAll failed:", err)
				break MainLoop
			}
			log.Infof("client got message:%s,status code:%s", string(b), resp.Status)
		}
	}
}
func (httpClient *HttpClient) Run() {
	for i := 0; i < *threadCount; i++ {
		go httpClient.initWorker(i)
	}
}
func (httpClient *HttpClient) Close() {
	defer httpClient.wg.Wait()
	for i := 0; i < httpClient.threadCount; i++ {
		httpClient.Stop <- struct{}{}
	}
	fmt.Printf("...stop %d worker success...", httpClient.threadCount)
}
func main() {
	flag.Parse()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	httpClient := NewHttpClient(*threadCount, *timeInterval, *serverURL)
	httpClient.Run()
	for {
		select {
		case <-sig:
			fmt.Println("..got stop signal..")
			httpClient.Close()
			return
		}
	}
}

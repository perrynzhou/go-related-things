package main

import (
	"bytes"
	"common"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
)

var (
	port      = flag.Int("port", 8081, "sample server port")
	urlPrefix = flag.String("prefix", "request", "default access prefix for url")
)

type HttpServer struct {
	port         int
	successCount uint64
	server       http.Server
	Stop         chan struct{}
	wg           *sync.WaitGroup
	writeBuffer  *bytes.Buffer
}

func NewHttpServer(port int, prefix string) *HttpServer {
	server := http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		ReadTimeout:    defaultReadTimeout,
		WriteTimeout:   defaultWriteTimeout,
		MaxHeaderBytes: 1 << 30,
	}
	httpServer := &HttpServer{
		port:         port,
		successCount: uint64(0),
		server:       server,
		Stop:         make(chan struct{}, 1),
		wg:           &sync.WaitGroup{},
		writeBuffer:  &bytes.Buffer{},
	}
	httpServer.wg.Add(1)
	return httpServer

}
func (httpServer *HttpServer) Run() {
	go func(server *http.Server, wg *sync.WaitGroup) {
		defer wg.Done()
		server.ListenAndServe()
	}(&httpServer.server, httpServer.wg)
}
func (httpServer *HttpServer) Close() {
	defer httpServer.wg.Wait()
	defer fmt.Printf("..stop httpserver...")
	for {
		select {
		case <-httpServer.Stop:
			if err := httpServer.server.Shutdown(context.Background()); err != nil {
				log.Printf("HTTP server Shutdown: %v", err)
			}
			httpServer.server.Close()
			return
		}
	}
}
func (httpServer *HttpServer) Do(w http.ResponseWriter, r *http.Request) {

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("ReadAll failed:", err)
	}
	req := common.Request{}
	err = json.Unmarshal(b, &req)
	if err != nil {
		log.Error("Unmarshal failed:", err)
		return
	}
	fmt.Printf("[server begin to  handle %s task]\n", common.ObtainClientInfo(r))
	curTime := time.Now().Format(common.TimeFmt)
	atomic.AddUint64(&httpServer.successCount, 1)
	httpServer.writeBuffer.Reset()
	log.Println("access method:",r.Method)
	switch r.Method {
	case http.MethodPost:
		httpServer.writeBuffer.WriteString(fmt.Sprintf("[finish worker %d task,time:%s]", req.Id, curTime))
		break
	case http.MethodGet:
		systemInfo := common.ObtainSystemInfo(r)
		b, err := json.Marshal(systemInfo)
		if err != nil {
			log.Error("Marshal failed:", err)
			return
		}
		httpServer.writeBuffer.WriteString(fmt.Sprintf("[finish worker %d get:\n%s\ntime:%s]", req.Id, string(b), curTime))
		break
	}
	_, err = w.Write(httpServer.writeBuffer.Bytes())
	if err != nil {
		log.Error("Write failed:", err)
		return
	}
	fmt.Printf("SuccessCount:%d,Recieved Request Info:Request{Id:%v,Time:%v,UUID:%v}  time:%s\n", httpServer.successCount, req.Id, req.Time, req.Uid, curTime)

}

func main() {
	flag.Parse()
	fmt.Printf("...start http server,server url:%s ...\n", fmt.Sprintf("127.0.0.1:%d/%s", *port, *urlPrefix))
	httpServer := NewHttpServer(*port, *urlPrefix)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("welcome to %s access this server...", common.ObtainClientInfo(r))
	})
	http.HandleFunc(fmt.Sprintf("/%s", *urlPrefix), httpServer.Do)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	httpServer.Run()
	for {
		select {
		case <-sig:
			fmt.Println("...recieve stop signal")
			httpServer.Stop <- struct{}{}
			httpServer.Close()
			return
		}
	}

}

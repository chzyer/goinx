package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"gopkg.in/logex.v1"
)

type Flag struct {
	Port         string
	Conf         string
	ReadTimeout  time.Duration // sec
	WriteTimeout time.Duration // sec
	Args         []string
}

func NewFlag() *Flag {
	f := &Flag{}
	flag.StringVar(&f.Port, "p", "80", "bind port")
	flag.StringVar(&f.Conf, "c", "router.conf", "router file path")
	var readTimeout, writeTimeout int
	flag.IntVar(&readTimeout, "rt", 10, "read timeout")
	flag.IntVar(&writeTimeout, "wt", 60, "write timeout")
	flag.Parse()
	f.ReadTimeout = time.Duration(readTimeout) * time.Second
	f.WriteTimeout = time.Duration(writeTimeout) * time.Second
	f.Args = flag.Args()
	return f
}

var (
	_flag  = NewFlag()
	router map[string]string
)

func main() {
	refreshConf()
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	mux.HandleFunc("/@reload", refreshHandler)
	mux.HandleFunc("/@router", routerHandler)
	mux.HandleFunc("/@goroutine", goroutineHandler)

	s := &http.Server{
		Addr:           ":" + _flag.Port,
		Handler:        mux,
		ReadTimeout:    _flag.ReadTimeout,
		WriteTimeout:   _flag.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	host := router[req.Host]
	if host == "" {
		http.NotFound(w, req)
		return
	}
	u, err := url.Parse("http://" + host)
	if err != nil {
		http.Error(w, "invalid host: "+host+"/"+err.Error(), 500)
		return
	}
	logex.Info(req.URL, u)
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ServeHTTP(w, req)
}

func refreshHandler(w http.ResponseWriter, req *http.Request) {
	refreshConf()
	w.Write([]byte("conf reloaded!\n"))
}

func routerHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(router[req.Host] + "\n"))
}

func goroutineHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%d", runtime.NumGoroutine())
}

func refreshConf() {
	router = readConf(_flag.Conf)
}

func readConf(p string) map[string]string {
	router := make(map[string]string)
	body, err := ioutil.ReadFile(p)
	if err != nil {
		return router
	}
	scanner := bufio.NewScanner(bytes.NewReader(body))
	for scanner.Scan() {
		sp := strings.Split(scanner.Text(), " ")
		router[sp[0]] = sp[1]
	}
	return router
}

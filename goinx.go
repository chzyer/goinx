package main

import (
	"bufio"
	"bytes"
	"flag"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type Flag struct {
	Port string
	Conf string
	Args []string
}

func NewFlag() *Flag {
	f := &Flag{}
	flag.StringVar(&f.Port, "p", "80", "bind port")
	flag.StringVar(&f.Conf, "c", "router.conf", "bind port")
	flag.Parse()
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
	mux.HandleFunc("/refresh", refreshHandler)

	if err := http.ListenAndServe(":"+_flag.Port, mux); err != nil {
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
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ServeHTTP(w, req)
}

func refreshHandler(w http.ResponseWriter, req *http.Request) {
	refreshConf()
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
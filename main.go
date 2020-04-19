package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
)

type config struct {
	StreamBase string
	Bind       string
}

func parseOptions() (config, error) {
	c := config{}
	flag.StringVar(&c.StreamBase, "stream-url", "", "The RTMP server to redirect to, excluding stream key")
	flag.StringVar(&c.Bind, "bind", "0.0.0.0:8080", "The address:port to bind to.")
	flag.Parse()

	if c.StreamBase == "" {
		return c, errors.New("you must specify a value for --stream-url")
	}
	return c, nil
}

type redirector struct {
	StreamBase string
}

func (rd *redirector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	u, err := url.Parse(rd.StreamBase)
	if err != nil {
		panic(err)
	}

	// For some reason we must give an IP address instead of a host.
	ips, err := net.LookupIP(u.Host)
	if err != nil || len(ips) == 0 {
		http.Error(w, fmt.Sprintf("couldn't resolve ip: %v", err), http.StatusInternalServerError)
		return
	}
	rand.Shuffle(len(ips), func(i, j int) { ips[i], ips[j] = ips[j], ips[i] })

	target := "rtmp://" + ips[0].String() + u.Path + "/" + name

	http.Redirect(w, r, target, http.StatusFound)
	log.Println(target)
}

func main() {
	c, err := parseOptions()
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	http.Handle("/stream", &redirector{StreamBase: c.StreamBase})
	log.Fatal(http.ListenAndServe(c.Bind, nil))
}

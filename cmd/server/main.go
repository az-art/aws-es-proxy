package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/az-art/aws-es-proxy/pkg/proxy"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var (
		verbose       bool
		prettify      bool
		logtofile     bool
		nosignreq     bool
		endpoint      string
		port          string
		listenAddress string
	)

	flag.StringVar(&endpoint, "e", "", "Amazon ElasticSearch Endpoint (e.g: https://dummy-host.eu-west-1.es.amazonaws.com)")
	flag.StringVar(&port, "p", "9200", "Amazon ElasticSearch port (e.g: 9200)")
	flag.StringVar(&listenAddress, "l", "0.0.0.0:"+port, "Local TCP port to listen on")
	flag.BoolVar(&verbose, "v", false, "Print user requests")
	flag.BoolVar(&logtofile, "logtofile", false, "Log user requests and ElasticSearch responses to files")
	flag.BoolVar(&prettify, "pretty", false, "Prettify verbose and file output")
	flag.BoolVar(&nosignreq, "nosign", false, "Disable AWS Signature v4")
	flag.Parse()

	if len(os.Args) < 3 {
		fmt.Println("You need to specify Amazon ElasticSearch endpoint.")
		fmt.Println("Please run with '-h' for a list of available arguments.")
		os.Exit(1)
	}

	p := proxy.New(
		endpoint,
		verbose,
		prettify,
		logtofile,
		nosignreq,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/", p.HandlerProxy)

	srv := &http.Server{
		Handler:      mux,
		Addr:         listenAddress,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Amazon ElasticSearch proxy listening on %s...\n", listenAddress)
		if logtofile {
			log.Printf("Writing logs to file \"enabled\"\n")
		}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	defer p.ShutDownProxy()
	waitForShutdown(srv)
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interruptChan
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Shutting down Amazon ElasticSearch proxy...")
	os.Exit(0)
}

package main

import (
	"context"
	"flag"
	"log"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:0", "listen address")
	browserMode := flag.Bool("browser", false, "open in the default browser")
	serveOnly := flag.Bool("serve-only", false, "run the local server without a window")
	debug := flag.Bool("debug", false, "enable webview developer mode")
	flag.Parse()

	store, err := NewStore()
	if err != nil {
		log.Fatal(err)
	}

	app := &Server{store: store}
	httpServer, url, err := startHTTPServer(*addr, app)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s running at %s", appName, url)
	if *browserMode {
		openURL(url)
		select {}
	}
	if *serveOnly {
		select {}
	}

	if err := runDesktop(url, *debug); err != nil {
		log.Fatal(err)
	}
	_ = httpServer.Shutdown(context.Background())
}

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// application version number
const version = "1.0.0"

// The configuration settings
type config struct {
	port int
	env  string //development, production, staging
}

// Dependency Injection
type application struct {
	config config
	logger *log.Logger
}

func main() {
	var cfg config
	//read in flags that are needed to populate config struct
	flag.IntVar(&cfg.port, "port", 4000, "API Server Port")
	flag.StringVar(&cfg.env, "env", "development", "Environment(development | staging | production)")
	flag.Parse()

	//Create a customized logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	//Create an instance of application struct
	app := &application{
		config: cfg,
		logger: logger,
	}

	//create our server ServeMux
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	//create a http server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	//start our server
	logger.Printf("Starting %s server on %s", cfg.env, srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}

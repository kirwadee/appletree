package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kirwadee/appletree/internal/data"
	"github.com/kirwadee/appletree/internal/jsonlog"
	_ "github.com/lib/pq"
)

// application version number
const (
	version = "1.0.0"
)

// The configuration settings
type config struct {
	port int
	env  string //development, production, staging
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// Dependency Injection
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	var cfg config
	//read in flags that are needed to populate config struct
	flag.IntVar(&cfg.port, "port", 4000, "API Server Port")
	flag.StringVar(&cfg.env, "env", "development", "Environment(development | staging | production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("APPLETREE_DB_DSN"), "Postgresql dsn")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "Postgresql max open conns")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Postgresql max idle conns")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "Postgresql max connection idle time")
	flag.Parse()

	//Create a customized logger instance
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	//create a connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, _ := time.ParseDuration(cfg.db.maxIdleTime)

	db.SetConnMaxIdleTime(duration)
	logger.PrintInfo("Connected to postgres db", nil)

	//Create an instance of application struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	//create our server ServeMux
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	//create a http server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		ErrorLog:     log.New(logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	//start our server
	logger.PrintInfo("Starting server", map[string]string{
		"addr": srv.Addr,
		"env":  cfg.env,
	})
	err = srv.ListenAndServe()
	logger.PrintFatal(err, nil)
}

// The openDB() returns pointer to *sql.DB connection pool
func openDB(cf config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cf.db.dsn)
	if err != nil {
		return nil, err
	}
	//create a context with 5 seconds timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//check if the connection to db is still alive
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

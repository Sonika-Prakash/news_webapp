package base

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"webapp/models"

	"log"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/upper/db/v4"
)

// Application ...
type Application struct {
	appName string
	server  Server
	debug   bool
	infoLog log.Logger
	errLog  log.Logger
	view    *jet.Set
	session *scs.SessionManager
	models  models.Models
}

// Server ...
type Server struct {
	host string
	port string
}

// GetApplicationInstance ...
func GetApplicationInstance(appName, host, port string, db *sql.DB, upperDB db.Session) *Application {
	jetSet := initJet()
	sess := initSession(appName, host, db)
	return &Application{
		appName: appName,
		server: Server{
			host: host,
			port: port,
		},
		debug:   true,
		infoLog: *log.New(os.Stdout, "INFO\t", log.Ltime|log.Ldate|log.Lshortfile),
		errLog:  *log.New(os.Stderr, "ERROR\t", log.Ltime|log.Ldate|log.Llongfile),
		view:    jetSet,
		session: sess,
		models:  models.NewModel(upperDB),
	}
}

// GetServer ...
func (a *Application) GetServer(h http.Handler) *http.Server {
	url := fmt.Sprintf("%s:%s", a.server.host, a.server.port)
	srv := http.Server{
		Handler: h,
		Addr:    url,
	}
	a.infoLog.Printf("Starting HTTP server on %s", url)
	return &srv
}

// CatchInterruptions ...
func (a *Application) CatchInterruptions(errs chan error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	a.errLog.Printf("caught %s, so application will attempt to gracefully shutdown", sig.String())
	errs <- fmt.Errorf("%s", sig)
}

// GracefulShutdown ...
func (a *Application) GracefulShutdown(srv *http.Server, e error) {
	a.infoLog.Println("Received an error on channel", e.Error())
	// exit gracefully
	if errHTTPServer := srv.Shutdown(context.Background()); errHTTPServer != nil {
		a.errLog.Println("failed to gracefully shutdown HTTP server", errHTTPServer.Error())
	}
	a.infoLog.Println("server shutdown complete, application will now exit")
}

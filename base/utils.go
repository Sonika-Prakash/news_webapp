package base

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
)

func initJet() *jet.Set {
	return jet.NewSet(
		jet.NewOSFileSystemLoader("./views"),
		jet.InDevelopmentMode(),
	)
}

func initSession(appName, host string, db *sql.DB) *scs.SessionManager {
	sess := scs.New()
	sess.Cookie.Domain = host
	sess.Cookie.Name = appName
	sess.Cookie.SameSite = http.SameSiteStrictMode
	sess.Store = postgresstore.New(db)
	return sess
}

// OpenDB ...
func OpenDB(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalln("Cannot connect to DB", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalln("Cannot connect to DB", err)
	}
	return db
}

func (a *Application) readIntFromURLQuery(r *http.Request, key string) int {
	val, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return 0
	}
	return val
}

func (a *Application) readIntDefault(r *http.Request, key string, defVal int) int {
	val := a.readIntFromURLQuery(r, key)
	if val <= 0 {
		return defVal
	}
	return val
}

func (a *Application) serverErr(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	a.errLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (a *Application) clientErr(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

package base

import (
	"log"
	"net/http"
	"os"

	"github.com/justinas/nosurf"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		debugLog := *log.New(os.Stdout, "DEBUG\t", log.Ltime|log.Ldate)
		debugLog.Println(r.Method, r.URL)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func (a *Application) loadSession(next http.Handler) http.Handler {
	return a.session.LoadAndSave(next)
}

func (a *Application) authRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := a.session.GetInt(r.Context(), sessionKeyUserID)
		if userID == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		w.Header().Add("Cache-Controle", "no-store")
		next.ServeHTTP(w, r)
	}
}

func (a *Application) csrfTokenRequired(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	return csrfHandler
}

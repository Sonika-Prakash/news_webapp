package base

import (
	"fmt"
	"net/http"

	"github.com/CloudyKit/jet/v6"
	"github.com/justinas/nosurf"
)

// TemplateData ...
type TemplateData struct {
	URL             string
	IsAuthenticated bool
	AuthUser        string
	Flash           string
	Success         string
	Error           string
	CSRFToken       string
}

func (a *Application) defaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.URL = fmt.Sprintf("http://%s:%s", a.server.host, a.server.port)

	if a.session != nil {
		if a.session.Exists(r.Context(), sessionKeyUserID) {
			// this means user is logged in
			td.IsAuthenticated = true
			td.AuthUser = a.session.GetString(r.Context(), sessionKeyUsername)
		}
		td.Flash = a.session.PopString(r.Context(), "flash")
		td.Success = a.session.PopString(r.Context(), "success")
	}
	td.CSRFToken = nosurf.Token(r)
	return td
}

func (a *Application) render(w http.ResponseWriter, r *http.Request, viewName string, vars jet.VarMap) error {
	td := &TemplateData{}
	td = a.defaultData(td, r)

	template, err := a.view.GetTemplate(fmt.Sprintf("%s.html", viewName))
	if err != nil {
		return err
	}
	if err := template.Execute(w, vars, td); err != nil {
		return err
	}

	return nil
}

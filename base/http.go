package base

import (
	"fmt"
	"net/http"
	"strconv"
	"webapp/forms"
	"webapp/models"
	"webapp/public"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/mux"
)

const (
	sessionKeyUserID   = "userID"
	sessionKeyUsername = "username"
)

// MakeHTTPHandler creates and returns the gin default router
func MakeHTTPHandler(app *Application) http.Handler {
	router := mux.NewRouter()
	if app.debug {
		router.Use(loggingMiddleware)
	}
	router.Use(app.csrfTokenRequired)
	router.Use(app.loadSession)

	// routes
	router.HandleFunc("/", app.homeHandler).Methods(http.MethodGet)
	router.HandleFunc("/comments/{postID}", app.commentHandler).Methods(http.MethodGet)
	router.HandleFunc("/login", app.loginHandler).Methods(http.MethodGet)
	router.HandleFunc("/login", app.loginPostHandler).Methods(http.MethodPost)
	router.HandleFunc("/signup", app.signupHandler).Methods(http.MethodGet)
	router.HandleFunc("/signup", app.signupPostHandler).Methods(http.MethodPost)
	router.HandleFunc("/logout", app.authRequired(app.logoutHandler)).Methods(http.MethodGet)
	router.HandleFunc("/vote", app.authRequired(app.voteHandler)).Methods(http.MethodGet)
	router.HandleFunc("/submit", app.authRequired(app.submitHandler)).Methods(http.MethodGet)
	router.HandleFunc("/submit", app.authRequired(app.submitPostHandler)).Methods(http.MethodPost)
	router.HandleFunc("/comments/{postID}", app.authRequired(app.commentPostHandler)).Methods(http.MethodPost)

	// exposing css and images via /public path which is referenced by html pages
	fileServer := http.FileServer(http.FS(public.Files))
	router.PathPrefix("/public/").Handler(http.StripPrefix("/public", fileServer))

	return router
}

func (a *Application) homeHandler(w http.ResponseWriter, r *http.Request) {
	//* dummy data for testing
	// dummyUser := models.Users{
	// 	Email:    "sonika@gmail.com",
	// 	Password: "password",
	// 	Username: "Sonika",
	// }
	// err := a.models.Users.Insert(&dummyUser)
	// a.models.Posts.Insert("Today's Headlines-1", "http://localhost:8080", dummyUser.ID)
	// a.models.Posts.Insert("Today's Headlines-2", "http://localhost:8080", dummyUser.ID)
	// a.models.Posts.Insert("Today's Headlines-3", "http://localhost:8080", dummyUser.ID)

	err := r.ParseForm()
	if err != nil {
		a.errLog.Println(err)
		a.serverErr(w, err)
		return
	}

	filter := models.Filters{
		Query:    r.URL.Query().Get("q"),
		Page:     a.readIntDefault(r, "page", 1),
		PageSize: a.readIntDefault(r, "page_size", 5),
		OrderBy:  r.URL.Query().Get("order_by"),
	}

	posts, meta, err := a.models.Posts.GetPosts(filter)
	if err != nil {
		a.errLog.Println(err)
		a.serverErr(w, err)
		return
	}

	queryURL := fmt.Sprintf("page_size=%d&order_by=%s&q=%s", meta.PageSize, filter.OrderBy, filter.Query)
	nextURL := fmt.Sprintf("%s&page=%d", queryURL, meta.NextPage)
	prevURL := fmt.Sprintf("%s&page=%d", queryURL, meta.PrevPage)

	vars := make(jet.VarMap)
	vars.Set("posts", posts)
	vars.Set("meta", meta)
	vars.Set("nextUrl", nextURL)
	vars.Set("prevUrl", prevURL)
	vars.Set("form", forms.New(r.Form))

	err = a.render(w, r, "index", vars)
	if err != nil {
		a.errLog.Println(err)
		a.serverErr(w, err)
		return
	}
}

func (a *Application) commentHandler(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)

	postID, err := strconv.Atoi(mux.Vars(r)["postID"])
	if err != nil {
		a.errLog.Println(err)
		a.clientErr(w, http.StatusBadRequest)
		return
	}

	post, err := a.models.Posts.GetByID(postID)
	if err != nil {
		a.errLog.Println(err)
		a.serverErr(w, err)
		return
	}

	comments, err := a.models.Comments.GetCommentsForPost(postID)
	if err != nil {
		a.errLog.Println(err)
		a.serverErr(w, err)
		return
	}

	vars.Set("post", post)
	vars.Set("comments", comments)

	err = a.render(w, r, "comments", vars)
	if err != nil {
		a.errLog.Println(err)
		a.serverErr(w, err)
		return
	}
}

func (a *Application) commentPostHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*2)

	err := r.ParseForm()
	if err != nil {
		a.errLog.Println(err)
		a.serverErr(w, err)
		return
	}

	posdtID, err := strconv.Atoi(mux.Vars(r)["postID"])
	if err != nil {
		a.errLog.Println(err)
		a.clientErr(w, http.StatusBadRequest)
		return
	}

	userID := a.session.GetInt(r.Context(), sessionKeyUserID)

	form := forms.New(r.PostForm)

	form.Required("comment").MaxLength("comment", 1000)
	if !form.Valid() {
		a.errLog.Println(form.Errors)
		a.session.Put(r.Context(), "flash", form.Errors.First("comment"))
		http.Redirect(w, r, fmt.Sprintf("/comments/%d", posdtID), http.StatusSeeOther)
		return
	}

	err = a.models.Comments.Insert(form.Get("comment"), posdtID, userID)
	if err != nil {
		a.errLog.Println(form.Errors)
		a.session.Put(r.Context(), "flash", "Error while commenting on the post")
		http.Redirect(w, r, fmt.Sprintf("/comments/%d", posdtID), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/comments/%d", posdtID), http.StatusSeeOther)

}

func (a *Application) loginHandler(w http.ResponseWriter, r *http.Request) {
	err := a.render(w, r, "login", nil)
	if err != nil {
		a.errLog.Println(err)
		a.serverErr(w, err)
	}
}

func (a *Application) signupHandler(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)
	vars.Set("form", forms.New(r.PostForm))
	err := a.render(w, r, "signup", vars)
	if err != nil {
		a.errLog.Println(err)
		a.serverErr(w, err)
	}
}

func (a *Application) loginPostHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*2)

	err := r.ParseForm()
	if err != nil {
		a.errLog.Println(err)
		a.serverErr(w, err)
		return
	}

	form := forms.New(r.PostForm)
	form.Email("email")
	form.MinLength("password", 3)
	form.MaxLength("password", 16)

	if !form.Valid() {
		// if there are form errors, render these errors in UI
		vars := make(jet.VarMap)
		vars.Set("errors", form.Errors)
		err := a.render(w, r, "login", vars)
		if err != nil {
			a.errLog.Println(err)
			a.serverErr(w, err)
			return
		}
	}

	// if no form errors, login the user
	user, err := a.models.Users.AuthenticateUser(form.Get("email"), form.Get("password"))
	if err != nil {
		a.session.Put(r.Context(), "flash", "Login error: "+err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	a.session.RenewToken(r.Context()) // create a fresh session for this newly logged in user
	a.session.Put(r.Context(), sessionKeyUserID, user.ID)
	a.session.Put(r.Context(), sessionKeyUsername, user.Username)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *Application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	a.session.Remove(r.Context(), sessionKeyUserID)
	a.session.Remove(r.Context(), sessionKeyUsername)
	a.session.Destroy(r.Context())
	a.session.RenewToken(r.Context())

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *Application) signupPostHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*2)

	err := r.ParseForm()
	if err != nil {
		a.errLog.Println(err)
		a.serverErr(w, err)
		return
	}

	form := forms.New(r.PostForm)
	vars := make(jet.VarMap)
	vars.Set("form", form)

	form.Required("name", "email", "password").Email("email")

	if !form.Valid() {
		vars.Set("errors", form.Errors)
		err := a.render(w, r, "signup", vars)
		if err != nil {
			a.errLog.Println(err)
			a.serverErr(w, err)
			return
		}
		return
	}

	// add a new user account
	user := models.Users{
		Username:  form.Get("name"),
		Password:  form.Get("password"),
		Email:     form.Get("email"),
		Activated: true, // TODO: static for now, can be made dynamic by activating account only after email confirmation
	}
	err = a.models.Users.Insert(&user)
	if err != nil {
		a.errLog.Println(err)
		form.Fail("signup", fmt.Sprintf("Failed to create a new user account for the user %s", form.Get("name")))
		vars.Set("errors", form.Errors)
		err := a.render(w, r, "signup", vars)
		if err != nil {
			a.serverErr(w, err)
		}
		return
	}

	a.session.Put(r.Context(), "success", "Account created successfully! You can login now.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *Application) voteHandler(w http.ResponseWriter, r *http.Request) {
	id := a.readIntDefault(r, "id", 0)

	post, err := a.models.Posts.GetByID(id)
	if err != nil {
		a.errLog.Println(err)
		a.session.Put(r.Context(), "flash", "Error while voting "+err.Error())
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// this gives the ID of user currently logged in
	userID := a.session.GetInt(r.Context(), sessionKeyUserID)
	err = a.models.Posts.AddVote(post.ID, userID)
	if err != nil {
		a.errLog.Println(err)
		a.session.Put(r.Context(), "flash", "Error while voting. "+err.Error()+".")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	flashMsg := a.session.GetString(r.Context(), "flash")
	if flashMsg == "" {
		a.session.Put(r.Context(), "success", "Voted successfully!")
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Get method is to just get the submit form for the user to enter title and url
func (a *Application) submitHandler(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)
	vars.Set("form", forms.New(r.PostForm))
	err := a.render(w, r, "submit", vars)
	if err != nil {
		a.errLog.Println(err)
		a.serverErr(w, err)
		return
	}
}

// Post method is to do the backend DB operation
func (a *Application) submitPostHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*2)

	err := r.ParseForm()
	if err != nil {
		a.errLog.Println(err)
		a.serverErr(w, err)
		return
	}

	form := forms.New(r.PostForm)
	userID := a.session.GetInt(r.Context(), sessionKeyUserID)
	vars := make(jet.VarMap)

	form.Required("title", "url").URL("url").MaxLength("title", 100).MaxLength("url", 255)
	vars.Set("form", form)
	if !form.Valid() {
		vars.Set("errors", form.Errors)
		err := a.render(w, r, "submit", vars)
		if err != nil {
			a.errLog.Println(err)
			a.serverErr(w, err)
			return
		}
	}

	_, err = a.models.Posts.Insert(form.Get("title"), form.Get("url"), userID)
	if err != nil {
		vars.Set("errors", form.Errors)
		err := a.render(w, r, "submit", vars)
		if err != nil {
			a.errLog.Println(err)
			a.serverErr(w, err)
			return
		}
	}

	a.session.Put(r.Context(), "success", "Post submitted successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

package service

import (
	"github.com/gorilla/mux"
	"net/http"
	"html/template"
	"log"
	"database/sql"
	"os"
	"github.com/tteige/simpleWorkschedule/models"
	"github.com/gorilla/sessions"
)

type TemplateInput struct {
	Authentication Authentication
	Users          Users
}

type Users struct {
	Users []models.Employee
}

type ErrorResponse struct {
	Error string
	Path  string
}

type Authentication struct {
	LoggedIn bool
	Admin    bool
}

const (
	AppName = "workscheduler"
)

type Server struct {
	templates   *template.Template
	DB          *sql.DB
	CookieStore *sessions.CookieStore
}

func (server *Server) Serve() {
	r := mux.NewRouter()
	r.HandleFunc("/", server.indexHandle).Methods(http.MethodGet)
	r.HandleFunc("/signup", server.signUpHandle).Methods(http.MethodPost, http.MethodGet)
	r.HandleFunc("/login", server.loginHandle).Methods(http.MethodPost)
	r.HandleFunc("/logout", server.logoutHandle).Methods(http.MethodPost)
	r.HandleFunc("/users", server.usersHandle).Methods(http.MethodGet)

	tmplLoc := "service/templates/"

	server.templates = template.Must(template.ParseFiles(tmplLoc+"footer.html", tmplLoc+"header.html",
		tmplLoc+"index.html", tmplLoc+"navbar.html", tmplLoc+"login.html", tmplLoc+"signup.html", tmplLoc+"users.html"))
	log.Print("Listening on localhost:8080")
	http.ListenAndServeTLS("localhost:8080", os.Getenv("SERVER_CERT"), os.Getenv("SERVER_KEY"), r)
}

func (server *Server) indexHandle(w http.ResponseWriter, r *http.Request) {

	tmplInput, err := server.generateTemplateInput(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = server.renderTemplate(w, "index", tmplInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (server *Server) usersHandle(w http.ResponseWriter, r *http.Request) {

	ok, err := server.verifyAccess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tmplInput, err := server.generateTemplateInput(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmplInput.Users.Users, err = models.GetAllEmployees(server.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = server.renderTemplate(w, "users", tmplInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) signUpHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmplInput, err := server.generateTemplateInput(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		err = server.renderTemplate(w, "signup", tmplInput)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else if r.Method == http.MethodPost {
		server.signUp(w, r)
	}
}

func (server *Server) loginHandle(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if email, ok := r.Form["email"]; ok {
		employee, err := models.GetEmployee(server.DB, email[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if pw, ok := r.Form["password"]; ok {
			matches, err := CheckPasswordHash(pw[0], employee.PasswordHash)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			if matches {
				session, err := server.CookieStore.Get(r, AppName)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				session.Values["authenticated"] = true
				session.Values["first_name"] = employee.FirstName
				session.Values["last_name"] = employee.LastName
				session.Values["admin"] = employee.Admin
				log.Print(session)
				session.Save(r, w)
			}
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func (server *Server) logoutHandle(w http.ResponseWriter, r *http.Request) {
	session, err := server.CookieStore.Get(r, AppName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if auth, ok := session.Values["authenticated"].(bool); ok || auth {
		session.Values["authenticated"] = false
	}
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (server *Server) signUp(w http.ResponseWriter, r *http.Request) {
	var employee models.Employee

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if email, ok := r.PostForm["email"]; ok {
		employee.Email = email[0]
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if firstName, ok := r.PostForm["first_name"]; ok {
		employee.FirstName = firstName[0]
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if lastName, ok := r.PostForm["last_name"]; ok {
		employee.LastName = lastName[0]
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if affiliation, ok := r.PostForm["affiliation"]; ok {
		employee.Affiliation = affiliation[0]
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if username, ok := r.PostForm["username"]; ok {
		employee.Username = username[0]
	}

	if pw, ok := r.PostForm["password"]; ok {
		hash, err := HashPassword(pw[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		employee.PasswordHash = hash
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = models.InsertEmployee(server.DB, employee)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (server *Server) generateTemplateInput(r *http.Request) (TemplateInput, error) {
	tmplInput := TemplateInput{
		Authentication: Authentication{
			LoggedIn: false,
			Admin:    false,
		},
	}
	session, err := server.CookieStore.Get(r, AppName)
	if err != nil {
		return TemplateInput{}, err
	}
	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		tmplInput.Authentication.LoggedIn = true
	}

	if admin, ok := session.Values["admin"].(bool); ok && admin {
		tmplInput.Authentication.Admin = true
	}

	log.Printf("Template input: %+v", tmplInput)

	return tmplInput, nil
}

func (server *Server) renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) error {
	err := server.templates.ExecuteTemplate(w, tmpl, p)
	if err != nil {
		return err
	}
	return nil
}

func (server *Server) verifyAccess(r *http.Request) (bool, error) {
	session, err := server.CookieStore.Get(r, AppName)
	if err != nil {
		return false, err
	}
	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		if auth, ok := session.Values["admin"].(bool); ok && auth {
			return true, nil
		}
	}
	return false, nil
}
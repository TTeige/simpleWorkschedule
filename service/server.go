package service

import (
	"github.com/gorilla/mux"
	"net/http"
	"html/template"
	"log"
	"database/sql"
	"os"
	"github.com/tteige/simpleWorkschedule/models"
	"golang.org/x/crypto/bcrypt"
	"github.com/gorilla/sessions"
)

type Server struct {
	templates   *template.Template
	DB          *sql.DB
	CookieStore *sessions.CookieStore
}

func (server *Server) Serve() {
	r := mux.NewRouter()
	r.HandleFunc("/", server.indexHandle).Methods("GET")
	r.HandleFunc("/signup", server.signUpHandle).Methods("POST")
	r.HandleFunc("/login", server.loginHandle).Methods("POST")

	tmplLoc := "service/templates/"

	server.templates = template.Must(template.ParseFiles(tmplLoc+"footer.html", tmplLoc+"header.html",
		tmplLoc+"index.html", tmplLoc+"navbar.html", tmplLoc+"login.html"))
	log.Print("Listening on localhost:8080")
	http.ListenAndServeTLS("localhost:8080", os.Getenv("SERVER_CERT"), os.Getenv("SERVER_KEY"), r)
}

func (server *Server) indexHandle(w http.ResponseWriter, r *http.Request) {
	err := server.renderTemplate(w, "index", []string{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) error {
	err := server.templates.ExecuteTemplate(w, tmpl, p)
	if err != nil {
		return err
	}
	return nil
}

func (server *Server) signUpHandle(w http.ResponseWriter, r *http.Request) {

	var employee models.Employee

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	if email, ok := r.PostForm["email"]; ok {
		employee.Email = email[0]
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	if firstName, ok := r.PostForm["first_name"]; ok {
		employee.FirstName = firstName[0]
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	if lastName, ok := r.PostForm["last_name"]; ok {
		employee.LastName = lastName[0]
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	if affiliation, ok := r.PostForm["affiliation"]; ok {
		employee.Affiliation = affiliation[0]
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	if username, ok := r.PostForm["username"]; ok {
		employee.Username = username[0]
	}

	if pw, ok := r.PostForm["password"]; ok {
		hash, err := HashPassword(pw[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		employee.PasswordHash = hash
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	err = models.InsertEmployee(server.DB, employee)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
				session, err := server.CookieStore.Get(r, employee.Email)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				session.Values["authenticated"] = true
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

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false, err
	}
	return true, nil
}

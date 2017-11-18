package main

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	uuid "github.com/satori/go.uuid"
)

type Server struct {
	db *DB
}

func NewServer(db *DB) (*Server, error) {

	return &Server{db: db}, nil
}

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	os.Open("index.html")

}
func (s *Server) FormHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		file, err := os.Open("form.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()
		if _, err := io.Copy(w, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	case "POST":
		user := User{
			ID:    uuid.NewV4().String(),
			Fname: r.FormValue("first_name"),
			Sname: r.FormValue("second_name"),
			Email: r.FormValue("email"),
			Phone: r.FormValue("phone"),
			About: r.FormValue("about"),
		}
		if err := s.db.AddUser(user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl := template.Must(template.New("").Parse(`
			<html>
			<h2> New user added!!</h2>
			<p><form method='get' action = "/" id = "ok"></form>
			<button form = "ok">OK</button></html>
			`))
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		return

	default:
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
}

func (s *Server) UsersHandler(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"inc": func(i int) int {
			return i + 1
		},
	}

	users, err := s.db.AllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		PageTitle string
		Users     []User
	}{"All Users: ", users}

	t, err := ioutil.ReadFile("users.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.New("").Funcs(funcMap).Parse(string(t)))
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
func main() {
	db, err := NewDB("./bolt.db")
	if err != nil {
		log.Fatal(err)
	}
	server, err := NewServer(db)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/form", server.FormHandler)
	http.HandleFunc("/users", server.UsersHandler)
	http.Handle("/", http.FileServer(http.Dir("./")))

	http.ListenAndServe(":8080", nil)
}

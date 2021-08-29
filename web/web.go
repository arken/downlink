package web

import (
	"database/sql"
	"net/http"
	"text/template"

	"github.com/arken/downlink/database"
	"github.com/arken/downlink/ipfs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Node struct {
	DB   *database.DB
	Node *ipfs.Node
}

func (n *Node) Start(addr string) {
	// Setup Chi Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Setup handler functions for api endpoints
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.Get("/*", n.handleMain)

	// Start http server and listen for incoming connections
	http.ListenAndServe(addr, r)
}

func (n *Node) handleMain(w http.ResponseWriter, r *http.Request) {
	node, err := n.DB.Get(r.URL.Path)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if node.Type == "file" {
		file, err := n.Node.Cat(node.CID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(file)
		return
	}

	children := []database.Node{}
	for i := 0; ; i++ {
		results, err := n.DB.GetChildren(node.Path, 50, i)
		if err != nil && err == sql.ErrNoRows {
			break
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		children = append(children, results...)
	}

	t := template.New("manifest.html")
	t, err = t.ParseFiles("web/templates/manifest.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, children)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

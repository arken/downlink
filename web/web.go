package web

import (
	"database/sql"
	"net/http"
	"text/template"

	"github.com/arken/downlink/database"
	"github.com/arken/downlink/ipfs"
	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Node struct {
	DB   *database.DB
	Node *ipfs.Node
}

func (n *Node) Start(addr string, forceHTTPS bool) {
	// Setup Chi Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	if forceHTTPS {
		r.Use(middleware.RouteHeaders().Route(
			"X-Forwarded-Proto",
			"http",
			upgradeToHTTPS,
		).Handler)
	}

	// Setup handler functions for api endpoints
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.Get("/*", n.handleMain)

	// Start http server and listen for incoming connections
	http.ListenAndServe(addr, r)
}

func upgradeToHTTPS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
	})
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
		err := n.Node.Cat(node.CID, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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

	t := template.New("manifest.html").Funcs(template.FuncMap{
		"hsize": func(input int64) string {
			return humanize.Bytes(uint64(input))
		},
	})
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

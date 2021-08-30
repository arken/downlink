package web

import (
	"database/sql"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/arken/downlink/database"
	"github.com/arken/downlink/ipfs"
	"github.com/dustin/go-humanize"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	files "github.com/ipfs/go-ipfs-files"
)

type Node struct {
	DB   *database.DB
	Node *ipfs.Node
}

type statusResponseWriter struct {
	http.ResponseWriter
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
		// Get file from IPFS node
		file, err := n.Node.Cat(node.CID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// Calculate file size
		size, err := file.Size()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// Build lazy seeker
		content := &lazySeeker{
			size:   size,
			reader: file,
		}

		// Determine file content type
		var ctype string

		// Check if the file is a symlink
		if _, isSymlink := file.(*files.Symlink); isSymlink {
			// FROM IPFS:
			// We should be smarter about resolving symlinks but this is the
			// "most correct" we can be without doing that.
			ctype = "inode/symlink"
		} else {
			ctype = mime.TypeByExtension(filepath.Ext(node.Name))
			if ctype == "" {
				// uses https://github.com/gabriel-vasile/mimetype library to determine the content type.
				// Fixes https://github.com/ipfs/go-ipfs/issues/7252
				mimeType, err := mimetype.DetectReader(content)
				if err != nil {
					http.Error(w, fmt.Sprintf("cannot detect content-type: %s", err.Error()), http.StatusInternalServerError)
					return
				}

				ctype = mimeType.String()
				_, err = content.Seek(0, io.SeekStart)
				if err != nil {
					http.Error(w, "seeker can't seek", http.StatusInternalServerError)
					return
				}
			}
			// Strip the encoding from the HTML Content-Type header and let the
			// browser figure it out.
			//
			// Fixes https://github.com/ipfs/go-ipfs/issues/2203
			if strings.HasPrefix(ctype, "text/html;") {
				ctype = "text/html"
			}
		}
		w.Header().Set("Content-Type", ctype)

		w = &statusResponseWriter{w}
		http.ServeContent(w, r, node.Name, node.Modified, content)
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

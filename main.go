package minee

import (
	"log"
	"net/http"
	"strings"
	"time"
)

const logRequests bool = true
const port string = ":8080"

func main() {

	sh := &StorageHandler{
		static:   "data/static",
		business: "data/business",
		user:     "data/user",
	}
	http.Handle("/", handleMiddleware(Gzip(sh)))
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		var caching bool = true
		if strings.HasSuffix(r.RequestURI, ".css") {
			w.Header().Add("Content-Type", "text/css")
		} else if strings.HasSuffix(r.RequestURI, ".html") {
			w.Header().Add("Content-Type", "text/html")
		} else if strings.HasSuffix(r.RequestURI, "/static/") {
			w.Header().Add("Content-Type", "text/html")
		} else if strings.HasSuffix(r.RequestURI, ".ico") {
			w.Header().Add("Content-Type", "image/x-icon")
		} else if strings.HasSuffix(r.RequestURI, ".png") {
			w.Header().Add("Content-Type", "image/png")
		} else if strings.HasSuffix(r.RequestURI, ".jpg") {
			w.Header().Add("Content-Type", "image/jpeg")
		} else if strings.HasSuffix(r.RequestURI, ".js") {
			w.Header().Add("Content-Type", "application/javascript")
		} else if strings.HasSuffix(r.RequestURI, "json") { // .json or ?json
			caching = false
			w.Header().Add("Content-Type", "application/json")
		} else if !strings.HasSuffix(r.RequestURI, "/") {
			caching = false
		} else {
			caching = false
		}
		if caching {
			w.Header().Add("Cache-Control", "public, max-age=2592000") //30 days
		}
		next.ServeHTTP(w, r)
		if logRequests {
			elapsed := time.Since(start)
			log.Printf("%v %v took %v", r.Method, r.RequestURI, elapsed)
		}
	})
}

type BusinessInfo struct {
	Name           string   `json:"name"`
	Instanceable   bool     `json:"instanceable"`
	Allow          []string `json:"allow"` // Allow other Business Types as Child
	CurrentVersion string   `json:"currentVersion"`

	//Version Specific
	//GetScript      string   `json:"getScript"`
	//PostScript     string   `json:"postScript"`
	// Component
	// Page
}
type StorageHandler struct {
	static   string
	business string
	user     string
}

func (handler *StorageHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	splits := strings.Split(req.RequestURI, "?")
	requestURI := splits[0]
	var queryParam string = ""
	if len(splits) > 1 {
		queryParam = splits[1]
	}
	if req.Method == "GET" {
		handler.handleGetUser(resp, requestURI, queryParam)
	} else if req.Method == "POST" {
		handler.handlePostUser(resp, req)
	} else if req.Method == "PUT" {
		handler.handlePutUser(resp, req)
	} else if req.Method == "PATCH" {
		handler.handlePatchUser(resp, req)
	} else {
		resp.Header().Set("Allow", "GET, POST")
		resp.WriteHeader(405)
	}
}

package minee

import (
	"log"
	"net/http"
	"os"
	"strings"
)

// Minee is a webservice, which can be used to crate a resiliant microservice
// including a pwa single page application.
type Minee struct {
	Port        string
	LogRequests bool
	static      *staticProvider
	business    string
	user        string
}

type staticProvider struct {
	root  string
	cache map[string][]byte
}

func (s *staticProvider) Get(name string) *[]byte {
	if val, ok := s.cache[name]; ok {
		return &val
	} else if fileInfo, err := os.Stat(s.root + name); err == nil && !fileInfo.IsDir() {
		dat, _ := os.ReadFile(s.root + name)
		s.cache[name] = dat
		return &dat
	}
	return nil
}

// New creates a new minee instance.
//
// The instance is preconfigured for a fast start in development mode. For
// example it uses http on port 80. In production mode you should use https on
// port 443.
func New() *Minee {
	return &Minee{
		Port:        ":80",
		LogRequests: true,
		static: &staticProvider{
			root:  "static",
			cache: make(map[string][]byte),
		},
		business: "data/business",
		user:     "data/user",
	}
}

func (minee *Minee) init() {

}

// Start starts the webservice.
//
// The method is blocking. You can use it directly in your main function.
func (minee *Minee) Start() {
	minee.init()
	server := http.Server{
		Addr:    minee.Port,
		Handler: handleMiddleware(gzipper(minee), minee.LogRequests),
	}

	log.Fatal(server.ListenAndServe())
}

type businessInfo struct {
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

func (handler *Minee) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
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

package minee

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Minee is a webservice, which can be used to crate a resiliant microservice
// including a pwa single page application.
type Minee struct {
	Port          string
	LogRequests   bool
	static        *staticProvider
	entities      map[string]*BusinessEntity
	businessRoots map[string]*BusinessEntity
	business      string
	user          string
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
			root: "static",
		},
		entities:      make(map[string]*BusinessEntity),
		businessRoots: make(map[string]*BusinessEntity),
		business:      "data/business",
		user:          "data/user",
	}
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

func (minee *Minee) AddType(be *BusinessEntity) {
	minee.entities[be.Type] = be
	minee.businessRoots[be.ContextRoot] = be
}

func (minee *Minee) init() {
	minee.static.init()
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

type staticResource struct {
	data []byte
}

func (res *staticResource) get(resp http.ResponseWriter) {
	resp.Write(res.data)
}
func (res *staticResource) post(resp http.ResponseWriter, _ []byte) {
	resp.Header().Set("Allow", "GET")
	resp.WriteHeader(405)
}
func (res *staticResource) put(resp http.ResponseWriter, _ []byte) {
	resp.Header().Set("Allow", "GET")
	resp.WriteHeader(405)
}

type statusCodeResource struct {
	statusCode int
	location   string
}

func (res *statusCodeResource) get(resp http.ResponseWriter) {
	resp.WriteHeader(res.statusCode)
	if res.statusCode == 301 {
		resp.Header().Add("Location", res.location)
	}
}
func (res *statusCodeResource) post(resp http.ResponseWriter, _ []byte) {
	resp.WriteHeader(res.statusCode)
}
func (res *statusCodeResource) put(resp http.ResponseWriter, _ []byte) {
	resp.WriteHeader(res.statusCode)
}

type businessResource struct {
	data []byte
}

func (res *businessResource) get(resp http.ResponseWriter) {
	resp.Write(res.data)
}
func (res *businessResource) post(resp http.ResponseWriter, _ []byte) {
	resp.Header().Set("Allow", "GET")
	resp.WriteHeader(405)
}
func (res *businessResource) put(resp http.ResponseWriter, _ []byte) {
	resp.Header().Set("Allow", "GET")
	resp.WriteHeader(405)
}

type userResource struct {
	data     []byte
	location string
}

func (res *userResource) get(resp http.ResponseWriter) {
	resp.Write(res.data)
}
func (res *userResource) post(resp http.ResponseWriter, _ []byte) {
	resp.Header().Set("Allow", "GET")
	resp.WriteHeader(405)
}
func (res *userResource) put(resp http.ResponseWriter, _ []byte) {
	resp.Header().Set("Allow", "GET")
	resp.WriteHeader(405)
}

type resource interface {
	get(http.ResponseWriter)
	post(http.ResponseWriter, []byte)
	put(http.ResponseWriter, []byte)
}

func (minee *Minee) getResource(requestURI string) resource {
	if dat := minee.static.Get(requestURI); dat != nil {
		return &staticResource{data: dat}
	}
	path := filepath.ToSlash(filepath.Dir(requestURI))
	if bi, ok := minee.businessRoots[path]; ok {
		if strings.HasSuffix(requestURI, "/") {
			return &businessResource{data: minee.static.Get("/index.html")}
		}
		base := filepath.Base(requestURI)
		if base == "page.js" && bi.Page != nil {
			dat, _ := os.ReadFile(minee.business + bi.Page.FileName)
			return &staticResource{data: dat}
		}
	}

	filename := minee.user + requestURI
	fileInfo, err := os.Stat(filename)
	if !os.IsNotExist(err) {
		if fileInfo.IsDir() {
			if strings.HasSuffix(filename, "/") {
				return &staticResource{data: minee.static.Get("/index.html")}
			}
			return &statusCodeResource{
				statusCode: 301,
				location:   requestURI + "/",
			}
		}
		dat, _ := os.ReadFile(filename)
		return &userResource{data: dat}
	}
	typePath := filepath.Dir(filename) + "/type"
	if _, err := os.Stat(typePath); err == nil {
		dat, _ := os.ReadFile(typePath)
		typeName := string(dat)
		if bi, ok := minee.entities[typeName]; ok {
			return &userResource{
				location: bi.ContextRoot + "/" + filepath.Base(requestURI),
			}
		}
	}
	return &statusCodeResource{statusCode: 404}
}

func (minee *Minee) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	splits := strings.Split(req.RequestURI, "?")
	requestURI := splits[0]
	//var queryParam string = ""
	//if len(splits) > 1 {
	//	queryParam = splits[1]
	//}

	resource := minee.getResource(requestURI)
	resource.get(resp)
}

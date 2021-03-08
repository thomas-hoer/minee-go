package minee

import (
	"errors"
	"io"
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

// AddType adds a BusinessEntity to the webservice.
func (minee *Minee) AddType(be *BusinessEntity) error {
	if len(be.AllowedSubTypes) != 0 && be.OnPostBefore == nil {
		return errors.New("BusinessEntity with non empty AllowedSubTypes must provide a OnPostBefore function")
	}
	minee.entities[be.Type] = be
	minee.businessRoots[be.ContextRoot] = be
	return nil
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
func (res *staticResource) post(resp http.ResponseWriter, _ []byte, _ map[string][]string) {
	resp.Header().Set("Allow", "GET")
	resp.WriteHeader(405)
}
func (res *staticResource) put(resp http.ResponseWriter, _ []byte, _ map[string][]string) {
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
func (res *statusCodeResource) post(resp http.ResponseWriter, _ []byte, _ map[string][]string) {
	resp.WriteHeader(res.statusCode)
}
func (res *statusCodeResource) put(resp http.ResponseWriter, _ []byte, _ map[string][]string) {
	resp.WriteHeader(res.statusCode)
}

type businessResource struct {
	data       []byte
	bi         *BusinessEntity
	minee      *Minee
	requestURI string
}

func (res *businessResource) get(resp http.ResponseWriter) {
	resp.Write(res.data)
}

func getFirst(headerParam map[string][]string, key string) string {
	if param, ok := headerParam[key]; ok {
		if len(param) >= 1 {
			return param[0]
		}
	}
	return ""
}
func contentTypeToBusinessType(contentType string) string {
	splits := strings.Split(contentType, "/")
	if len(splits) == 2 {
		return strings.ReplaceAll(splits[1], ".", "/")
	}
	return ""
}

func (res *businessResource) post(resp http.ResponseWriter, data []byte, headerParam map[string][]string) {
	businessType := contentTypeToBusinessType(getFirst(headerParam, "Content-Type"))
	if res.bi.AllowedSubTypes.contains(businessType) {
		createType := res.minee.entities[businessType]
		object, err := createType.Unmarshal(data)
		if err != nil {
			log.Print(err)
			return
		}
		context := Context{}
		object, key := res.bi.OnPostBefore(context, object)
		resp.Header().Add("id", key)
		if res.bi.OnPostCompute != nil {
			object = res.bi.OnPostCompute(context, object)
		}
		if createType.OnPostCompute != nil {
			object = createType.OnPostCompute(context, object)
		}
		if createType.OnPostAfter != nil {
			object = createType.OnPostAfter(context, object)
		}
		dataToWrite, err := createType.Marshal(object)
		if err != nil {
			log.Print(err)
			return
		}
		fileName := res.minee.user + res.requestURI + key + "/data.json"
		os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
		if err := os.WriteFile(fileName, dataToWrite, os.ModePerm); err != nil {
			resp.WriteHeader(500)
		} else {
			resp.WriteHeader(201)
		}
	} else {
		resp.WriteHeader(415)
	}
}
func (res *businessResource) put(resp http.ResponseWriter, _ []byte, headerParam map[string][]string) {
	resp.Header().Set("Allow", "GET")
	resp.WriteHeader(405)
}

type resource interface {
	get(http.ResponseWriter)
	post(http.ResponseWriter, []byte, map[string][]string)
	put(http.ResponseWriter, []byte, map[string][]string)
}

func (minee *Minee) getResource(requestURI string) resource {
	if dat := minee.static.Get(requestURI); dat != nil {
		return &staticResource{data: dat}
	}
	path := filepath.ToSlash(filepath.Dir(requestURI))
	if bi, ok := minee.businessRoots[path]; ok {
		if strings.HasSuffix(requestURI, "/") {
			return &businessResource{
				bi:         bi,
				data:       minee.static.Get("/index.html"),
				minee:      minee,
				requestURI: requestURI,
			}
		}
		base := filepath.Base(requestURI)
		if base == "page.js" && bi.Page != nil {
			dat, _ := os.ReadFile(minee.business + bi.Page.FileName)
			return &staticResource{data: dat}
		}
	}

	filename := minee.user + requestURI
	typeName := getType(filepath.Dir(filename) + "/type")
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
		if typeName == nil {
			return &staticResource{data: dat}
		}
		if bi, ok := minee.entities[*typeName]; ok {
			return &businessResource{
				bi:         bi,
				data:       dat,
				minee:      minee,
				requestURI: requestURI,
			}
		}
	}
	if typeName != nil {
		if bi, ok := minee.entities[*typeName]; ok {
			return &statusCodeResource{
				statusCode: 301,
				location:   bi.ContextRoot + "/" + filepath.Base(requestURI),
			}
		}
	}
	return &statusCodeResource{statusCode: 404}
}
func getType(typePath string) *string {
	if dat, err := os.ReadFile(typePath); err == nil {
		typeName := string(dat)
		return &typeName
	}
	return nil
}
func (minee *Minee) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	splits := strings.Split(req.RequestURI, "?")
	requestURI := splits[0]
	//var queryParam string = ""
	//if len(splits) > 1 {
	//	queryParam = splits[1]
	//}

	resource := minee.getResource(requestURI)
	if req.Method == "GET" {
		resource.get(resp)
	} else if req.Method == "POST" {
		dat, _ := io.ReadAll(req.Body)
		resource.post(resp, dat, req.Header)
	} else if req.Method == "PUT" {
		dat, _ := io.ReadAll(req.Body)
		resource.put(resp, dat, req.Header)
	} else {
		resp.WriteHeader(404)
	}
}

package minee

import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	authenticator func(session string) *string
}

func (minee *Minee) Authenticator(authenticator func(session string) *string) {
	minee.authenticator = authenticator
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
	if err := minee.init(); err != nil {
		log.Fatal(err)
	}
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

func (minee *Minee) init() error {
	return minee.static.init()
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
			dat, err := os.ReadFile(minee.business + bi.Page.FileName)
			if err == nil {
				return &staticResource{data: dat}
			}
			return &statusCodeResource{
				statusCode: 500,
				fault:      err,
			}
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
		dat, err := os.ReadFile(filename)
		if err != nil {
			return &statusCodeResource{
				statusCode: 500,
				fault:      err,
			}
		}
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
func getSessionCookie(resp http.ResponseWriter, cookies []*http.Cookie) string {
	for _, cookie := range cookies {
		if cookie.Name == "Session" {
			return cookie.Value
		}
	}
	value := strconv.FormatInt(time.Now().UnixNano(), 36) + "-" + strconv.FormatInt(rand.Int63(), 36) + "-" + strconv.FormatInt(rand.Int63(), 36)
	cookie := &http.Cookie{
		Name:    "Session",
		Value:   value,
		Expires: time.Now().Add(time.Hour * 24 * 30),
	}
	http.SetCookie(resp, cookie)
	return value
}
func (minee *Minee) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	splits := strings.Split(req.RequestURI, "?")
	requestURI := splits[0]
	//var queryParam string = ""
	//if len(splits) > 1 {
	//	queryParam = splits[1]
	//}

	resource := minee.getResource(requestURI)

	var user *string
	if minee.authenticator != nil {
		user = minee.authenticator(getSessionCookie(resp, req.Cookies()))
	}
	context := &Context{
		UserContext: UserContext{
			User: user,
		},
		RequestContext: RequestContext{
			resp: resp,
			req:  req,
		},
	}

	if req.Method == "GET" {
		resource.get(context)
	} else if req.Method == "POST" {
		resource.post(context)
	} else if req.Method == "PUT" {
		resource.put(context)
	} else {
		resp.WriteHeader(404)
	}
}

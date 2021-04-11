package minee

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type resource interface {
	get(*Context)
	post(*Context)
	put(*Context)
}

type staticResource struct {
	data []byte
}

func (res *staticResource) get(context *Context) {
	context.RequestContext.resp.Write(res.data)
}
func (res *staticResource) post(context *Context) {
	resp := context.RequestContext.resp
	resp.Header().Set("Allow", "GET")
	resp.WriteHeader(405)
}
func (res *staticResource) put(context *Context) {
	resp := context.RequestContext.resp
	resp.Header().Set("Allow", "GET")
	resp.WriteHeader(405)
}

type statusCodeResource struct {
	statusCode int
	location   string
	fault      error
}

func (res *statusCodeResource) get(context *Context) {
	res.handleResponse(context.RequestContext.resp)
}
func (res *statusCodeResource) post(context *Context) {
	res.handleResponse(context.RequestContext.resp)
}
func (res *statusCodeResource) put(context *Context) {
	res.handleResponse(context.RequestContext.resp)
}
func (res *statusCodeResource) handleResponse(resp http.ResponseWriter) {
	resp.WriteHeader(res.statusCode)
	if res.fault != nil {
		resp.Header().Add("Error", res.fault.Error())
	}
	if res.statusCode == 301 {
		resp.Header().Add("Location", res.location)
	}
}

type businessResource struct {
	data       []byte
	bi         *BusinessEntity
	minee      *Minee
	requestURI string
}

func (res *businessResource) get(context *Context) {
	resp := context.RequestContext.resp
	biType := res.minee.entities[res.bi.Type]
	if biType.Filter != nil {
		object, err := biType.Unmarshal(res.data)
		if err != nil {
			log.Print(err)
			return
		}
		filteredObject := biType.Filter(context, object)
		data, err := biType.Marshal(filteredObject)
		if err != nil {
			log.Print(err)
			return
		}
		resp.Write(data)
	} else {
		resp.Write(res.data)
	}
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

func (res *businessResource) post(context *Context) {
	resp := context.RequestContext.resp
	businessType := contentTypeToBusinessType(getFirst(resp.Header(), "Content-Type"))
	if res.bi.AllowedSubTypes.contains(businessType) {
		createType := res.minee.entities[businessType]
		data, err := io.ReadAll(context.RequestContext.req.Body)
		if err != nil {
			log.Print(err)
			resp.WriteHeader(500)
			return
		}
		object, err := createType.Unmarshal(data)
		if err != nil {
			log.Print(err)
			resp.WriteHeader(500)
			return
		}
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
			resp.WriteHeader(500)
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
func (res *businessResource) put(context *Context) {
	resp := context.RequestContext.resp
	resp.Header().Set("Allow", "GET")
	resp.WriteHeader(405)
}

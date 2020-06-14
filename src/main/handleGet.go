package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (handler *StorageHandler) handleGetUser(resp http.ResponseWriter, requestURI, queryParam string) {
	filename := handler.user + requestURI
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		handler.handleGetType(resp, requestURI, queryParam)
	} else if fileInfo.IsDir() {
		if !strings.HasSuffix(filename, "/") {
			resp.Header().Set("Location", requestURI+"/")
			resp.WriteHeader(301)
		} else {
			handler.handleGetIndex(resp, handler.user, requestURI, queryParam)
		}
	} else {
		dat, _ := ioutil.ReadFile(filename)
		resp.Write(dat)
	}
}
func (handler *StorageHandler) handleGetIndex(resp http.ResponseWriter, base, requestURI, queryParam string) {
	pathToRoot := strings.Repeat("../", strings.Count(requestURI, "/")-1)
	fileInfos, _ := ioutil.ReadDir(base + requestURI)
	names := make([]string, 0)
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			names = append(names, `"`+fileInfo.Name()+`/"`)
		} else {
			names = append(names, `"`+fileInfo.Name()+`"`)
		}
	}
	jsonOutput := `[` + strings.Join(names, `,`) + `]`
	if queryParam == "json" {
		resp.Header().Add("Content-Type", "application/json")
		resp.Write([]byte(jsonOutput))
	} else if queryParam == "module" {
		resp.Header().Add("Content-Type", "application/javascript")
		resp.Write([]byte("'use strict';\nconst data=" + jsonOutput + "\nexport {data}"))
	} else {
		resp.Header().Add("Content-Type", "text/html")
		templData, _ := ioutil.ReadFile(handler.static + "/index.html")
		tmpl, err := template.New("index").Parse(string(templData))
		if err != nil {
			resp.Write([]byte(err.Error()))
			resp.WriteHeader(500)
			return
		}
		tmpl.Execute(resp, struct{ PageTitle, PathToRoot, JsonOutput string }{"TODO", pathToRoot, jsonOutput})
	}
}
func contains(list []string, stringToFind string) bool {
	for _, le := range list {
		if strings.Contains(le, stringToFind) {
			return true
		}
	}
	return false
}
func (handler *StorageHandler) handleGetType(resp http.ResponseWriter, requestURI, queryParam string) {
	typefile := filepath.Dir(handler.user+requestURI) + "/type"
	if fileInfo, err := os.Stat(typefile); err == nil && !fileInfo.IsDir() {
		dat, _ := ioutil.ReadFile(typefile)
		typeRoot := string(dat)
		filename := filepath.Base(handler.user + requestURI)
		redirect := "/" + typeRoot + "/" + filename
		if redirect == requestURI {
			handler.handleGetStatic(resp, requestURI, queryParam)
		} else {
			resp.Header().Add("Location", redirect)
			resp.WriteHeader(303)
		}
	} else {
		handler.handleGetStatic(resp, requestURI, queryParam)
	}
}
func (handler *StorageHandler) handleGetStatic(resp http.ResponseWriter, requestURI, queryParam string) {
	if fileInfo, err := os.Stat(handler.static + requestURI); err == nil && !fileInfo.IsDir() {
		dat, _ := ioutil.ReadFile(handler.static + requestURI)
		resp.Write(dat)
	} else {
		handler.handleGetBusiness(resp, requestURI, queryParam)
	}
}
func (handler *StorageHandler) handleGetBusiness(resp http.ResponseWriter, requestURI, queryParam string) {
	if fileInfo, err := os.Stat(handler.business + requestURI); err == nil {
		if !fileInfo.IsDir() {
			dat, _ := ioutil.ReadFile(handler.business + requestURI)
			resp.Write(dat)
		} else {
			handler.handleGetIndex(resp, handler.business, requestURI, queryParam)
		}
		return
	}

	root := filepath.Dir(handler.business + requestURI)
	businessInfo := root + "/info.json"
	if fileInfo, err := os.Stat(businessInfo); err == nil && !fileInfo.IsDir() {
		dat, _ := ioutil.ReadFile(businessInfo)
		var bi BusinessInfo
		json.Unmarshal(dat, &bi)
		if bi.CurrentVersion != "" {
			filename := filepath.Base(handler.user + requestURI)
			redirect := "versions/" + bi.CurrentVersion + "/" + filename
			resp.Header().Add("Location", redirect)
			resp.WriteHeader(303)
			return
		}
	}
	resp.WriteHeader(404)

}

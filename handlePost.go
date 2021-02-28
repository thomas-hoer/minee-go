package minee

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type sequence struct {
	NextID int `json:"nextId"`
}

func (handler *storageHandler) handlePostUser(resp http.ResponseWriter, req *http.Request) {
	filename := handler.user + req.RequestURI
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		resp.WriteHeader(404)
		return
	} else if !fileInfo.IsDir() {
		resp.Header().Set("Allow", "GET")
		resp.WriteHeader(405)
		return
	}

	bc := handler.getBusinessContext(req.RequestURI)
	dataType := typeOf(req)
	if dataType == nil {
		resp.WriteHeader(415)
		return
	}
	bc.setContentType(*dataType)
	newData, _ := ioutil.ReadAll(req.Body)
	newDataString := string(newData)

	newID := bc.generateID(newDataString)

	newPath := filename + newID + "/"
	bc.setTargetURI(req.RequestURI + newID + "/")

	os.MkdirAll(newPath, os.ModePerm)
	ioutil.WriteFile(newPath+"type", []byte(*dataType), os.ModePerm)
	ioutil.WriteFile(newPath+"data.json", []byte(newDataString), os.ModePerm)
	resp.Header().Add("Location", req.RequestURI+newID+"/"+bc.relocate)
	resp.WriteHeader(201)
}

func (handler *storageHandler) handlePutUser(resp http.ResponseWriter, req *http.Request) {
	filename := handler.user + req.RequestURI
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		resp.WriteHeader(404)
		return
	} else if !fileInfo.IsDir() {
		resp.Header().Set("Allow", "GET")
		resp.WriteHeader(405)
		return
	}

	bc := handler.getBusinessContext(req.RequestURI)

	if contentTypeData, err := ioutil.ReadFile(filename + "type"); err != nil {
		bc.setContentType(string(contentTypeData))
	}
	newData, _ := ioutil.ReadAll(req.Body)
	newDataString := string(newData)
	if dataType := typeOf(req); dataType != nil {
		ioutil.WriteFile(filename+"type", []byte(*dataType), os.ModePerm)
		bc.setContentType(*dataType)
	}
	ioutil.WriteFile(filename+"data.json", []byte(newDataString), os.ModePerm)
	resp.WriteHeader(204)
}

func (handler *storageHandler) handlePatchUser(resp http.ResponseWriter, req *http.Request) {
	filename := handler.user + req.RequestURI
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		resp.WriteHeader(404)
		return
	} else if !fileInfo.IsDir() {
		resp.Header().Set("Allow", "GET")
		resp.WriteHeader(405)
		return
	}

	bc := handler.getBusinessContext(req.RequestURI)

	if contentTypeData, err := ioutil.ReadFile(filename + "type"); err != nil {
		bc.setContentType(string(contentTypeData))
	}
	newData, _ := ioutil.ReadAll(req.Body)
	patchData := string(newData)
	if dataType := typeOf(req); dataType != nil {
		ioutil.WriteFile(filename+"type", []byte(*dataType), os.ModePerm)
		bc.setContentType(*dataType)
	}
	dataString := readAsJsString(filename + "data.json")
	dataString = bc.doPatch(patchData, dataString)
	ioutil.WriteFile(filename+"data.json", []byte(dataString), os.ModePerm)
	resp.WriteHeader(201)
}

func readAsJsString(path string) string {
	if dat, err := ioutil.ReadFile(path); err == nil {
		return string(dat)
	}
	return "null"
}
func typeOf(req *http.Request) *string {
	ct := strings.Split(req.Header.Get("Content-Type"), "/")
	if len(ct) < 2 {
		return nil
	}
	applicationType := strings.Replace(ct[1], ".", "/", -1)
	return &applicationType
}

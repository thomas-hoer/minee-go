package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Sequence struct {
	NextId int `json:"nextId"`
}

func (handler *StorageHandler) handlePostUser(resp http.ResponseWriter, req *http.Request) {
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

	seqFileName := handler.user + req.RequestURI + "sequence.json"
	seqdat, err := ioutil.ReadFile(seqFileName)
	if err != nil {
		log.Print(seqFileName+"sequence.json for POST event not found", err)
		resp.Header().Set("Allow", "GET")
		resp.WriteHeader(405)
		return
	}
	var seq Sequence
	json.Unmarshal(seqdat, &seq)
	nextId := seq.NextId
	seq.NextId++
	seqdat, _ = json.Marshal(&seq)
	ioutil.WriteFile(seqFileName, seqdat, os.ModePerm)
	bc.setNewId(nextId)
	newPath := filename + strconv.Itoa(nextId) + "/"
	bc.setTargetURI(req.RequestURI + strconv.Itoa(nextId) + "/")
	newData, _ := ioutil.ReadAll(req.Body)
	newDataString := string(newData)
	os.MkdirAll(newPath, os.ModePerm)
	if dataType := typeOf(req); dataType != nil {
		ioutil.WriteFile(newPath+"type", []byte(*dataType), os.ModePerm)
		bc.setContentType(*dataType)
	}
	oldDataString := "null"
	newDataString = bc.beforePost(newDataString)
	newDataString = bc.compute(newDataString, oldDataString)
	newDataString = bc.afterPost(newDataString)
	ioutil.WriteFile(newPath+"data.json", []byte(newDataString), os.ModePerm)
	resp.Header().Add("Location", req.RequestURI+strconv.Itoa(nextId)+"/"+bc.relocate)
	resp.WriteHeader(201)
}

func (handler *StorageHandler) handlePutUser(resp http.ResponseWriter, req *http.Request) {
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
	oldDataString := readAsJsString(filename + "data.json")
	newDataString = bc.compute(newDataString, oldDataString)
	newDataString = bc.afterPut(newDataString, oldDataString)
	ioutil.WriteFile(filename+"data.json", []byte(newDataString), os.ModePerm)
	resp.WriteHeader(204)
}

func (handler *StorageHandler) handlePatchUser(resp http.ResponseWriter, req *http.Request) {
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
	} else {
		return "null"
	}

}
func typeOf(req *http.Request) *string {
	ct := strings.Split(req.Header.Get("Content-Type"), "/")
	if len(ct) < 2 {
		return nil
	}
	applicationType := strings.Replace(ct[1], ".", "/", -1)
	return &applicationType
}

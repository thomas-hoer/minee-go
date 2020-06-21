package main

import (
	"encoding/json"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type businessContext struct {
	vm           *otto.Otto
	requestURI   string
	targetURI    string
	newId        string
	contentType  string
	rootBusiness string
	rootUser     string
	relocate     string
}

func (handler *StorageHandler) getBusinessContext(requestURI string) *businessContext {
	bc := &businessContext{
		vm:           otto.New(),
		requestURI:   requestURI,
		targetURI:    requestURI,
		rootBusiness: handler.business,
		rootUser:     handler.user,
	}
	bc.vm.Set("writeFile", func(call otto.FunctionCall) otto.Value {
		fileName := call.Argument(0).String()
		data := call.Argument(1).String()
		fullPath := bc.rootUser + bc.targetURI + fileName
		os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
		ioutil.WriteFile(fullPath, []byte(data), os.ModePerm)
		return otto.Value{}
	})
	bc.vm.Set("locate", func(call otto.FunctionCall) otto.Value {
		relativePath := call.Argument(0).String()
		bc.relocate = relativePath
		return otto.Value{}
	})
	bc.vm.Set("post", func(call otto.FunctionCall) otto.Value {
		path := call.Argument(0).String()
		mimeType := call.Argument(1).String()
		data := call.Argument(2).String()
		resp, _ := http.Post("http://localhost"+port+bc.targetURI+path, mimeType, strings.NewReader(data))
		respData, _ := ioutil.ReadAll(resp.Body)
		respString := string(respData)
		respValue, _ := bc.vm.ToValue(respString)
		return respValue
	})
	return bc
}

func (bc *businessContext) setContentType(contentType string) {
	bc.contentType = contentType
}
func (bc *businessContext) setTargetURI(targetURI string) {
	bc.targetURI = targetURI
}

func (bc *businessContext) compute(input string, oldData string) string {
	path := bc.rootBusiness + "/" + bc.contentType + "/compute.js"
	if script, err := ioutil.ReadFile(path); err == nil {
		bc.vm.Run(string(script))
	} else {
		log.Print(err)
		return input
	}
	bc.vm.Run("var data = " + input)
	bc.vm.Run("var oldData = " + oldData)
	bc.vm.Run(`var context = {
	id:"` + bc.newId + `",
	type:"` + bc.contentType + `"
	}`)
	if result, err := bc.vm.Run("JSON.stringify(compute(data,oldData,context),null,'\t')"); err != nil {
		log.Print(path, err)
		return input
	} else {
		return result.String()
	}
}

func (bc *businessContext) doPatch(patchData string, oldData string) string {
	path := bc.rootBusiness + "/" + bc.contentType + "/onPatch.js"
	if script, err := ioutil.ReadFile(path); err == nil {
		bc.vm.Run(string(script))
	} else {
		log.Print(err)
		return oldData
	}

	bc.vm.Run("var patchData = " + patchData)
	bc.vm.Run("var oldData = " + oldData)
	if result, err := bc.vm.Run("JSON.stringify(onPatch(patchData,oldData),null,'\t')"); err != nil {
		log.Print(path, err)
		return oldData
	} else {
		return result.String()
	}
}

func (bc *businessContext) beforePost(input string) string {
	bc.vm.Run("var data = " + input)
	var result string
	result = runScript(bc.vm, input, bc.rootBusiness+bc.requestURI+"beforePost.js", "beforePost(data)")
	bc.vm.Run("var data = " + result)
	return runScript(bc.vm, result, bc.rootBusiness+"/"+bc.contentType+"/beforePost.js", "beforePost(data)")
}

func (bc *businessContext) afterPost(input string) string {
	bc.vm.Run(`var context = {
	id:"` + bc.newId + `",
	type:"` + bc.contentType + `"
	}`)
	bc.vm.Run("var data = " + input)
	var result string
	result = runScript(bc.vm, input, bc.rootBusiness+"/"+bc.contentType+"/afterPost.js", "afterPost(data,context)")
	bc.vm.Run("var data = " + result)
	return runScript(bc.vm, result, bc.rootBusiness+bc.requestURI+"afterPost.js", "afterPost(data,context)")
}

func (bc *businessContext) afterPut(input string, oldData string) string {
	path := bc.rootBusiness + "/" + bc.contentType + "/afterPut.js"
	if script, err := ioutil.ReadFile(path); err == nil {
		bc.vm.Run(string(script))
	} else {
		log.Print(err)
		return input
	}

	bc.vm.Run("var data = " + input)
	bc.vm.Run("var oldData = " + oldData)
	bc.vm.Run(`var context = {
	id:"` + bc.newId + `",
	type:"` + bc.contentType + `"
	}`)
	if result, err := bc.vm.Run("JSON.stringify(afterPut(data,oldData,context),null,'\t')"); err != nil {
		log.Print(path, err)
		return input
	} else {
		return result.String()
	}
}

func (bc *businessContext) generateId(data string) string {
	path := bc.rootBusiness + "/" + bc.contentType + "/generateId.js"
	if script, err := ioutil.ReadFile(path); err == nil {
		bc.vm.Run(string(script))
	} else {
		log.Print(err)
		return bc.generateIdFromSequence()
	}

	bc.vm.Run("var data = " + data)
	if result, err := bc.vm.Run("generateId(data)"); err != nil {
		log.Print(path, err)
		return bc.generateIdFromSequence()
	} else {
		bc.newId = result.String()
		return bc.newId
	}
}

func (bc *businessContext) generateIdFromSequence() string {
	seqFileName := bc.rootUser + bc.requestURI + "sequence.json"
	seqdat, err := ioutil.ReadFile(seqFileName)
	var seq Sequence
	if err != nil {
		log.Print(seqFileName + "sequence.json not found, generating new one")
	} else {
		json.Unmarshal(seqdat, &seq)
	}
	nextId := seq.NextId
	seq.NextId++
	seqdat, _ = json.Marshal(&seq)
	ioutil.WriteFile(seqFileName, seqdat, os.ModePerm)
	bc.newId = strconv.Itoa(nextId)
	return bc.newId
}

func runScript(vm *otto.Otto, input, path, execution string) string {
	if script, err := ioutil.ReadFile(path); err == nil {
		vm.Run(string(script))
	} else {
		log.Print(err)
		return input
	}
	log.Print("run ", execution, " from ", path)
	if result, err := vm.Run("JSON.stringify(" + execution + ",null,'\t')"); err != nil {
		log.Print(path, err)
		return input
	} else {
		return result.String()
	}
}

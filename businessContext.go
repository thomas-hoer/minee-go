package minee

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type businessContext struct {
	requestURI   string
	targetURI    string
	newID        string
	contentType  string
	rootBusiness string
	rootUser     string
	relocate     string
}

func (handler *storageHandler) getBusinessContext(requestURI string) *businessContext {
	bc := &businessContext{
		requestURI:   requestURI,
		targetURI:    requestURI,
		rootBusiness: handler.business,
		rootUser:     handler.user,
	}
	return bc
}

func (bc *businessContext) setContentType(contentType string) {
	bc.contentType = contentType
}
func (bc *businessContext) setTargetURI(targetURI string) {
	bc.targetURI = targetURI
}

func (bc *businessContext) doPatch(patchData string, oldData string) string {
	return oldData
}

func (bc *businessContext) generateID(data string) string {
	return bc.generateIDFromSequence()
}

func (bc *businessContext) generateIDFromSequence() string {
	seqFileName := bc.rootUser + bc.requestURI + "sequence.json"
	seqdat, err := ioutil.ReadFile(seqFileName)
	var seq sequence
	if err != nil {
		log.Print(seqFileName + "sequence.json not found, generating new one")
	} else {
		json.Unmarshal(seqdat, &seq)
	}
	nextID := seq.NextID
	seq.NextID++
	seqdat, _ = json.Marshal(&seq)
	ioutil.WriteFile(seqFileName, seqdat, os.ModePerm)
	bc.newID = strconv.Itoa(nextID)
	return bc.newID
}

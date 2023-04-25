package http

import (
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ImposterHandler create specific handler for the received imposter
func ImposterHandler(imposter Imposter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if imposter.Delay() > 0 {
			time.Sleep(imposter.Delay())
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			updateResponseTemplate(r, imposter)
		}
		writeHeaders(imposter, w)
		w.WriteHeader(imposter.Response.Status)
		writeBody(imposter, w)
	}
}

func updateResponseTemplate(r *http.Request, imposter Imposter) {
	if imposter.Response.BodyTemplateFile != nil {
		bodyTemplateFile := imposter.CalculateFilePath(*imposter.Response.BodyTemplateFile)
		if _, err := os.Stat(bodyTemplateFile); os.IsNotExist(err) {
			log.Printf("the body template file %s not found\n", bodyTemplateFile)
			return
		}
		bodyFile := imposter.CalculateFilePath(*imposter.Response.BodyFile)
		if _, err := os.Stat(bodyFile); os.IsNotExist(err) {
			log.Printf("the body file %s not found\n", bodyFile)
			return
		}
		requestBody, _ := io.ReadAll(r.Body)

		templateBody := fetchBodyFromFile(bodyTemplateFile)

		updatedBody := strings.Replace(string(templateBody), "\"data\": \"DATA\"", "\"data\": "+string(requestBody), 1)
		err := ioutil.WriteFile(bodyFile, []byte(updatedBody), fs.ModePerm)
		if err != nil {
			log.Printf("Error updating the response for  the file %s: %v\n", bodyFile, err)
		}
	}
	return
}

func writeHeaders(imposter Imposter, w http.ResponseWriter) {
	if imposter.Response.Headers == nil {
		return
	}

	for key, val := range *imposter.Response.Headers {
		w.Header().Set(key, val)
	}
}

func writeBody(imposter Imposter, w http.ResponseWriter) {
	wb := []byte(imposter.Response.Body)

	if imposter.Response.BodyFile != nil {
		bodyFile := imposter.CalculateFilePath(*imposter.Response.BodyFile)
		wb = fetchBodyFromFile(bodyFile)
	}
	w.Write(wb)
}

func fetchBodyFromFile(bodyFile string) (bytes []byte) {
	if _, err := os.Stat(bodyFile); os.IsNotExist(err) {
		log.Printf("the body file %s not found\n", bodyFile)
		return
	}

	f, _ := os.Open(bodyFile)
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("imposible read the file %s: %v\n", bodyFile, err)
	}
	return
}

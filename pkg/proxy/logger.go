package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"log"
	"os"
	"time"
)

type Logger struct {
	fileRequest  *os.File
	fileResponse *os.File
}

func (l *Logger) log(request *request, response *response, logtofile bool, prettify bool) {
	if logtofile {
		l.logToFile(request, response)
	}
	l.logToConsole(request, response, prettify)
}

func (l *Logger) logToFile(request *request, response *response) {
	req, _ := json.Marshal(request)
	res, _ := json.Marshal(response)
	if _, err := l.fileRequest.Write(append(req, "\n"...)); err != nil {
		log.Printf("Error: ", err.Error())
	}
	if _, err := l.fileResponse.Write(append(res, "\n"...)); err != nil {
		log.Fatal("Error: ", err.Error())
	}
}

func (l *Logger) logToConsole(request *request, response *response, prettify bool) {
	if prettify {
		var prettyBody bytes.Buffer
		json.Indent(&prettyBody, []byte(request.Body), "", "  ")
		t := time.Now()

		fmt.Println()
		fmt.Println("========================")
		fmt.Println(t.Format("2006/01/02 15:04:05"))
		fmt.Println("Remote Address: ", request.RemoteAddr)
		fmt.Println("Request URI: ", request.RequestUri)
		fmt.Println("Method: ", request.Method)
		fmt.Println("Status: ", response.StatusCode)
		fmt.Printf("Took: %.3fs\n", request.Elapsed)
		fmt.Println("Body: ")
		fmt.Println(string(prettyBody.Bytes()))
	} else {
		log.Printf(" -> %s; %s; %s; %s; %d; %.3fs\n",
			request.Method, request.RemoteAddr,
			request.RequestUri, request.Body,
			response.StatusCode, request.Elapsed)
	}
}

func (l *Logger) enableFileLogger() {
	var (
		err          error
		fileRequest  *os.File
		fileResponse *os.File
	)

	u1 := uuid.NewV4()
	u2 := uuid.NewV4()
	requestFname := fmt.Sprintf("request-%s.log", u1.String())
	responseFname := fmt.Sprintf("response-%s.log", u2.String())

	if fileRequest, err = os.Create(requestFname); err != nil {
		log.Fatalln(err.Error())
	}
	if fileResponse, err = os.Create(responseFname); err != nil {
		log.Fatalln(err.Error())
	}

	l.fileRequest = fileRequest
	l.fileResponse = fileResponse

}

func (l *Logger) ShutDownFileLogger() {
	l.fileRequest.Close()
	l.fileResponse.Close()
}

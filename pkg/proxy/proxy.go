package proxy

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	uuid "github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type request struct {
	RequestId  string
	DateTime   string
	RemoteAddr string
	RequestUri string
	Method     string
	Elapsed    float64
	Body       string
}

type response struct {
	RequestId  string
	Body       string
	StatusCode int
}

type proxy struct {
	scheme       string
	host         string
	region       string
	service      string
	endpoint     string
	verbose      bool
	prettify     bool
	logtofile    bool
	nosignreq    bool
	fileRequest  *os.File
	fileResponse *os.File
	logger       *Logger
	credentials  *credentials.Credentials
}

//var logger *logger

func New(endpoint string, verbose bool, prettify bool, logtofile bool, nosignreq bool) *proxy {
	p := &proxy{
		endpoint:  endpoint,
		verbose:   verbose,
		prettify:  prettify,
		logtofile: logtofile,
		nosignreq: nosignreq,
	}
	p.parseEndpoint()
	if p.logtofile {
		p.logger = new(Logger)
		p.logger.enableFileLogger()
	}
	return p
}

func (p *proxy) HandlerProxy(w http.ResponseWriter, r *http.Request) {
	requestStarted := time.Now()
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Fatalln("error while dumping request. Error: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	ep := *r.URL
	ep.Host = p.host
	ep.Scheme = p.scheme
	ep.Path = path.Clean(ep.Path)

	req, err := http.NewRequest(r.Method, ep.String(), r.Body)
	if err != nil {
		log.Fatalln("error creating new request. ", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	addKibanaHeaders(r.Header, req.Header)

	// Make signV4 optional
	if !p.nosignreq {
		// Start AWS session from ENV, Shared Creds or EC2Role
		signer := p.getSigner()

		// Sign the request with AWSv4
		payload := bytes.NewReader(replaceBody(req))
		signer.Sign(req, payload, p.service, p.region, time.Now())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !p.nosignreq {
		// AWS credentials expired, need to generate fresh ones
		if resp.StatusCode == 403 {
			p.credentials = nil
			return
		}
	}

	defer resp.Body.Close()

	// Write back headers to requesting client
	copyHeaders(w.Header(), resp.Header)

	// Send response back to requesting client
	body := bytes.Buffer{}
	if _, err := io.Copy(&body, resp.Body); err != nil {
		log.Fatalln(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body.Bytes())

	requestEnded := time.Since(requestStarted)

	if p.verbose {
		/*############################
		## Logging
		############################*/

		rawQuery := string(dump)
		rawQuery = strings.Replace(rawQuery, "\n", " ", -1)
		regex, _ := regexp.Compile("{.*}")
		regEx, _ := regexp.Compile("_msearch|_bulk")
		queryEx := regEx.FindString(rawQuery)

		var query string

		if len(queryEx) == 0 {
			query = regex.FindString(rawQuery)
		} else {
			query = ""
		}
		requestID := uuid.NewV4()
		request := &request{
			RequestId:  requestID.String(),
			DateTime:   time.Now().Format("2006/01/02 15:04:05"),
			RemoteAddr: r.RemoteAddr,
			RequestUri: ep.RequestURI(),
			Method:     r.Method,
			Elapsed:    requestEnded.Seconds(),
			Body:       query,
		}

		response := &response{
			RequestId:  requestID.String(),
			StatusCode: resp.StatusCode,
			Body:       string(body.Bytes()),
		}
		p.logger.log(request, response, p.logtofile, p.prettify)
	}
}

func (p *proxy) getSigner() *v4.Signer {
	// Refresh credentials after expiration. Required for STS
	if p.credentials == nil {
		sess := session.Must(session.NewSession())
		creds := sess.Config.Credentials
		_, err := creds.Get()
		if err != nil {
			log.Fatal("error while getting AWS creds. Error: ", err.Error())
		}
		p.credentials = creds
		log.Println("Generated fresh AWS Credentials object")
	}
	return v4.NewSigner(p.credentials)
}

func (p *proxy) parseEndpoint() {
	var link *url.URL
	var err error

	if link, err = url.Parse(p.endpoint); err != nil {
		log.Fatal("error: failure while parsing endpoint: %s. Error: %s",
			p.endpoint, err.Error())
	}

	// Only http/https are supported schemes
	switch link.Scheme {
	case "http", "https":
	default:
		link.Scheme = "https"
	}

	// Unknown schemes sometimes result in empty host value
	if link.Host == "" {
		log.Fatal("error: empty host or protocol information in submitted endpoint (%s)",
			p.endpoint)
	}

	// AWS SignV4 enabled, extract required parts for signing process
	if !p.nosignreq {
		// Extract region and service from link
		parts := strings.Split(link.Host, ".")

		if len(parts) == 5 {
			p.region, p.service = parts[1], parts[2]
		} else {
			log.Fatal("error: submitted endpoint is not a valid Amazon ElasticSearch Endpoint")
		}
	}

	// Update proxy struct
	p.scheme = link.Scheme
	p.host = link.Host
}

// Recent versions of ES/Kibana require
// "kbn-version" and "content-type: application/json"
// headers to exist in the request.
// If missing requests fails.
func addKibanaHeaders(src, dest http.Header) {
	if val, ok := src["Kbn-Version"]; ok {
		dest.Add("Kbn-Version", val[0])
	}

	if val, ok := src["Content-Type"]; ok {
		dest.Add("Content-Type", val[0])
	}
}

// Signer.Sign requires a "seekable" body to sum body's sha256
func replaceBody(req *http.Request) []byte {
	if req.Body == nil {
		return []byte{}
	}
	payload, _ := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewReader(payload))
	return payload
}

func copyHeaders(dst, src http.Header) {
	for k, vals := range src {
		for _, v := range vals {
			dst.Add(k, v)
		}
	}
}

func (p *proxy) ShutDownProxy() {
	p.logger.ShutDownFileLogger()
}

package queue

import (
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
	"encoding/json"
	"time"
	"bytes"
	"crypto/tls"
	class "github.com/matehaxor03/holistic_db_client/class"
)

type WebServer struct {
	Start      			func() ([]error)
}

func NewWebServer(port string, server_crt_path string, server_key_path string, queue_domain_name string, queue_port string) (*WebServer, []error) {
	var errors []error

	transport_config := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	http_client := http.Client{
		Timeout: 120 * time.Second,
		Transport: transport_config,
	}

	domain_name, domain_name_errors := class.NewDomainName(&queue_domain_name)
	if domain_name_errors != nil {
		errors = append(errors, domain_name_errors...)
	}
	//var this_holisic_queue_server *HolisticQueueServer

	//todo: add filters to fields
	data := class.Map{
		"[port]": class.Map{"value": class.CloneString(&port), "mandatory": true},
		"[server_crt_path]": class.Map{"value": class.CloneString(&server_crt_path), "mandatory": true},
		"[server_key_path]": class.Map{"value": class.CloneString(&server_key_path), "mandatory": true},
		"[queue_port]": class.Map{"value": class.CloneString(&queue_port), "mandatory": true},
		"[queue_domain_name]": class.Map{"value": class.CloneDomainName(domain_name), "mandatory": true},
	}

	getPort := func() *string {
		port, _ := data.M("[port]").GetString("value")
		return class.CloneString(port)
	}

	getServerCrtPath := func() *string {
		crt, _ := data.M("[server_crt_path]").GetString("value")
		return class.CloneString(crt)
	}

	getServerKeyPath := func() *string {
		key, _ := data.M("[server_key_path]").GetString("value")
		return class.CloneString(key)
	}

	
	getQueuePort := func() *string {
		port, _ := data.M("[queue_port]").GetString("value")
		return class.CloneString(port)
	}

	getQueueDomainName := func() *class.DomainName {
		return class.CloneDomainName(data.M("[queue_domain_name]").GetObject("value").(*class.DomainName))
	}

	queue_url := fmt.Sprintf("https://%s:%s/", *(getQueueDomainName().GetDomainName()), *getQueuePort())

	validate := func() []error {
		return class.ValidateData(data, "WebServer")
	}

	/*
	setHolisticQueueServer := func(holisic_queue_server *HolisticQueueServer) {
		this_holisic_queue_server = holisic_queue_server
	}*/

	/*
	getHolisticQueueServer := func() *HolisticQueueServer {
		return this_holisic_queue_server
	}*/

	formatRequest := func(r *http.Request) string {
		var request []string
	
		url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
		request = append(request, url)
		request = append(request, fmt.Sprintf("Host: %v", r.Host))
		for name, headers := range r.Header {
			name = strings.ToLower(name)
			for _, h := range headers {
				request = append(request, fmt.Sprintf("%v: %v", name, h))
			}
		}
	
		if r.Method == "POST" {
			r.ParseForm()
			request = append(request, "\n")
			request = append(request, r.Form.Encode())
		}
	
		return strings.Join(request, "\n")
	}

	processRequest := func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" || req.Method == "PATCH" || req.Method == "PUT" {
			json_payload := class.Map{}
			body_payload, body_payload_error := ioutil.ReadAll(req.Body);
			if body_payload_error != nil {
				w.Write([]byte(body_payload_error.Error()))
			} else {
				json.Unmarshal([]byte(body_payload), &json_payload)
				uuid, _ := ioutil.ReadFile("/proc/sys/kernel/random/uuid")
				trace_id := fmt.Sprintf("%v%s", time.Now().UnixNano(), string(uuid))
				json_payload.SetString("trace_id", &trace_id)

				json_bytes := []byte(json_payload.ToJSONString())
				json_reader := bytes.NewReader(json_bytes)

				request, request_error := http.NewRequest(http.MethodPost, queue_url, json_reader)

				if request_error != nil {
					w.Write([]byte(request_error.Error()))
				} else {
					request.Header.Set("Content-Type", "application/json")
					_, http_response_error := http_client.Do(request)
					if http_response_error != nil {
						w.Write([]byte(http_response_error.Error()))
					} else {
						w.Write([]byte("ok"))
					}
				}
			}
		} else {
			w.Write([]byte(formatRequest(req)))
		}
	}

	x := WebServer{
		Start: func() []error {
			var errors []error

			buildHandler := http.FileServer(http.Dir("static"))
			http.Handle("/", buildHandler)

			http.HandleFunc("/api", processRequest)

			err := http.ListenAndServeTLS(":" + *(getPort()), *(getServerCrtPath()), *(getServerKeyPath()), nil)
			if err != nil {
				errors = append(errors, err)
			}

			if len(errors) > 0 {
				return errors
			}

			return nil
		},
	}
	//setHolisticQueueServer(&x)

	validate_errors := validate()
	if validate_errors != nil {
		errors = append(errors, validate_errors...)
	}

	if len(errors) > 0 {
		return nil, errors
	}

	return &x, nil
}

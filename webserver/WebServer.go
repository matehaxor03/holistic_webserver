package queue

import (
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
	//"encoding/json"
	"time"
	"bytes"
	"crypto/tls"
	"sync"
	"crypto/rand"
	class "github.com/matehaxor03/holistic_db_client/class"
)

type WebServer struct {
	Start      			func() ([]error)
}

func NewWebServer(port string, server_crt_path string, server_key_path string, queue_domain_name string, queue_port string) (*WebServer, []error) {
	var errors []error
	var trace_id_lock sync.Mutex
	var messageCount uint64


	generate_guid := func() string {
		byte_array := make([]byte, 16)
		rand.Read(byte_array)
		guid := fmt.Sprintf("%X-%X-%X-%X-%X", byte_array[0:4], byte_array[4:6], byte_array[6:8], byte_array[8:10], byte_array[10:])
		return guid
	}

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

	incrementMessageCount := func() uint64 {
		messageCount++
		return messageCount
	}

	get_trace_id := func() string {
		trace_id_lock.Lock()
		defer trace_id_lock.Unlock()
		trace_id := fmt.Sprintf("%v-%s-%d", time.Now().UnixNano(), generate_guid(), incrementMessageCount())
		return trace_id
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

	/*
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
	}*/

	write_response := func(w http.ResponseWriter, result class.Map, write_response_errors []error) {
		if len(write_response_errors) > 0 {
			result.SetNil("data")
			result.SetErrors("[errors]", &write_response_errors)
		}

		result_as_string, result_as_string_errors := result.ToJSONString()
		if result_as_string_errors != nil {
			write_response_errors = append(write_response_errors, result_as_string_errors...)
		}
		
		if result_as_string_errors == nil {
			w.Write([]byte(*result_as_string))
		} else {
			w.Write([]byte(fmt.Sprintf("{\"[errors]\":\"%s\", \"data\":null}", strings.ReplaceAll(fmt.Sprintf("%s", result_as_string_errors), "\"", "\\\""))))
		}
	}

	processRequest := func(w http.ResponseWriter, req *http.Request) {
		var process_request_errors []error
		result := class.Map{}
		if !(req.Method == "POST" || req.Method == "PATCH" || req.Method == "PUT") {
			process_request_errors = append(process_request_errors, fmt.Errorf("http request method not supported: %s", req.Method))
		}

		if len(process_request_errors) > 0 {
			write_response(w, result, process_request_errors)
			return
		}

		body_payload, body_payload_error := ioutil.ReadAll(req.Body);
		if body_payload_error != nil {
			process_request_errors = append(process_request_errors, body_payload_error)
		}

		if body_payload == nil {
			process_request_errors = append(process_request_errors, fmt.Errorf("body payload is nil"))
		}

		if len(process_request_errors) > 0 {
			write_response(w, result, process_request_errors)
			return
		}

		json_payload, json_payload_errors := class.ParseJSON(string(body_payload))

		if json_payload_errors != nil {
			process_request_errors = append(process_request_errors, json_payload_errors...)
		}

		if json_payload == nil {
			process_request_errors = append(process_request_errors, fmt.Errorf("json is nil"))
		}

		if len(process_request_errors) > 0 {
			write_response(w, result, process_request_errors)
			return
		}

		trace_id := get_trace_id()
		json_payload.SetString("[trace_id]", &trace_id)

		json_payload_as_string, payload_as_string_errors := json_payload.ToJSONString()
		if payload_as_string_errors != nil {
			process_request_errors = append(process_request_errors, payload_as_string_errors...)
		}

		if json_payload_as_string == nil {
			process_request_errors = append(process_request_errors, fmt.Errorf("json tostring is nil"))
		}

		if len(process_request_errors) > 0 {
			write_response(w, result, process_request_errors)
			return
		}

		json_bytes := []byte(*json_payload_as_string)
		json_reader := bytes.NewReader(json_bytes)

		request, request_error := http.NewRequest(http.MethodPost, queue_url, json_reader)
		if request_error != nil {
			process_request_errors = append(process_request_errors, request_error)
		}

		if len(process_request_errors) > 0 {
			write_response(w, result, process_request_errors)
			return
		}

		request.Header.Set("Content-Type", "application/json")
		http_response, http_response_error := http_client.Do(request)
		if http_response_error != nil {
			process_request_errors = append(process_request_errors, http_response_error)
		} 

		if len(process_request_errors) > 0 {
			write_response(w, result, process_request_errors)
			return
		}

		response_payload, response_payload_error := ioutil.ReadAll(http_response.Body);
		if response_payload_error != nil {
			process_request_errors = append(process_request_errors, response_payload_error)
		}

		if len(process_request_errors) > 0 {
			write_response(w, result, process_request_errors)
			return
		}

		w.Write([]byte(response_payload))
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

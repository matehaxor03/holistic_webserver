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
	json "github.com/matehaxor03/holistic_json/json"
)

type WebServer struct {
	Start      			func() ([]error)
}

func NewWebServer(port string, server_crt_path string, server_key_path string, queue_domain_name string, queue_port string) (*WebServer, []error) {
	struct_type := "*webserver.WebServer"
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

	domain_name, domain_name_errors := class.NewDomainName(queue_domain_name)
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
	data := json.Map{
		"[fields]": json.Map{},
		"[schema]": json.Map{},
		"[system_fields]": json.Map{
			"[port]":&port,
			"[server_crt_path]":&server_crt_path,
			"[server_key_path]":&server_key_path,
			"[queue_port]":&queue_port,
			"[queue_domain_name]":domain_name,	
		},
		"[system_schema]":json.Map{
			"[port]": json.Map{"type":"string"},
			"[server_crt_path]": json.Map{"type":"string"},
			"[server_key_path]": json.Map{"type":"string"},
			"[queue_port]": json.Map{"type":"string"},
			"[queue_domain_name]": json.Map{"type":"class.DomainName"},
		},
	}

	getData := func() *json.Map {
		return &data
	}

	getPort := func() (string, []error) {
		temp_value, temp_value_errors := class.GetField(struct_type, getData(), "[system_schema]", "[system_fields]", "[port]", "string")
		if temp_value_errors != nil {
			return "",temp_value_errors
		}
		return temp_value.(string), nil
	}

	getServerCrtPath := func() (string, []error) {
		temp_value, temp_value_errors := class.GetField(struct_type, getData(), "[system_schema]", "[system_fields]", "[server_crt_path]", "string")
		if temp_value_errors != nil {
			return "",temp_value_errors
		}
		return temp_value.(string), nil
	}

	getServerKeyPath := func() (string, []error) {
		temp_value, temp_value_errors := class.GetField(struct_type, getData(), "[system_schema]", "[system_fields]", "[server_key_path]", "string")
		if temp_value_errors != nil {
			return "",temp_value_errors
		}
		return temp_value.(string), nil
	}

	getQueuePort := func() (string, []error) {
		temp_value, temp_value_errors := class.GetField(struct_type, getData(), "[system_schema]", "[system_fields]", "[queue_port]", "string")
		if temp_value_errors != nil {
			return "", temp_value_errors
		}
		return temp_value.(string), nil
	}

	getQueueDomainName := func() (*class.DomainName, []error) {
		temp_value, temp_value_errors := class.GetField(struct_type, getData(), "[system_schema]", "[system_fields]", "[queue_domain_name]", "*class.DomainName")
		if temp_value_errors != nil {
			return nil,temp_value_errors
		} else if temp_value == nil {
			return nil, nil
		}
		return temp_value.(*class.DomainName), nil
	}

	queue_domain_name_object, queue_domain_name_object_errors := getQueueDomainName()
	if queue_domain_name_object_errors != nil {
		return nil, queue_domain_name_object_errors
	}

	queue_domain_name_object_value, queue_domain_name_object_value_errors := queue_domain_name_object.GetDomainName()
	if queue_domain_name_object_value_errors != nil {
		return nil, queue_domain_name_object_value_errors
	}

	queue_port_value, queue_port_value_errors := getQueuePort()
	if queue_port_value_errors != nil {
		return nil, queue_port_value_errors
	}

	queue_url := fmt.Sprintf("https://%s:%s/", queue_domain_name_object_value, queue_port_value)

	validate := func() []error {
		return class.ValidateData(getData(), struct_type)
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

	write_response := func(w http.ResponseWriter, result json.Map, write_response_errors []error) {
		if len(write_response_errors) > 0 {
			result.SetNil("data")
			result.SetErrors("[errors]", &write_response_errors)
		}

		var json_payload_builder strings.Builder
		result_as_string_errors := result.ToJSONString(&json_payload_builder)
		if result_as_string_errors != nil {
			write_response_errors = append(write_response_errors, result_as_string_errors...)
		}
		
		if result_as_string_errors == nil {
			w.Write([]byte(json_payload_builder.String()))
		} else {
			w.Write([]byte(fmt.Sprintf("{\"[errors]\":\"%s\", \"data\":null}", strings.ReplaceAll(fmt.Sprintf("%s", result_as_string_errors), "\"", "\\\""))))
		}
	}

	processRequest := func(w http.ResponseWriter, req *http.Request) {
		var process_request_errors []error
		result := json.Map{}
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

		json_payload, json_payload_errors := json.ParseJSON(string(body_payload))

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
		keys := json_payload.Keys()
		if len(keys) != 1 {
			errors = append(errors, fmt.Errorf("root attributes had more than 1"))
		} else {
			json_payload_inner, json_payload_inner_errors := json_payload.GetMap(keys[0])
			if json_payload_inner_errors != nil {
				errors = append(errors, json_payload_inner_errors...)
			} else if json_payload_inner == nil {
				errors = append(errors, fmt.Errorf("json payload is nil"))
			} else {
				json_payload_inner.SetString("[trace_id]", &trace_id)
			}
		}

		var json_payload_builder strings.Builder
		payload_as_string_errors := json_payload.ToJSONString(&json_payload_builder)
		if payload_as_string_errors != nil {
			process_request_errors = append(process_request_errors, payload_as_string_errors...)
		}

		if len(process_request_errors) > 0 {
			write_response(w, result, process_request_errors)
			return
		}

		json_bytes := []byte(json_payload_builder.String())
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

			temp_port, temp_port_errors := getPort()
			if temp_port_errors != nil {
				return temp_port_errors
			}

			temp_server_crt_path, temp_server_crt_path_errors := getServerCrtPath()
			if temp_server_crt_path_errors != nil {
				return temp_server_crt_path_errors
			}

			temp_server_key_path, temp_server_key_path_errors := getServerKeyPath()
			if temp_server_key_path_errors != nil {
				return temp_server_key_path_errors
			}

			err := http.ListenAndServeTLS(":" + temp_port, temp_server_crt_path, temp_server_key_path, nil)
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

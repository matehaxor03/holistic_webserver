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
	data := class.Map{
		"[port]": class.Map{"value": &port, "mandatory": true},
		"[server_crt_path]": class.Map{"value": &server_crt_path, "mandatory": true},
		"[server_key_path]": class.Map{"value": &server_key_path, "mandatory": true},
		"[queue_port]": class.Map{"value": &queue_port, "mandatory": true},
		"[queue_domain_name]": class.Map{"value": domain_name, "mandatory": true},
	}

	getData := func() *class.Map {
		return &data
	}

	getPort := func() (string, []error) {
		temp_port_map, temp_port_map_errors := data.GetMap("[port]")
		if temp_port_map_errors != nil {
			return "", temp_port_map_errors
		}

		temp_port, temp_port_errors := temp_port_map.GetString("value")
		if temp_port_errors != nil {
			return "", temp_port_errors
		}
		return *temp_port, nil
	}

	getServerCrtPath := func() (string, []error) {
		x_map, x_map_errors := data.GetMap("[server_crt_path]")
		if x_map_errors != nil {
			return "", x_map_errors
		}

		temp_x, temp_x_errors := x_map.GetString("value")
		if temp_x_errors != nil {
			return "", temp_x_errors
		}
		return *temp_x, nil
	}

	getServerKeyPath := func() (string, []error) {
		x_map, x_map_errors := data.GetMap("[server_key_path]")
		if x_map_errors != nil {
			return "", x_map_errors
		}

		temp_x, temp_x_errors := x_map.GetString("value")
		if temp_x_errors != nil {
			return "", temp_x_errors
		}
		return *temp_x, nil
	}

	
	getQueuePort := func() (string, []error) {
		temp_port_map, temp_port_map_errors := data.GetMap("[queue_port]")
		if temp_port_map_errors != nil {
			return "", temp_port_map_errors
		}

		temp_port, temp_port_errors := temp_port_map.GetString("value")
		if temp_port_errors != nil {
			return "", temp_port_errors
		}
		return *temp_port, nil
	}

	getQueueDomainName := func() (*class.DomainName, []error) {
		temp_queue_domain_name, temp_queue_domain_name_errors := data.GetMap("[queue_domain_name]")
		if temp_queue_domain_name_errors != nil {
			return nil, temp_queue_domain_name_errors
		}

		temp_queue_domain_name_obj := temp_queue_domain_name.GetObject("value").(*class.DomainName)
		return temp_queue_domain_name_obj, nil
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
		return class.ValidateData(getData(), "WebServer")
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

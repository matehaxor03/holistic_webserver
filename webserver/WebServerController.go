package queue

import (
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
	"time"
	"bytes"
	"crypto/tls"
	"sync"
	"crypto/rand"
	dao "github.com/matehaxor03/holistic_db_client/dao"
	common "github.com/matehaxor03/holistic_common/common"
	json "github.com/matehaxor03/holistic_json/json"
	http_extension "github.com/matehaxor03/holistic_http/http_extension"
	validate "github.com/matehaxor03/holistic_db_client/validate"

)

type WebServerController struct {
	GetProcessRequestFunction func() *func(w http.ResponseWriter, req *http.Request)
	SetQueuePushBackFunction func(new_push_back_function *func(*json.Map) (*json.Map, []error))
}

func NewWebServerController(queue_name string, queue_domain_name string, queue_port string) (*WebServerController, []error) {
	verfiy := validate.NewValidator()
	var errors []error
	var trace_id_lock sync.Mutex
	var messageCount uint64
	var queue_push_back_function *func(*json.Map) (*json.Map, []error)


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

	_, domain_name_errors := dao.NewDomainName(verfiy, queue_domain_name)
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


	getQueueName := func() string {
		return queue_name
	}

	getQueuePort := func() string {
		return queue_port
	}

	getQueueDomainName := func() string {
		return queue_domain_name
	}


	queue_url := fmt.Sprintf("https://%s:%s/queue_api/" + getQueueName(), getQueueDomainName(), getQueuePort())

	validate := func() []error {
		return nil
	}

	process_request_function := func(w http.ResponseWriter, req *http.Request) {
		var process_request_errors []error
		var response_payload_result string
		result := json.NewMap()
		if !(req.Method == "POST" || req.Method == "PATCH" || req.Method == "PUT") {
			process_request_errors = append(process_request_errors, fmt.Errorf("http request method not supported: %s", req.Method))
		}

		if len(process_request_errors) > 0 {
			http_extension.WriteResponse(w, result, process_request_errors)
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
			http_extension.WriteResponse(w, result, process_request_errors)
			return
		}

		json_payload, json_payload_errors := json.Parse(string(body_payload))

		if json_payload_errors != nil {
			process_request_errors = append(process_request_errors, json_payload_errors...)
		}

		if json_payload == nil {
			process_request_errors = append(process_request_errors, fmt.Errorf("json is nil"))
		}

		if len(process_request_errors) > 0 {
			http_extension.WriteResponse(w, result, process_request_errors)
			return
		}

		trace_id := get_trace_id()
		json_payload.SetString("[trace_id]", &trace_id)


		if queue_push_back_function != nil {
			queue, queue_errors := json_payload.GetString("[queue]")
			if queue_errors != nil {
				process_request_errors = append(process_request_errors, queue_errors...)
			} else if common.IsNil(queue) {
				process_request_errors = append(process_request_errors, fmt.Errorf("queue is nil"))
			}

			queue_mode, queue_mode_errors := json_payload.GetString("[queue_mode]")
			if queue_mode_errors != nil {
				process_request_errors = append(process_request_errors, queue_mode_errors...)
			} else if common.IsNil(queue_mode) {
				temp_queue_mode := "PushBack"
				queue_mode = &temp_queue_mode
				json_payload.SetString("[queue_mode]", queue_mode)
			}

			async, async_errors := json_payload.GetBool("[async]")
			if async_errors != nil {
				process_request_errors = append(process_request_errors, async_errors...)
			} else if common.IsNil(async) {
				temp_async := false
				async = &temp_async
				json_payload.SetBool("[async]", async)
			}

			if len(process_request_errors) == 0 {
				function_response_payload, function_response_payload_errors := (*queue_push_back_function)(json_payload)
				if function_response_payload_errors != nil {
					process_request_errors = append(process_request_errors, function_response_payload_errors...)
				} else if common.IsNil(function_response_payload) {
					process_request_errors = append(process_request_errors, fmt.Errorf("function_response_payload is nil"))
				} else {
					var function_builder strings.Builder 
					to_json_string_errors := function_response_payload.ToJSONString(&function_builder)
					if to_json_string_errors != nil {
						process_request_errors = append(process_request_errors, to_json_string_errors...)
					} else {
						response_payload_result = function_builder.String()
					}
				}
			}
		} else {
			var json_payload_builder strings.Builder
			payload_as_string_errors := json_payload.ToJSONString(&json_payload_builder)
			if payload_as_string_errors != nil {
				process_request_errors = append(process_request_errors, payload_as_string_errors...)
			}

			if len(process_request_errors) > 0 {
				http_extension.WriteResponse(w, result, process_request_errors)
				return
			}

			json_bytes := []byte(json_payload_builder.String())
			json_reader := bytes.NewReader(json_bytes)

			request, request_error := http.NewRequest(http.MethodPost, queue_url, json_reader)
			if request_error != nil {
				process_request_errors = append(process_request_errors, request_error)
			}

			if len(process_request_errors) > 0 {
				http_extension.WriteResponse(w, result, process_request_errors)
				return
			}

			request.Header.Set("Content-Type", "application/json")
			http_response, http_response_error := http_client.Do(request)
			if http_response_error != nil {
				process_request_errors = append(process_request_errors, http_response_error)
			} 

			if len(process_request_errors) > 0 {
				http_extension.WriteResponse(w, result, process_request_errors)
				return
			}

			response_payload, response_payload_error := ioutil.ReadAll(http_response.Body);
			if response_payload_error != nil {
				process_request_errors = append(process_request_errors, response_payload_error)
			} else if common.IsNil(response_payload) {
				process_request_errors = append(process_request_errors, fmt.Errorf("response_payload is nil"))
			} else {
				response_payload_result = string(response_payload)
			}
		}

		if len(process_request_errors) > 0 {
			http_extension.WriteResponse(w, result, process_request_errors)
		} else {
			w.Write([]byte(response_payload_result))
		}
	}

	x := WebServerController{
		GetProcessRequestFunction: func() *func(w http.ResponseWriter, req *http.Request) {
			function := process_request_function
			return &function
		},
		SetQueuePushBackFunction: func(new_push_back_function *func(*json.Map) (*json.Map, []error)) {
			queue_push_back_function = new_push_back_function
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

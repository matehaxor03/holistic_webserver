package queue

import (
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
	class "github.com/matehaxor03/holistic_db_client/class"
)

type WebServer struct {
	Start      			func() ([]error)
}

func NewWebServer(port string, server_crt_path string, server_key_path string) (*WebServer, []error) {
	var errors []error
	//var this_holisic_queue_server *HolisticQueueServer

	//todo: add filters to fields
	data := class.Map{
		"[port]": class.Map{"value": class.CloneString(&port), "mandatory": true},
		"[server_crt_path]": class.Map{"value": class.CloneString(&server_crt_path), "mandatory": true},
		"[server_key_path]": class.Map{"value": class.CloneString(&server_key_path), "mandatory": true},
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
			
			ioutil.ReadAll(req.Body);
			// todo call queue
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

package main

import (
	webserver "github.com/matehaxor03/holistic_webserver/webserver"
	"os"
	"fmt"
)

func main() {
	var errors []error
	web_server, web_server_errors := webserver.NewWebServer("5001", "server.crt", "server.key")
	if web_server_errors != nil {
		errors = append(errors, web_server_errors...)	
	} else {
		web_server_start_errors := web_server.Start()
		if web_server_start_errors != nil {
			errors = append(errors, web_server_start_errors...)
		}
	}

	if len(errors) > 0 {
		fmt.Println(errors)
		os.Exit(1)
	}
	
	os.Exit(0)
}

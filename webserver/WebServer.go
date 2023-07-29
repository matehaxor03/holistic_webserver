package queue

import (
	"fmt"
	"sync"
	dao "github.com/matehaxor03/holistic_db_client/dao"
	common "github.com/matehaxor03/holistic_common/common"
	validate "github.com/matehaxor03/holistic_validator/validate"
	"net/http"
)

type WebServer struct {
	Start      			func() ([]error)
}

func NewWebServer(port string, server_crt_path string, server_key_path string, queue_domain_name string, queue_port string) (*WebServer, []error) {
	verfiy := validate.NewValidator()
	get_controller_by_name_lock := &sync.RWMutex{}
	var errors []error

	client_manager, client_manager_errors := dao.NewClientManager()
	if client_manager_errors != nil {
		return nil, client_manager_errors
	}

	test_read_client, test_read_client_errors := client_manager.GetClient("127.0.0.1", "3306", "holistic", "holistic_r")
	if test_read_client_errors != nil {
		return nil, test_read_client_errors
	}

	getQueuePort := func() string {
		return queue_port
	}

	getQueueDomainName := func() string {
		return queue_domain_name
	}
	
	test_read_database := test_read_client.GetDatabase()

	controllers := make(map[string](*WebServerController))
	table_names, table_names_errors := test_read_database.GetTableNames()
	if table_names_errors != nil {
		return nil, table_names_errors
	}

	controllers["Run_Sync"], _ = NewWebServerController("Run_Sync", queue_domain_name, getQueuePort())

	for _, table_name := range table_names {
		controllers["CreateRecords_"+table_name], _ = NewWebServerController("CreateRecords_"+table_name, getQueueDomainName(), getQueuePort())
		controllers["CreateRecord_"+table_name], _ = NewWebServerController("CreateRecord_"+table_name, getQueueDomainName(), getQueuePort())
		controllers["ReadRecords_"+table_name], _ = NewWebServerController("ReadRecords_"+table_name, getQueueDomainName(), getQueuePort())
		controllers["UpdateRecords_"+table_name], _ = NewWebServerController("UpdateRecords_"+table_name, getQueueDomainName(), getQueuePort())
		controllers["UpdateRecord_"+table_name], _ = NewWebServerController("UpdateRecord_"+table_name, getQueueDomainName(), getQueuePort())
		controllers["CreateRecords_"+table_name], _ = NewWebServerController("CreateRecords_"+table_name, getQueueDomainName(), getQueuePort())
		controllers["DeleteRecords_"+table_name], _ = NewWebServerController("DeleteRecords_"+table_name, getQueueDomainName(), getQueuePort())
		controllers["GetSchema_"+table_name], _ = NewWebServerController("GetSchema_"+table_name, getQueueDomainName(), getQueuePort())
	}

	controllers["Run_StartBranchInstance"], _ = NewWebServerController("Run_StartBranchInstance", getQueueDomainName(), getQueuePort())
	controllers["Run_NotStarted"], _ = NewWebServerController("Run_NotStarted", getQueueDomainName(), getQueuePort())
	controllers["Run_Start"], _ = NewWebServerController("Run_Start", getQueueDomainName(), getQueuePort())
	controllers["Run_CreateSourceFolder"], _ = NewWebServerController("Run_CreateSourceFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_CreateDomainNameFolder"], _ = NewWebServerController("Run_CreateDomainNameFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_CreateRepositoryAccountFolder"], _ = NewWebServerController("Run_CreateRepositoryAccountFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_CreateRepositoryFolder"], _ = NewWebServerController("Run_CreateRepositoryFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_CreateBranchesFolder"], _ = NewWebServerController("Run_CreateBranchesFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_CreateTagsFolder"], _ = NewWebServerController("Run_CreateTagsFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_CreateBranchInstancesFolder"], _ = NewWebServerController("Run_CreateBranchInstancesFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_CreateTagInstancesFolder"], _ = NewWebServerController("Run_CreateTagInstancesFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_CreateBranchOrTagFolder"], _ = NewWebServerController("Run_CreateBranchOrTagFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_CloneBranchOrTagFolder"], _ = NewWebServerController("Run_CloneBranchOrTagFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_PullLatestBranchOrTagFolder"], _ = NewWebServerController("Run_PullLatestBranchOrTagFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_CreateInstanceFolder"], _ = NewWebServerController("Run_CreateInstanceFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_CopyToInstanceFolder"], _ = NewWebServerController("Run_CopyToInstanceFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_CreateGroup"], _ = NewWebServerController("Run_CreateGroup", getQueueDomainName(), getQueuePort())
	controllers["Run_CreateUser"], _ = NewWebServerController("Run_CreateUser", getQueueDomainName(), getQueuePort())
	controllers["Run_AssignGroupToUser"], _ = NewWebServerController("Run_AssignGroupToUser", getQueueDomainName(), getQueuePort())
	controllers["Run_AssignGroupToInstanceFolder"], _ = NewWebServerController("Run_AssignGroupToInstanceFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_Clean"], _ = NewWebServerController("Run_Clean", getQueueDomainName(), getQueuePort())
	controllers["Run_Lint"], _ = NewWebServerController("Run_Lint", getQueueDomainName(), getQueuePort())
	controllers["Run_Build"], _ = NewWebServerController("Run_Build", getQueueDomainName(), getQueuePort())
	controllers["Run_UnitTests"], _ = NewWebServerController("Run_UnitTests", getQueueDomainName(), getQueuePort())
	controllers["Run_IntegrationTests"], _ = NewWebServerController("Run_IntegrationTests", getQueueDomainName(), getQueuePort())
	controllers["Run_IntegrationTestSuite"], _ = NewWebServerController("Run_IntegrationTestSuite", getQueueDomainName(), getQueuePort())
	controllers["Run_RemoveGroupFromInstanceFolder"], _ = NewWebServerController("Run_RemoveGroupFromInstanceFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_RemoveGroupFromUser"], _ = NewWebServerController("Run_RemoveGroupFromUser", getQueueDomainName(), getQueuePort())
	controllers["Run_DeleteGroup"], _ = NewWebServerController("Run_DeleteGroup", getQueueDomainName(), getQueuePort())
	controllers["Run_DeleteUser"], _ = NewWebServerController("Run_DeleteUser", getQueueDomainName(), getQueuePort())
	controllers["Run_DeleteInstanceFolder"], _ = NewWebServerController("Run_DeleteInstanceFolder", getQueueDomainName(), getQueuePort())
	controllers["Run_End"], _ = NewWebServerController("Run_End", getQueueDomainName(), getQueuePort())

	controllers["GetTableNames"], _ = NewWebServerController("GetTableNames", getQueueDomainName(), getQueuePort())

	_, domain_name_errors := dao.NewDomainName(verfiy, queue_domain_name)
	if domain_name_errors != nil {
		errors = append(errors, domain_name_errors...)
	}

	get_controller_by_name := func(contoller_name string) (*WebServerController, error) {
		get_controller_by_name_lock.Lock()
		defer get_controller_by_name_lock.Unlock()
		queue_obj, queue_found := controllers[contoller_name]
		if !queue_found {
			return nil, fmt.Errorf("queue not found %s", contoller_name)
		} else if common.IsNil(queue_obj) {
			return nil, fmt.Errorf("queue found but is nil %s", contoller_name)
		} 
		return queue_obj, nil
	}
	
	get_controller_names := func() []string {
		get_controller_by_name_lock.Lock()
		defer get_controller_by_name_lock.Unlock()
		controller_names := make([]string, len(controllers))
		for key, _ := range controllers {
			controller_names = append(controller_names, key)
		}
		return controller_names
	}
	//var this_holisic_queue_server *HolisticQueueServer

	//todo: add filters to fields

	getPort := func() string {
		return port
	}

	getServerCrtPath := func() string {
		return server_crt_path
	}

	getServerKeyPath := func() string {
		return server_key_path
	}

	

	validate := func() []error {
		return nil
	}

	x := WebServer{
		Start: func() []error {
			buildHandler := http.FileServer(http.Dir("static"))
			http.Handle("/", buildHandler)

			var start_server_errors []error

			temp_controller_names := get_controller_names()
			//http.HandleFunc("/queue_api", processRequest)

			for _, temp_controller_name := range temp_controller_names {
				temp_controller, temp_controller_errors := get_controller_by_name(temp_controller_name)
				if temp_controller_errors != nil {
					start_server_errors = append(start_server_errors, temp_controller_errors)
					continue
				} else if common.IsNil(temp_controller) {
					start_server_errors = append(start_server_errors, fmt.Errorf("controller is nil: %s", temp_controller))
					continue
				}
				http.HandleFunc("/webserver_api/" + temp_controller_name, *(temp_controller.GetProcessRequestFunction()))
			}

			err := http.ListenAndServeTLS(":"+ getPort(),  getServerCrtPath(), getServerKeyPath(), nil)
			if err != nil {
				start_server_errors = append(start_server_errors, err)
			}

			if len(start_server_errors) > 0 {
				return start_server_errors
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

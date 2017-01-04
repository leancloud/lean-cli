package console

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ahmetalpbalkan/go-linq"
	"github.com/aisk/chrysanthemum"
	"github.com/gorilla/mux"
	"github.com/levigross/grequests"
)

var hookNames = map[string]string{
	"__before_save_for_":   "beforeSave",
	"__after_save_for_":    "afterSave",
	"__before_update_for_": "beforeUpdate",
	"__after_update_for_":  "afterUpdate",
	"__before_delete_for_": "beforeDelete",
	"__after_delete_for_":  "afterDelete",
	"__on_login_":          "onLogin",
}

// Server is a struct for develoment console server
type Server struct {
	AppID       string
	AppKey      string
	MasterKey   string
	AppPort     string
	ConsolePort string
	Errors      chan error
}

func (server *Server) getFunctions() ([]string, error) {
	url := fmt.Sprintf("http://localhost:%s/1.1/functions/_ops/metadatas", server.AppPort)
	response, err := grequests.Get(url, &grequests.RequestOptions{
		Headers: map[string]string{
			"x-avoscloud-application-id": server.AppID,
			"x-avoscloud-master-key":     server.MasterKey,
		},
	})
	if err != nil {
		return nil, err
	}

	result := new(struct {
		Result []string `json:"result"`
	})
	err = response.JSON(result)
	if err != nil {
		return nil, err
	}
	return result.Result, nil
}

func (server *Server) indexHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, resources["index.html"])
}

func (server *Server) resourcesHandler(w http.ResponseWriter, req *http.Request) {
	resourceName := mux.Vars(req)["resourceName"]
	if resource, ok := resources[resourceName]; ok {
		fmt.Fprintf(w, resource)
	} else {
		http.NotFound(w, req)
	}
}

func (server *Server) appInfoHandler(w http.ResponseWriter, req *http.Request) {
	port, err := strconv.Atoi(server.AppPort)
	if err != nil {
		panic(err)
	}
	content, err := json.Marshal(map[string]interface{}{
		"appId":          server.AppID,
		"appKey":         server.AppKey,
		"masterKey":      server.MasterKey,
		"leanenginePort": port,
	})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(content)
}

func (server *Server) functionsHandler(w http.ResponseWriter, req *http.Request) {
	functions, err := server.getFunctions()
	if err != nil {
		fmt.Println("get functions error: ", err)
		return
	}

	result := linq.From(functions).Where(func(in interface{}) bool {
		function := in.(string)
		return !strings.HasPrefix(function, "__")
	}).Results()
	if len(result) > 0 {
		result = linq.From(result).OrderBy(func(in interface{}) interface{} {
			function := in.(string)
			return function[0]
		}).Select(func(in interface{}) interface{} {
			function := in.(string)
			return map[string]string{
				"name": function,
				"sign": signCloudFunc(server.MasterKey, function, timeStamp()),
			}
		}).Results()
	}

	w.Header().Set("Content-Type", "application/json")
	j, _ := json.MarshalIndent(result, "", "  ")
	w.Write(j)
}

func (server *Server) classesHandler(w http.ResponseWriter, req *http.Request) {
	functions, err := server.getFunctions()
	if err != nil {
		fmt.Println("get functions error: ", err)
		return
	}

	result := linq.From(functions).Where(func(in interface{}) bool {
		funcName := in.(string)
		for key := range hookNames {
			if strings.HasPrefix(funcName, key) {
				return true
			}
		}
		return false
	}).Select(func(in interface{}) interface{} {
		funcName := in.(string)
		for key := range hookNames {
			if strings.HasPrefix(funcName, key) {
				return strings.TrimPrefix(funcName, key)
			}
		}
		panic("impossible")
	}).Results()

	if len(result) > 0 {
		result = linq.From(result).OrderBy(func(in interface{}) interface{} {
			function := in.(string)
			return function[0]
		}).Results()
	}

	w.Header().Set("Content-Type", "application/json")
	j, _ := json.MarshalIndent(result, "", "  ")
	w.Write(j)
}

func (server *Server) classActionHandler(w http.ResponseWriter, req *http.Request) {
	className := mux.Vars(req)["className"]

	functions, err := server.getFunctions()
	if err != nil {
		fmt.Println("get functions error: ", err)
		return
	}

	result := linq.From(functions).Where(func(in interface{}) bool {
		funcName := in.(string)
		if strings.HasPrefix(funcName, "__") && strings.HasSuffix(funcName, className) {
			return true
		}
		return false
	}).Select(func(in interface{}) interface{} {
		funcName := in.(string)
		action := ""
		for key, value := range hookNames {
			if strings.HasPrefix(funcName, key) {
				action = value
			}
		}
		signFuncName := ""
		if strings.HasPrefix(funcName, "__before") {
			signFuncName = "__before_for_" + className
		} else if strings.HasPrefix(funcName, "__after") {
			signFuncName = "__after_for_" + className
		}
		return map[string]string{
			"className": className,
			"action":    action,
			"sign":      signCloudFunc(server.MasterKey, signFuncName, timeStamp()),
		}
	}).Results()

	w.Header().Set("Content-Type", "application/json")
	j, _ := json.MarshalIndent(result, "", "  ")
	w.Write(j)

}

// Run the dev server
func (server *Server) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/", server.indexHandler)
	router.HandleFunc("/{resourceName}", server.resourcesHandler)
	router.HandleFunc("/__engine/1/appInfo", server.appInfoHandler)
	router.HandleFunc("/__engine/1/functions", server.functionsHandler)
	router.HandleFunc("/__engine/1/classes", server.classesHandler)
	router.HandleFunc("/__engine/1/classes/{className}/actions", server.classActionHandler)

	addr := "localhost:" + server.ConsolePort
	chrysanthemum.Println("云函数调试服务已启动，请使用浏览器访问：http://" + addr)

	go func() {
		server.Errors <- http.ListenAndServe(addr, router)
	}()
}

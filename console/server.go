package console

import (
	"embed"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ahmetalpbalkan/go-linq"
	"github.com/aisk/logp"
	"github.com/gorilla/mux"
	"github.com/leancloud/lean-cli/api"
	"github.com/levigross/grequests"
)

var hookNames = map[string]string{
	"__before_save_for_":   "beforeSave",
	"__after_save_for_":    "afterSave",
	"__before_update_for_": "beforeUpdate",
	"__after_update_for_":  "afterUpdate",
	"__before_delete_for_": "beforeDelete",
	"__after_delete_for_":  "afterDelete",
}
var specialHookNames = map[string]string{
	"__on_authdata_":      "onAuthData",
	"__on_login_":         "onLogin",
	"__on_verified_sms":   "onVerifiedSms",
	"__on_verified_email": "onVerifiedEmail",
}

// Server is a struct for develoment console server
type Server struct {
	AppID       string
	AppKey      string
	MasterKey   string
	HookKey     string
	RemoteURL   string
	ConsolePort string
	Errors      chan error
}

//go:embed resources
var resources embed.FS

func init() {
	// for Windows compatibility
	mime.AddExtensionType(".js", "text/javascript; charset=utf-8")
	mime.AddExtensionType(".css", "text/css; charset=utf-8")
}

func (server *Server) getFunctions() ([]string, error) {
	url := fmt.Sprintf("%s/1.1/functions/_ops/metadatas", server.RemoteURL)
	response, err := grequests.Get(url, &grequests.RequestOptions{
		Headers: map[string]string{
			"x-avoscloud-application-id": server.AppID,
			"x-avoscloud-master-key":     server.MasterKey,
		},
	})
	if err != nil {
		return nil, err
	}

	if !response.Ok {
		return nil, api.NewErrorFromResponse(response)
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
	bytes, err := resources.ReadFile("resources/index.html")

	if err != nil {
		panic(err)
	}

	w.Write(bytes)
}

func (server *Server) resourcesHandler(w http.ResponseWriter, req *http.Request) {
	filename := mux.Vars(req)["filename"]

	bytes, err := resources.ReadFile(path.Join("resources", filename))

	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, req)
		} else {
			panic(err)
		}
	}

	w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(filename)))
	w.Write(bytes)
}

func (server *Server) appInfoHandler(w http.ResponseWriter, req *http.Request) {
	url := fmt.Sprintf("%s/1.1/functions/_ops/metadatas", server.RemoteURL)
	response, err := grequests.Options(url, &grequests.RequestOptions{
		Headers: map[string]string{
			"Access-Control-Request-Method": "GET",
			"Origin":                        fmt.Sprint("http://localhost:", server.ConsolePort),
		},
	})
	if err != nil {
		panic(err)
	}
	if !response.Ok {
		panic(api.NewErrorFromResponse(response))
	}

	content, err := json.Marshal(map[string]interface{}{
		"appId":       server.AppID,
		"appKey":      server.AppKey,
		"masterKey":   server.MasterKey,
		"hookKey":     server.HookKey,
		"sendHookKey": strings.Contains(response.Header.Get("Access-Control-Allow-Headers"), "X-LC-Hook-Key"),
		"remoteUrl":   server.RemoteURL,
		"warnings":    []string{},
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
			if function == "" {
				return " "[0]
			}
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
	j, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
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
	}).Distinct().Results()

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
		signFuncName := funcName
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

func (server *Server) specialHooksHandler(w http.ResponseWriter, req *http.Request) {
	functions, err := server.getFunctions()
	if err != nil {
		fmt.Println("get functions error: ", err)
		return
	}

	result := linq.From(functions).Where(func(in interface{}) bool {
		funcName := in.(string)
		for key := range specialHookNames {
			if strings.HasPrefix(funcName, key) {
				return true
			}
		}
		return false
	}).Select(func(in interface{}) interface{} {
		funcName := in.(string)
		action := ""
		for key, value := range specialHookNames {
			if strings.HasPrefix(funcName, key) {
				action = value
			}
		}
		return map[string]string{
			"className": "_User",
			"action":    action,
			"sign":      signCloudFunc(server.MasterKey, funcName, timeStamp()),
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
	router.HandleFunc("/__engine/1/appInfo", server.appInfoHandler)
	router.HandleFunc("/__engine/1/functions", server.functionsHandler)
	router.HandleFunc("/__engine/1/classes", server.classesHandler)
	router.HandleFunc("/__engine/1/classes/{className}/actions", server.classActionHandler)
	router.HandleFunc("/__engine/1/special-hooks", server.specialHooksHandler)
	router.HandleFunc("/resources/{filename}", server.resourcesHandler)

	addr := "localhost:" + server.ConsolePort
	logp.Info("Cloud function debug console (if available) is accessible at: http://" + addr)

	go func() {
		server.Errors <- http.ListenAndServe(addr, router)
	}()
}

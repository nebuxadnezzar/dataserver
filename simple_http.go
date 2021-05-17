package main

import (
	"bytes"
	du "dataserver/util"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/yuin/gopher-lua"
)

const DEFAULT_ADDRESS = ":8080"
const DEFAULT_SCRIPT = `sendData('{"status":"ok", "handler":null}')`
const FAILED_TEMPLATE = `{"status":"fail", "message":"%s"}`

var FORM_HEADERS [2]string = [...]string{"application/x-www-form-urlencoded", "multipart/form-data"}
var SUPPORTED_METHODS [3]string = [...]string{"GET", "POST", "HEAD"}

const BODY_PARAM = "_body_"
const MAX_BODY_SIZE = 1024

type requestHandler struct {
	method  func(http.ResponseWriter, *http.Request)
	pattern string
}

var sendData func(http.ResponseWriter, []byte)
var handlerResolver func(string) string
var cgiResolver func(string) string

//-----------------------------------------------------------------------------
func createSendDataFunc(headers map[string]string) func(http.ResponseWriter, []byte) {

	hdrs := headers
	return func(w http.ResponseWriter, bytes []byte) {

		if hdrs != nil {
			for k, v := range hdrs {
				w.Header().Set(k, v)
			}
		}
		w.Write(bytes)
	}
}

//-----------------------------------------------------------------------------
func createSendDataFromLuaFunc(w http.ResponseWriter, f func(http.ResponseWriter, []byte)) func(L *lua.LState) int {
	writer := w
	dataFunc := f
	return func(L *lua.LState) int {

		s := L.ToString(1)
		l := len(s)
		//fmt.Printf( "\nSTRING FROM LUA %s\n", s )
		//fmt.Printf( "BYTES FROM LUA %v\n", []byte(s) )
		L.Push(lua.LNumber(l))
		dataFunc(writer, []byte(s))
		return l
	}
}

//-----------------------------------------------------------------------------
func createHandlerResolver(handlerMap map[string]string) func(string) string {

	m := make(map[string]string)

	for pattern, path := range handlerMap {

		if data, err := ioutil.ReadFile(path); err == nil {
			m[pattern] = string(data)
		}
	}

	return func(path string) string {

		var handlerCode string = ""

		for pattern, data := range m {

			if matched, err := regexp.MatchString(pattern, path); err == nil && matched {
				handlerCode = data
				break
			} else if err != nil {
				log.Printf("Broken pattern %s %v\n", pattern, err.Error())
			}
		}

		return handlerCode
	}
}

//-----------------------------------------------------------------------------
func createCgiResolver(cgiMap map[string]string) func(string) string {

	m := make(map[string]string)

	for pattern, path := range cgiMap {
		m[pattern] = path
	}

	return func(url string) string {

		var cgiScript string = ""

		for pattern, path := range m {

			if matched, err := regexp.MatchString(pattern, url); err == nil && matched {
				cgiScript = path
				break
			} else if err != nil {
				log.Printf("Broken pattern %s %v\n", pattern, err.Error())
			}
		}

		return cgiScript
	}
}

//-----------------------------------------------------------------------------
func sendError(w http.ResponseWriter, code int, msg string) {

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintln(w, fmt.Sprintf(FAILED_TEMPLATE, msg))
}

//-----------------------------------------------------------------------------
func collectParams(r *http.Request) (map[string][]string, error) {

	err := &du.Error{}

	if !du.ItemExists(SUPPORTED_METHODS, r.Method) {
		err.Set(http.StatusMethodNotAllowed, fmt.Sprintf("Usupported method %s", r.Method))
		return nil, err
	}

	if du.ItemExists(FORM_HEADERS, r.Header.Get("Content-Type")) {
		e := r.ParseForm()
		if e != nil {
			err.Set(http.StatusUnsupportedMediaType, e.Error())
			return nil, err
		}
	}
	m := map[string][]string(r.URL.Query())

	for k, v := range r.Form {
		m[k] = v
	}
	fmt.Printf("\nPARAM MAP: %v\n", m)
	return m, nil
}

//-----------------------------------------------------------------------------
func cgi(w http.ResponseWriter, r *http.Request) {
	fmt.Println("cgi called")
	m, err := collectParams(r)

	if err != nil {
		sendError(w, err.(*du.Error).Code(), err.Error())
		return
	}

	script := cgiResolver(r.URL.Path)
	w.Header().Set("Content-type", "text/plain")
	if err := du.RunCgi(w, script, du.CreateKeyValuePairs(m, ` `, false)); err != nil {
		sendError(w, 500, err.Error())
		return
	}

	//sendData(w, []byte(fmt.Sprintf("Hi From CGI %v \n %s", m, script)))
}

//-----------------------------------------------------------------------------
func worker(w http.ResponseWriter, r *http.Request) {
	fmt.Println("worker called")
	m, err := collectParams(r)

	if err != nil {
		sendError(w, err.(*du.Error).Code(), err.Error())
		return
	}

	L := lua.NewState()
	defer L.Close()
	luaTbl := L.NewTable()

	buff := make([]byte, MAX_BODY_SIZE)
	n, err := r.Body.Read(buff)
	fmt.Printf("REQUEST BODY: %v %v %v\n", n, err, string(buff))

	if n > 0 && err.Error() == "EOF" {
		luaTbl.RawSetH(lua.LString(BODY_PARAM), lua.LString(string(buff[:n])))
	} else {
		fmt.Printf("ERROR reading request body: %v\n", err.Error())
	}

	for k, v := range m {
		luaTbl.RawSetH(lua.LString(k), lua.LString(strings.Join(v, ",")))
	}

	L.SetGlobal("requestParams", luaTbl)

	t := time.Now().Local()
	fmt.Printf("Handler %s %d%d", t.Format("20060102150405"), t.Year(), t.Month)

	dataFunc := createSendDataFromLuaFunc(w, sendData)
	L.SetGlobal("sendData", L.NewFunction(dataFunc)) // Register our function in Lua
	script := handlerResolver(r.URL.Path)

	if script == "" {
		sendError(w, http.StatusNotFound, "404 not found")
		return
	}
	if err := L.DoString(script); err != nil {
		sendData(w, []byte(fmt.Sprintf("Error executing lua script\n\n%s\n\n%s\n", err.Error())))
	}
}

//-----------------------------------------------------------------------------
func echo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("echo called")
	paramMap, err := collectParams(r)

	if err != nil {
		sendError(w, err.(*du.Error).Code(), err.Error())
		return
	}

	var buf bytes.Buffer

	for k, v := range paramMap {
		buf.WriteString(fmt.Sprintf("key: %s\tval: %s\n", k, strings.Join(v, ", ")))
	}
	t := time.Now().Local()
	buf.WriteString(fmt.Sprintf("Timestamp: %s %d %d", t.Format("2006-01-02 15:04:05"), t.Year(), t.Month))

	sendData(w, buf.Bytes())
}

//-----------------------------------------------------------------------------
func createSvr(addr string, keepalive bool, handlers []requestHandler) error {

	http.DefaultTransport.(*http.Transport).MaxIdleConns = 2
	http.DefaultTransport.(*http.Transport).IdleConnTimeout = 2 * time.Second
	mx := http.NewServeMux()
	// register handler for the pattern
	//
	for _, v := range handlers {
		mx.HandleFunc(v.pattern, v.method)
	}
	// fmt.Printf("MUX %v\n", mx)
	server := &http.Server{Addr: addr, Handler: mx}
	server.SetKeepAlivesEnabled(keepalive)
	finalCleanup(server)
	return server.ListenAndServe()
}

//-----------------------------------------------------------------------------
func finalCleanup(svr *http.Server) {

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nReceived an interrupt, stopping server...\n")
		svr.Close()
	}()
}

//-----------------------------------------------------------------------------
func createConfigMap(fileName string) map[string]map[string]string {

	records := du.ReadFileLines(fileName)
	configMap := du.ParseIniFile(records)
	return configMap
}

//-----------------------------------------------------------------------------
func main() {

	if len(os.Args) < 2 {
		println("path to config.ini file is missing")
		os.Exit(-1)
	}
	args := os.Args[1:]

	configMap := createConfigMap(args[0])

	fmt.Printf("%v\n", configMap)

	keepAlive := true
	address := DEFAULT_ADDRESS
	sendData = createSendDataFunc(configMap["headers"])
	handlerResolver = createHandlerResolver(configMap["handlers"])
	cgiResolver = createCgiResolver(configMap["cgi"])

	if b, err := strconv.ParseBool(configMap["interface"]["keepalive"]); err == nil {
		keepAlive = b
	}

	if addr := configMap["interface"]["address"]; addr != "" {
		address = addr
	}
	rh := []requestHandler{
		{pattern: "/echo", method: echo},
		{pattern: "/cgi/", method: cgi},
		{pattern: "/", method: worker}}
	err := createSvr(address, keepAlive, rh)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

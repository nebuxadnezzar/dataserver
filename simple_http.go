package main

import (
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

type requestHandler struct {
	method  func(http.ResponseWriter, *http.Request)
	pattern string
}

var sendData func(http.ResponseWriter, []byte)
var handlerResolver func(string) string

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

		if handlerCode == "" {
			handlerCode = DEFAULT_SCRIPT
		}
		return handlerCode
	}
}

//-----------------------------------------------------------------------------
func worker(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	//fmt.Printf("FORM: %v\n",r.Form) // print form information in server side
	fmt.Printf("path: %s scheme: %s\n", r.URL.Path, r.URL.Scheme)
	//fmt.Printf("URL: %v\n", r.URL.Query())

	fmt.Println("query params:")
	for k, v := range r.URL.Query() {
		fmt.Printf("\t%s ==> %s\n", k, strings.Join(v, ", "))
	}

	L := lua.NewState()
	defer L.Close()
	luaTbl := L.NewTable()

	for k, v := range r.Form {
		log.Printf("key: %s\tval: %s\n", k, strings.Join(v, ", "))
		luaTbl.RawSetH(lua.LString(k), lua.LString(strings.Join(v, ",")))
	}
	L.SetGlobal("requestParams", luaTbl)

	t := time.Now().Local()
	fmt.Printf("Handler %s %d%d", t.Format("20060102150405"), t.Year(), t.Month)

	dataFunc := createSendDataFromLuaFunc(w, sendData)
	L.SetGlobal("sendData", L.NewFunction(dataFunc)) // Register our function in Lua
	script := handlerResolver(r.URL.Path)

	if err := L.DoString(script); err != nil {
		sendData(w, []byte(fmt.Sprintf("Error executing lua script\n\n%s\n\n%s\n", err.Error())))
	}
	//dataFunc( L )
}

//-----------------------------------------------------------------------------
func echo(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	fmt.Printf("FORM: %v\n", r.Form) // print form information in server side
	fmt.Printf("path: %s scheme: %s\n", r.URL.Path, r.URL.Scheme)
	//fmt.Printf("URL: %v\n", r.URL.Query())

	fmt.Println("query params:")
	for k, v := range r.URL.Query() {
		fmt.Printf("\t%s -> %s\n", k, strings.Join(v, ", "))
	}
	var paramMap map[string][]string = r.Form

	for k, v := range paramMap {
		log.Printf("key: %s\tval: %s\n", k, strings.Join(v, ", "))
	}
	t := time.Now().Local()
	//fmt.Fprintf( w, "Hello astaxie %s %d%d", t.Format(time.RFC850), t.Year(), t.Month )
	data := fmt.Sprintf("Hello astaxie %s %d%d", t.Format("20060102150405"), t.Year(), t.Month)
	sendData(w, []byte(data))
}

//-----------------------------------------------------------------------------
func createSvr(addr string,
	keepalive bool,
	handlers []requestHandler) error {

	http.DefaultTransport.(*http.Transport).MaxIdleConns = 2
	http.DefaultTransport.(*http.Transport).IdleConnTimeout = 2 * time.Second
	mx := http.NewServeMux()
	// register handler for the pattern
	//
	for _, v := range handlers {
		mx.HandleFunc(v.pattern, v.method)
	}
	fmt.Printf("MUX %v\n", mx)
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
func main() {

	if len(os.Args) < 2 {
		println("path to config.ini file is missing")
		os.Exit(-1)
	}
	args := os.Args[1:]

	configMap := du.ParseIniFile(args[0])

	fmt.Printf("%v\n", configMap)

	keepAlive := true
	address := DEFAULT_ADDRESS
	sendData = createSendDataFunc(configMap["headers"])
	handlerResolver = createHandlerResolver(configMap["handlers"])

	if b, err := strconv.ParseBool(configMap["interface"]["keepalive"]); err == nil {
		keepAlive = b
	}

	if addr := configMap["interface"]["address"]; addr != "" {
		address = addr
	}
	rh := []requestHandler{{pattern: "/", method: worker}, {pattern: "/echo", method: echo}}
	err := createSvr(address, keepAlive, rh)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

package main;

import (
    "fmt"
    "net/http"
    "strings"
    "log"
    "time"
    "os"
    "strconv"
    "os/signal"
    "syscall"
    "github.com/yuin/gopher-lua"
    du "./dataserver/util"
)

const DEFAULT_ADDRESS = ":8080"

type requestHandler struct {
    method func( http.ResponseWriter, *http.Request)
    pattern string
}

var  sendData func( http.ResponseWriter, [] byte)
//-----------------------------------------------------------------------------
func createSendDataFunc( headers map[string] string ) func ( http.ResponseWriter, [] byte ) {

    hdrs := headers;
    return func ( w http.ResponseWriter, bytes [] byte ){

        if hdrs != nil {
            for k, v := range hdrs {
                w.Header().Set( k, v )
            }
        }
        w.Write( bytes)
    }
}
//-----------------------------------------------------------------------------
func createSendDataFromLuaFunc( w http.ResponseWriter, f func( http.ResponseWriter, [] byte) ) func ( L * lua.LState ) int {
    writer := w
    dataFunc := f
    return func ( L * lua.LState ) int {

        s := L.ToString( 1 )
        l := len(s)
        fmt.Printf( "STRING FROM LUA %s\n", s )
        fmt.Printf( "BYTES FROM LUA %v\n", []byte(s) )
        L.Push( lua.LNumber( l ))
        dataFunc( writer, []byte(s))
        return l;
    }
}
//-----------------------------------------------------------------------------
func sendDataFromLua( L * lua.LState ) int {

    s := L.ToString( 1 )
    l := len(s)
    fmt.Printf( "STRING FROM LUA %s\n", s )
    fmt.Printf( "BYTES FROM LUA %v\n", []byte(s) )
    L.Push( lua.LNumber( l ))
    return l;
}
//-----------------------------------------------------------------------------
func worker( w http.ResponseWriter, r *http.Request ){

    r.ParseForm()
    fmt.Printf("FORM: %v\n",r.Form) // print form information in server side
    fmt.Printf("path: %s scheme: %s\n", r.URL.Path, r.URL.Scheme)
    //fmt.Printf("URL: %v\n", r.URL.Query())

    fmt.Println( "query params:" )
    for k, v := range r.URL.Query() {
        fmt.Printf( "\t%s ==> %s\n", k, strings.Join( v, ", ") )
    }
    var paramMap map[string][]string = r.Form

    for k, v := range paramMap {
        log.Printf( "key: %s\tval: %s\n", k, strings.Join( v, ", ") )
    }
    t := time.Now().Local()
    fmt.Printf( "Handler %s %d%d", t.Format("20060102150405"), t.Year(), t.Month )
    L := lua.NewState()
    defer L.Close()
    dataFunc := createSendDataFromLuaFunc( w, sendData )
    L.SetGlobal("sendData", L.NewFunction(dataFunc)) // Register our function in Lua
    //dataFunc( L )
}
//-----------------------------------------------------------------------------
func echo( w http.ResponseWriter, r *http.Request ){

    r.ParseForm()
    fmt.Printf("FORM: %v\n",r.Form) // print form information in server side
    fmt.Printf("path: %s scheme: %s\n", r.URL.Path, r.URL.Scheme)
    //fmt.Printf("URL: %v\n", r.URL.Query())

    fmt.Println( "query params:" )
    for k, v := range r.URL.Query() {
        fmt.Printf( "\t%s -> %s\n", k, strings.Join( v, ", ") )
    }
    var paramMap map[string][]string = r.Form

    for k, v := range paramMap {
        log.Printf( "key: %s\tval: %s\n", k, strings.Join( v, ", ") )
    }
    t := time.Now().Local()
    //fmt.Fprintf( w, "Hello astaxie %s %d%d", t.Format(time.RFC850), t.Year(), t.Month )
    data := fmt.Sprintf( "Hello astaxie %s %d%d", t.Format("20060102150405"), t.Year(), t.Month )
    sendData( w, []byte(data))
/*
    y := t.Year()
    mon := t.Month()
    d := t.Day()
    h := t.Hour()
    m := t.Minute()
    s := t.Second()
    n := t.Nanosecond()

    fmt.Println("Year   :",y)
    fmt.Printf("Month   :%02d\n",int(mon))
    fmt.Println("Day   :",d)
    fmt.Println("Hour   :",h)
    fmt.Println("Minute :",m)
    fmt.Println("Second :",s)
    fmt.Println("Nanosec:",n)

    year, month, day := t.Date()
 fmt.Println("Year : ", year)
 fmt.Println("Month : ", month)
 fmt.Println("Day : ", day)
*/
}
//-----------------------------------------------------------------------------
func createSvr( addr string,
                keepalive bool,
                handlers [] requestHandler ) error {

    http.DefaultTransport.(*http.Transport).MaxIdleConns = 2
    http.DefaultTransport.(*http.Transport).IdleConnTimeout = 2 * time.Second
    mx := http.NewServeMux()
    // register handler for the pattern
    //
    for _, v := range handlers {
        mx.HandleFunc( v.pattern, v.method )
    }
    fmt.Printf( "MUX %v\n", mx )
    server := &http.Server{Addr: addr, Handler: mx}
    server.SetKeepAlivesEnabled( keepalive)
    finalCleanup( server )
    return server.ListenAndServe()
}
//-----------------------------------------------------------------------------
func finalCleanup( svr * http.Server ) {

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

    if( len( os.Args ) < 2 ) {
        println( "path to config.ini file is missing")
        os.Exit( -1 )
    }
    args := os.Args[1:]

    configMap := du.ParseIniFile( args[ 0 ] )

    fmt.Printf( "%v\n", configMap)

    keepAlive := true;
    address := DEFAULT_ADDRESS;
    sendData = createSendDataFunc( configMap["headers"] )

    if b, err := strconv.ParseBool( configMap["interface"]["keepalive"] ); err == nil {
        keepAlive = b;
    }

    if addr := configMap["interface"]["address"]; addr != "" {
        address = addr;
    }
    rh := []requestHandler{ {pattern:"/", method:echo}, {pattern:"echo", method:echo}}
    err := createSvr( address, keepAlive, rh )
    if err != nil {
        log.Fatal( "ListenAndServe: ", err )
    }
}

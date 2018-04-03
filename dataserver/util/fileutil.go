package util

import (
    "fmt"
    "strings"
    "os"
    "bufio"
    "regexp"
)

func ReadFileLines( fileName string ) []string {

    a := make( []string, 5 );
    f, err := os.Open( fileName )

    if err != nil {
        fmt.Printf("error opening file: %v\n",err)
        return nil;
    }
    defer f.Close()

    var e error = nil;
    var s string;

    r := bufio.NewReader(f)

    for e == nil {
        s,e = r.ReadString('\n')
        a = append( a, strings.Trim( s, "\r\n ") )
    }
    return a;
}

func ParseIniFile( fn string ) map[string]map[string]string {

    mp := make( map[string]map[string]string );
    var m map[string]string;
    re := regexp.MustCompile("\\[(.*)\\]");
    rb := regexp.MustCompile( "(\\[|\\])")

    for _, v := range ReadFileLines(fn ) {

        if( re.MatchString( v )){
            //fmt.Printf( "HEADER %s\n", rb.ReplaceAllString( v, "") )
            m = make( map[string]string )
            mp[ rb.ReplaceAllString( v, "") ] = m
            continue
        }
        //fmt.Println( v )

        if( strings.HasPrefix( v, ";")){ continue }

        if ss := strings.SplitN( v, "=", 2); len( ss ) > 1 {
            m[ ss[0] ] = ss[ 1 ]
        }
    }
    return mp;
}
func main(){

    if( len( os.Args ) < 2 ) {
        println( "ini file name missing")
        os.Exit( -1 )
    }
    args := os.Args[1:]
    fn   := args[ 0 ];

    m := ParseIniFile( fn )
    fmt.Printf( "%v\n", m )
}


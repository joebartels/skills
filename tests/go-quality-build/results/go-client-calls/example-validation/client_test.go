package remote

import("context";"errors";"io";"net/http";"strings";"testing")
type roundTrip func(*http.Request)(*http.Response,error)
func(f roundTrip)RoundTrip(r *http.Request)(*http.Response,error){return f(r)}
type body struct{io.Reader;closed *int}
func(b body)Close()error{*b.closed++;return nil}
func TestFetchOnce(t *testing.T){for _,value:=range []string{"ok","12345"}{closed:=0;calls:=0;c:=&http.Client{Transport:roundTrip(func(*http.Request)(*http.Response,error){calls++;return &http.Response{StatusCode:200,Body:body{strings.NewReader(value),&closed},Header:make(http.Header)},nil})};got,err:=FetchOnce(context.Background(),c,"http://dep.invalid",4);if value=="ok"&&(err!=nil||string(got)!="ok"){t.Fatalf("success %q %v",got,err)};if value!="ok"&&err==nil{t.Fatal("oversized result accepted")};if calls!=1||closed!=1{t.Fatalf("calls=%d closes=%d",calls,closed)}}}
func TestFetchOnceRetainsCause(t *testing.T){cause:=errors.New("transport cause");c:=&http.Client{Transport:roundTrip(func(*http.Request)(*http.Response,error){return nil,cause})};_,err:=FetchOnce(context.Background(),c,"http://dep.invalid",4);if !errors.Is(err,cause){t.Fatal("cause lost")}}

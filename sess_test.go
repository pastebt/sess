package sess

import (
    "testing"

    "time"
    "bufio"
    "strings"
    "net/http"
    "net/http/httptest"
)


func str2br(s string) *bufio.Reader {
    return bufio.NewReader(strings.NewReader(s))
}


func TestInit(tst *testing.T) {
    e := Init("/tmp", time.Second)
    if e != nil { tst.Error(e) }
    e = Init("/tmp", 0)
    if e != nil { tst.Error(e) }
}


func TestStart(tst *testing.T) {
    e := Init("/tmp", 2 * time.Second)
    if e != nil { tst.Error(e) }
    w := httptest.NewRecorder()
    tst.Logf("w = %v", w)
    r, e := http.ReadRequest(str2br("GET / HTTP/1.1\r\n\r\n"))
    if e != nil { tst.Error(e, r) }
    s := Start(w, r)
    s.SetCookieExpire(3 * time.Second)
    s.Set("abc", "123")
    v := s.Get("abc")
    tst.Logf("Get %v", v)
}

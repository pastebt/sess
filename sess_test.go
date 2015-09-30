package sess

import (
    "testing"

    "time"
    "bufio"
    "strings"
    "syscall"
    "net/http"
    "io/ioutil"
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
    e = Init("/tmp", 0)     // save session into file
    if e != nil { tst.Error(e) }
    // 
    w = httptest.NewRecorder()
    c := http.Cookie{Name: COOKIENAME, Value: s.si.id, Expires: time.Now().Add(10 * time.Second) }
    r.AddCookie(&c)
    s = Start(w, r)
    //
}


func TestReadOneSessFile(tst *testing.T) {
    fn := "/tmp/abcd.aaa"
    syscall.Unlink(fn)
    s, e := readOneSessFile(fn)
    if s != nil || e == nil {
        tst.Errorf("Should return error, get s=%v, e=%v", s, e)
    }

    e = ioutil.WriteFile(fn, []byte("1234"), 0666)
    if e != nil {
        tst.Errorf("Write file error %v", e)
    }
    s, e = readOneSessFile(fn)
    if s != nil || e == nil {
        tst.Errorf("Should return error, get s=%v, e=%v", s, e)
    }

    syscall.Unlink(fn)
    e = ioutil.WriteFile(fn, []byte("1234\n2015\n11"), 0666)
    if e != nil {
        tst.Errorf("Write file error %v", e)
    }
    s, e = readOneSessFile(fn)
    if e == nil {
        tst.Errorf("Should return error, get s=%v, e=%v", s, e)
    } else {
        tst.Logf("return s=%v, e=%v", s, e)
    }

    // time format ok, but expired
    syscall.Unlink(fn)
    e = ioutil.WriteFile(fn, []byte("1234\n2015-01-02 12:13:14 -0700\n{}"), 0666)
    if e != nil {
        tst.Errorf("Write file error %v", e)
    }
    s, e = readOneSessFile(fn)
    if s != nil || e != nil {
        tst.Errorf("Should return not error, get s=%v, e=%v", s, e)
    } else {
        tst.Logf("return s=%v, e=%v", s, e)
    }

    // every thing ok
    syscall.Unlink(fn)
    e = ioutil.WriteFile(fn, []byte("1234\n2030-01-02 12:13:14 -0700\n{}"), 0666)
    if e != nil {
        tst.Errorf("Write file error %v", e)
    }
    s, e = readOneSessFile(fn)
    if s == nil || e != nil {
        tst.Errorf("Should return not error, get s=%v, e=%v", s, e)
    } else {
        tst.Logf("return s=%v, e=%v", s, e)
    }

}

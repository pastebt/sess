package sess


import (
    "io"
    "os"
    "fmt"
    "sync"
    "time"
    "bytes"
    "net/http"
    "io/ioutil"
    "crypto/md5"
    "crypto/rand"
    "path/filepath"
    "encoding/json"
    "github.com/pastebt/gslog"
)


const (
    COOKIENAME string = "MYGOSESSIONID"
)


type sessInfo struct {
    id string
    m *sync.Mutex
    expire time.Time
    data map[string]string
}


type Session struct {
    si *sessInfo
    w *http.ResponseWriter
}


var sessPool = struct {
    m *sync.Mutex
    path string
    sess map[string]*sessInfo
} {&(sync.Mutex{}), "", make(map[string]*sessInfo)}


// save file name to be used save session data
// start the monitor thread to save it periodly
// Should be called before server start
func Init(sfn string) error {
    sessPool.m.Lock()
    defer sessPool.m.Unlock()

    if sessPool.path == "" { // load from disk
        logging.Debug("load session from disk")
        sessPool.path = sfn
        fs, e := filepath.Glob(filepath.Join(sfn, "*.sess"))
        if e != nil { return e }
        for _, f := range fs {
            logging.Debug("load session", f)
            dat, e := ioutil.ReadFile(f)
            if e != nil { logging.Error(e); return e }
            lines := bytes.SplitN(dat, []byte("\n"), 3)
            si := &sessInfo{id:string(lines[0]), m:&sync.Mutex{}}
            //si.expire = TODO
            e = json.Unmarshal(lines[2], &si.data)
            if e != nil { logging.Error(e); return e }
            sessPool.sess[si.id] = si
        }
    } else {                // save to disk
        logging.Debug("save session to disk")
        for n, si := range sessPool.sess {
            fout, e := os.OpenFile(filepath.Join(sfn, n) + ".sess",
                                   os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
            if e != nil { logging.Error(e); return e }
            _, e = fout.WriteString(si.id + "\n" + "1" + "\n")
            if e != nil { logging.Error(e); return e }
            j, e := json.Marshal(si.data)
            if e != nil { logging.Error(e); return e }
            fout.Write(j)
            fout.Close()
        }
    }
    _ = time.AfterFunc(time.Minute, func(){_ = Init(sfn)})
    return nil
}


var logging = gslog.GetLogger("")


func genId(addr string) (ret string) {
    h := md5.New()
    b := make([]byte, 10)
    _, _ = rand.Read(b)
    io.WriteString(h, string(b))
    io.WriteString(h, addr)
    io.WriteString(h, time.Now().String())
    for _, b := range h.Sum(nil) {
        if '0' <= b && b <= '9' ||
           'a' <= b && b <= 'z' ||
           'A' <= b && b <= 'Z' {
            ret += string(b)
        } else {
            ret += fmt.Sprintf("%x", b)
        }
    }
    return
}


func Start(w http.ResponseWriter, r *http.Request) (ses *Session) {
    c, e := r.Cookie(COOKIENAME)
    logging.Debug("e =", e)
    var si *sessInfo
    if e == nil {
        logging.Debugf("Start n = %s, v = %s", c.Name, c.Value)
        sessPool.m.Lock()
        si = sessPool.sess[c.Value]
        sessPool.m.Unlock()
    }
    logging.Debug("Start si =", si)
    if si == nil {
        logging.Debug("Start new si")
        id := genId(r.RemoteAddr)
        si = &sessInfo{id:id, m:&(sync.Mutex{}),
                       data:make(map[string]string)}
    }
    ses = &Session{si, &w}
    return
}


func (s *Session)Set(name string, value string) {
    si := s.si
    c := http.Cookie{Name:COOKIENAME, Value:si.id} //, Domain:"/"}
    http.SetCookie(*(s.w), &c)
    si.m.Lock()
    si.data[name] = value
    si.m.Unlock()

    l := gslog.GetLogger("")
    l.Debugf("n=%s, v=%s, s.id=%s", name, value, si.id)
    // TODO update expire
    sessPool.m.Lock()
    defer sessPool.m.Unlock()
    sessPool.sess[si.id] = si
    l.Debug("sessPool =", sessPool.sess)
    l.Debug("si =", si.data)
}


func (s *Session)Get(name string) (value string) {
    s.si.m.Lock()
    defer s.si.m.Unlock()
    l := gslog.GetLogger("")
    l.Debugf("Get s.id=%s", s.si.id)
    l.Debug("Get si =", s.si.data)
    value = s.si.data[name]
    return
}

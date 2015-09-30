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
    "sync/atomic"
    mrand "math/rand"
    "path/filepath"
    "encoding/json"
    "github.com/pastebt/gslog"
)


const (
    COOKIENAME string = "MYGOSESSIONID"
    TIMELAYOUT string = "2006-01-02 15:04:05 -0700"
)


type sessInfo struct {
    id string
    m *sync.Mutex
    expire time.Time
    flag int32      // == 1 means need save
    data map[string]string
}


type Session struct {
    si *sessInfo
    w *http.ResponseWriter
}


var sessPool = struct {
    m *sync.Mutex
    path string
    keep time.Duration
    sess map[string]*sessInfo
} {&(sync.Mutex{}), "", time.Hour * 7 * 24,  make(map[string]*sessInfo)}


func readOneSessFile(fn string) (si *sessInfo, err error) {
    logging.Debug("load session", fn)
    dat, err := ioutil.ReadFile(fn)
    if err != nil { return }

    lines := bytes.SplitN(dat, []byte("\n"), 3)
    if len(lines) < 3 {
        err = fmt.Errorf("Bad format")
        return
    }
    si = &sessInfo{id:string(lines[0]), m:&sync.Mutex{}}

    //si.expire
    si.expire, err = time.Parse(TIMELAYOUT, string(lines[1]))
    if err != nil { return }
    if si.expire.Before(time.Now()) {
        // this session is expired, remove file
        logging.Debugf("expire line = %s, %v", string(lines[1]), time.Now())
        err = os.Remove(fn)
        return nil, err
    }
    err = json.Unmarshal(lines[2], &si.data)
    //if err != nil { return }
    return
}


func saveOneSessFile(sfn string, si *sessInfo) (err error) {
    if (si.expire.Before(time.Now())) {
        // this session expired, skip it
        return
    }
    if atomic.LoadInt32(&si.flag) == 0 { return }
    atomic.StoreInt32(&si.flag, 0)
    logging.Debug("saveOneSessFile ", si.id)
    j, err := json.Marshal(si.data)
    if err != nil { return }

    buf := new(bytes.Buffer)
    buf.WriteString(si.id + "\n")
    buf.WriteString(si.expire.Format(TIMELAYOUT) + "\n")
    buf.Write(j)

    fout, err := os.OpenFile(filepath.Join(sfn, si.id) + ".sess",
                             os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
    if err != nil { return }
    defer fout.Close()

    _, err = buf.WriteTo(fout)
    return
}


// save file name to be used save session data
// start the monitor thread to save it periodly
// Should be called before server start
func Init(sfn string, keep time.Duration) (err error) {
    var si *sessInfo
    sessPool.m.Lock()
    defer sessPool.m.Unlock()

    if keep > 0 {
        sessPool.keep = keep
    }

    if sessPool.path == "" { // load from disk
        logging.Debug("load session from disk")
        sessPool.path = sfn
        fs, e := filepath.Glob(filepath.Join(sfn, "*.sess"))
        if e != nil { return e }
        for _, f := range fs {
            si, err = readOneSessFile(f)
            if err != nil {
                logging.Error(err)
            } else if si != nil {
                sessPool.sess[si.id] = si
            }
        }
    } else {                // save to disk
        logging.Debug("save session to disk")
        for _, si = range sessPool.sess {
            err = saveOneSessFile(sfn, si)
            if err != nil {
                logging.Error(si.id, ": ", err)
            }
        }
    }
    _ = time.AfterFunc(time.Minute, func(){ _ = Init(sfn, 0) })
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
    for ;len(ret) < 32; {
        ret += string(ret[mrand.Intn(len(ret))])
    }
    return
}


func Start(w http.ResponseWriter, r *http.Request) (ses *Session) {
    c, e := r.Cookie(COOKIENAME)
    logging.Debug("sess.Start: e =", e)
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


func (s *Session)SetCookieExpire(t time.Duration) {
    si := s.si
    var e = time.Time{}
    if t > 0 {
        e = time.Now().Add(t)
    }
    c := http.Cookie{Name:COOKIENAME, Value:si.id, Expires:e} //, Domain:"/"}
    http.SetCookie(*(s.w), &c)
    si.m.Lock()
    if e.After(si.expire) {
        si.expire = e
    }
    si.m.Unlock()

}


func (s *Session)Set(name string, value string) {
    si := s.si
    e := time.Now().Add(sessPool.keep)
    c := http.Cookie{Name:COOKIENAME, Value:si.id, Expires:e} //, Domain:"/"}
    http.SetCookie(*(s.w), &c)
    si.m.Lock()
    if e.After(si.expire) {
        si.expire = e
    }
    si.data[name] = value
    si.m.Unlock()
    atomic.StoreInt32(&si.flag, 1)

    l := gslog.GetLogger("")
    l.Debugf("n=%s, v=%s, s.id=%s", name, value, si.id)
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
    s.si.expire = time.Now().Add(sessPool.keep)
    atomic.StoreInt32(&s.si.flag, 1)
    value = s.si.data[name]
    return
}

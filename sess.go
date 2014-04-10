package sess


import (
    "sync"
    "time"
    "net/http"
    "github.com/pastebt/gslog"
)


const (
    COOKIENAME string = "MYGOSESSIONID"
)

type Session struct {
    id string
    m *sync.Mutex
    expire time.Time
    w *http.ResponseWriter
    data map[string]string
}


var sessPool = struct {
    m *sync.Mutex
    path string
    sess map[string]*Session
} {&(sync.Mutex{}), "", make(map[string]*Session)}


// save file name to be used save session data
// start the monitor thread to save it periodly
// Should be called before server start
func Init(sfn string) (err error) {
    return
}


var logging = gslog.GetLogger("")


func Start(w http.ResponseWriter, r *http.Request) (ses *Session) {
    c, e := r.Cookie(COOKIENAME)
    logging.Debug("e =", e)
    if e == nil {
        logging.Debugf("n = %s, v = %s", c.Name, c.Value)
        sessPool.m.Lock()
        defer sessPool.m.Unlock()
        ses = sessPool.sess[c.Value]
    }
    logging.Debug("ses =", ses)
    if ses == nil {
        id := "1234"    // TODO, generate id
        ses = &Session{id:id, m:&(sync.Mutex{}),
                       data:make(map[string]string)}
    }
    ses.w = &w
    return
}


func (s *Session)Set(name string, value string) {
    c := http.Cookie{Name:COOKIENAME, Value:s.id, Domain:"/"}
    http.SetCookie(*(s.w), &c)
    s.m.Lock()
    defer s.m.Unlock()
    s.data[name] = value
    l := gslog.GetLogger("")
    l.Debugf("n=%s, v=%s, s.id=%s", name, value, s.id)
    // TODO update expire
    sessPool.m.Lock()
    defer sessPool.m.Unlock()
    sessPool.sess[s.id] = s
    l.Debug("sessPool =", sessPool.sess)
    l.Debug("s =", s.data)
}


func (s *Session)Get(name string) (value string) {
    s.m.Lock()
    defer s.m.Unlock()
    value = s.data[name]
    return
}

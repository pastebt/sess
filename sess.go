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
func Init(sfn string) (err error) {
    return
}


var logging = gslog.GetLogger("")


func Start(w http.ResponseWriter, r *http.Request) (ses *Session) {
    c, e := r.Cookie(COOKIENAME)
    logging.Debug("e =", e)
    var si *sessInfo
    if e == nil {
        logging.Debugf("n = %s, v = %s", c.Name, c.Value)
        sessPool.m.Lock()
        defer sessPool.m.Unlock()
        si = sessPool.sess[c.Value]
    }
    logging.Debug("si =", si)
    if ses == nil {
        id := "1234"    // TODO, generate id
        si = &sessInfo{id:id, m:&(sync.Mutex{}),
                       data:make(map[string]string)}
    }
    ses = &Session{si, &w}
    return
}


func (s *Session)Set(name string, value string) {
    si := s.si
    c := http.Cookie{Name:COOKIENAME, Value:si.id, Domain:"/"}
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
    value = s.si.data[name]
    return
}

# Sess
Simple easy Session System

## Install

```bash
go get github.com/pastebt/sess
```

## Usage
### init
You have to init the session system when your server start, give the path 
where to keep persistance data for your session. 

```go
import "github.com/pastebt/sess"

sess.Init("")
```
If leave it as default "", session will keep in memory.


### Get / Set session
In a net/http handler function, you can use it

```go
func handler(w ResponseWriter, r *Request) {
  s : = sess.Start(w, r)
  name := s.Get("username")
  s.Set("password", "whateveryourpassword")
}
```

# Sess
Simple easy Session System

## Install

```bash
go get github.com/pastebt/sess
```

## Usage
### Init
You have to init the session system when your server start, give the path 
where to save persistent data for your session. 

```go
import "github.com/pastebt/sess"

sess.Init("/path/to/save/data/")
```
If leave it as default "", session will keep in memory and will be lost when server stop.


### Get / Set session
In a net/http handler function, you can use it like this:

```go
func handler(w ResponseWriter, r *Request) {
  s := sess.Start(w, r)
  name := s.Get("username")
  s.Set("password", "whateveryourpassword")
  
  ...
  
}
```

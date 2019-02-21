package common

import (
    "encoding/json"
    "fmt"
    "github.com/satori/go.uuid"
    "net"
    "net/http"
    "time"
)

const (
    TimeFmt = "2006-02-01 15:04:05.000"

)
type Request struct {
    Id int
    Time string
    Uid  string
}
func ObtainClientInfo(r *http.Request) string {
    ip,port,err := net.SplitHostPort(r.RemoteAddr)
    if err !=nil {
        return err.Error()
    }
    return fmt.Sprintf("client info:%s:%d",ip,port)
}
func RequestToString(id int) ([]byte, error) {
    uid, err := uuid.NewV4()
    if err != nil {
        return nil, err
    }

    req := Request{
        Id:id,
        Time: time.Now().Format(TimeFmt),
        Uid:  uid.String(),
    }
    b, err := json.Marshal(&req)
    if err != nil {
        return nil, err
    }
    return b, nil

}
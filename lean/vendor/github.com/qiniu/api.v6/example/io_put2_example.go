package main

import (
    "bufio"
    "fmt"
    . "github.com/qiniu/api.v6/conf"
    qnio "github.com/qiniu/api.v6/io"
    "github.com/qiniu/api.v6/rs"
    "os"
)

func main() {
    ACCESS_KEY = "xxxxxxxx"
    SECRET_KEY = "xxxxxxxx"

    var ret qnio.PutRet

    var extra = &qnio.PutExtra{
        MimeType: "image/jepg",
        CheckCrc: 0,
    }
    key := "1024x1024.jpg"

    scope := fmt.Sprintf("skypixeltest:%s", key)

    putPolicy := rs.PutPolicy{
        Scope: scope,
        // Expires:      expires,
    }
    uptoken := putPolicy.Token(nil)

    fi, err := os.Open("/Users/qpzhang/Downloads/1024x1024.jpg")
    st, _ := fi.Stat()
    if err != nil {
        panic(err)
    }
    defer fi.Close()
    data := bufio.NewReader(fi)

    fmt.Println("size ", st.Size())
    err = qnio.Put2(nil, &ret, uptoken, key, data, st.Size(), extra)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println("put sucess......", ret)
    }
}

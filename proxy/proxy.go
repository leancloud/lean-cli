package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/aisk/logp"
	"github.com/leancloud/lean-cli/api"
	"nhooyr.io/websocket"
)

type ProxyInfo struct {
	AppID        string
	ClusterId    int
	Name         string
	Runtime      string
	AuthUser     string
	AuthPassword string
	LocalPort    string
	Connected    chan bool

	baseURL   string
	headers   map[string]string
	cookieJar http.CookieJar
}

func Run(proxyInfo *ProxyInfo) error {
	client := api.NewClientByApp(proxyInfo.AppID)
	proxyInfo.baseURL = client.GetBaseURL()
	proxyInfo.headers = client.GetAuthHeaders()
	proxyInfo.cookieJar = client.CookieJar

	path := fmt.Sprintf("/1.1/leandb/proxy/ws?clusterid=%d", proxyInfo.ClusterId)
	remoteURL := strings.Replace(proxyInfo.baseURL, "https", "wss", 1) + path

	l, err := net.Listen("tcp", ":"+proxyInfo.LocalPort)
	if err != nil {
		return err
	}

	logp.Infof("Now, you can connect instance via [127.0.0.1:%s] with username [%s] password [%s]\r\n", proxyInfo.LocalPort, proxyInfo.AuthUser, proxyInfo.AuthPassword)

	// notify shell proxy action
	proxyInfo.Connected <- true

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go proxy(conn, remoteURL, proxyInfo)
	}
}

func proxy(conn net.Conn, remoteURL string, proxyInfo *ProxyInfo) {
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	opts := &websocket.DialOptions{
		HTTPHeader: http.Header{},
		HTTPClient: &http.Client{},
	}
	for k, v := range proxyInfo.headers {
		opts.HTTPHeader.Add(k, v)
	}
	if proxyInfo.cookieJar != nil {
		// TODO
		// opts.HTTPClient.Jar = proxyInfo.cookieJar
	}

	c, _, err := websocket.Dial(ctx, remoteURL, opts)
	if err != nil {
		log.Println(err)
		return
	}

	remote := websocket.NetConn(ctx, c, websocket.MessageBinary)
	defer remote.Close()

	go io.Copy(remote, conn)
	io.Copy(conn, remote)
}

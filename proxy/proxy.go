package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/aisk/logp"
	"github.com/leancloud/lean-cli/api"
	"github.com/urfave/cli"
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

	switch proxyInfo.Runtime {
	case "redis":
	case "mongo":
	case "udb":
		return proxyTcp(proxyInfo)
	case "es":
		return proxyHttp(proxyInfo)
	default:
		return cli.NewExitError(fmt.Sprintf("proxy not support runtime %s", proxyInfo.Runtime), 1)
	}
	return nil
}

//
// http
//
func proxyHttp(proxyInfo *ProxyInfo) error {
	path := fmt.Sprintf("/1.1/leandb/proxy/http?clusterid=%d", proxyInfo.ClusterId)
	remoteURL := proxyInfo.baseURL + path

	logp.Infof("Now, you can connect instance via [127.0.0.1:%s]", proxyInfo.LocalPort)

	http.HandleFunc("/", newHttpProxyHandler(remoteURL, proxyInfo))
	log.Fatal(http.ListenAndServe(":"+proxyInfo.LocalPort, nil))

	return nil
}

func newHttpProxyHandler(remoteURL string, proxyInfo *ProxyInfo) func(http.ResponseWriter, *http.Request) {
	url, err := url.Parse(remoteURL)
	if err != nil {
		panic(err)
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(url)
	originDirector := reverseProxy.Director
	reverseProxy.Director = func(r *http.Request) {
		originDirector(r)
		for k, v := range proxyInfo.headers {
			r.Header.Add(k, v)
		}
		if proxyInfo.cookieJar != nil {
			for _, c := range proxyInfo.cookieJar.Cookies(url) {
				r.AddCookie(c)
			}
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		reverseProxy.ServeHTTP(w, r)
	}
}

//
// tcp
//
func proxyTcp(proxyInfo *ProxyInfo) error {
	path := fmt.Sprintf("/1.1/leandb/proxy/ws?clusterid=%d", proxyInfo.ClusterId)
	remoteURL := strings.Replace(proxyInfo.baseURL, "https", "wss", 1) + path

	l, err := net.Listen("tcp", ":"+proxyInfo.LocalPort)
	if err != nil {
		return err
	}

	logp.Infof("Now, you can connect instance via [127.0.0.1:%s] with username [%s] password [%s]", proxyInfo.LocalPort, proxyInfo.AuthUser, proxyInfo.AuthPassword)

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleTcpProxy(conn, remoteURL, proxyInfo)
	}
}

func handleTcpProxy(conn net.Conn, remoteURL string, proxyInfo *ProxyInfo) {
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
		opts.HTTPClient.Jar = proxyInfo.cookieJar
	}

	c, _, err := websocket.Dial(ctx, remoteURL, opts)
	if err != nil {
		log.Println(err)
		return
	}

	// notify shell proxy action
	if proxyInfo.Connected != nil {
		proxyInfo.Connected <- true
	}

	remote := websocket.NetConn(ctx, c, websocket.MessageBinary)
	defer remote.Close()

	go io.Copy(remote, conn)
	io.Copy(conn, remote)
}

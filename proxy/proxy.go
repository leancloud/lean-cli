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

	logp.Infof("Now, you can connect instance via %s\r\n", getCliArgs(proxyInfo))

	// notify shell proxy action
	proxyInfo.Connected <- true

	// TODO shell proxy need two Ctrl-C
	// sigs := make(chan os.Signal, 1)
	// done := make(chan bool, 1)
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	//
	// go func() {
	// <-sigs
	// done <- true
	// }()

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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
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

	pingWithTicker(ctx, c)

	remote := websocket.NetConn(ctx, c, websocket.MessageBinary)
	defer remote.Close()

	go io.Copy(remote, conn)
	io.Copy(conn, remote)
}

func pingWithTicker(ctx context.Context, c *websocket.Conn) {
	ticker := time.NewTicker(4 * time.Minute)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				c.Ping(ctx)
			}
		}
	}()
}

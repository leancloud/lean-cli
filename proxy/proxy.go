package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
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
}

var connTimeout = 2 * time.Hour
var pingInterval = 4 * time.Minute

func RunProxy(p *ProxyInfo) error {
	l, err := net.Listen("tcp", ":"+p.LocalPort)
	if err != nil {
		return err
	}

	cli, err := GetCli(p, false)
	if err != nil {
		return err
	}
	logp.Infof("Now, you can connect [%s] via %s\r\n", p.Name, cli)

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go proxy(conn, p)
	}
}

func RunShellProxy(p *ProxyInfo, started, term chan bool) error {
	l, err := net.Listen("tcp", ":"+p.LocalPort)
	if err != nil {
		return err
	}

	started <- true
	go func() {
		select {
		case <-term:
			os.Exit(0)
		}
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go proxy(conn, p)
	}
}

func proxy(conn net.Conn, p *ProxyInfo) {
	client := api.NewClientByApp(p.AppID)
	path := fmt.Sprintf("/1.1/leandb/proxy/ws?clusterid=%d", p.ClusterId)
	remoteURL := strings.Replace(client.GetBaseURL(), "https", "wss", 1) + path

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()

	c, _, err := websocket.Dial(ctx, remoteURL, buildOpts(p, client))
	if err != nil {
		logp.Warnf("Dial remote websocket endpoint get error: %s", err)
		return
	}

	pingWithTicker(ctx, c)

	remote := websocket.NetConn(ctx, c, websocket.MessageBinary)
	defer remote.Close()

	go io.Copy(remote, conn)
	io.Copy(conn, remote)
}

func buildOpts(p *ProxyInfo, client *api.Client) *websocket.DialOptions {
	opts := &websocket.DialOptions{
		HTTPHeader: http.Header{},
		HTTPClient: &http.Client{},
	}
	for k, v := range client.GetAuthHeaders() {
		opts.HTTPHeader.Add(k, v)
	}
	if client.AccessToken == "" && client.CookieJar != nil {
		opts.HTTPClient.Jar = client.CookieJar
	}

	return opts
}

func pingWithTicker(ctx context.Context, c *websocket.Conn) {
	ticker := time.NewTicker(pingInterval)

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

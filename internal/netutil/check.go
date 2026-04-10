package netutil

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

type CheckResult struct {
	Host      string `json:"host"`
	PingOK    bool   `json:"ping_ok"`
	TCPPort   int    `json:"tcp_port,omitempty"`
	TCPOK     bool   `json:"tcp_ok,omitempty"`
	HTTPURL   string `json:"http_url,omitempty"`
	HTTPCode  int    `json:"http_code,omitempty"`
	HTTPOK    bool   `json:"http_ok,omitempty"`
	Advice    string `json:"advice,omitempty"`
	Timestamp string `json:"timestamp"`
}

func Ping(host string, timeout time.Duration) bool {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd = exec.Command("ping", "-c", "1", "-W", fmt.Sprintf("%d", int(timeout.Seconds())), host)
	default:
		return false
	}
	return cmd.Run() == nil
}

func TCPCheck(host string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func HTTPCheck(url string, timeout time.Duration) (int, bool) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return 0, false
	}
	defer resp.Body.Close()
	return resp.StatusCode, resp.StatusCode < 400
}

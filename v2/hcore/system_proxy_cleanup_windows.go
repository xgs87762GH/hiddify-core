//go:build windows

package hcore

import (
	"strings"

	"github.com/sagernet/sing/common/wininet"
	"golang.org/x/sys/windows/registry"
)

const internetSettingsKey = `Software\Microsoft\Windows\CurrentVersion\Internet Settings`

// clearStaleSystemProxy clears a system proxy left behind by an abnormal exit
// (crash, force-kill, shutdown) when it still points at a local listener this
// app owns. A normal stop already disables it; this only covers the paths that
// never ran, so a stale 127.0.0.1 proxy cannot outlive the app.
func clearStaleSystemProxy() {
	key, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsKey, registry.QUERY_VALUE)
	if err != nil {
		return
	}
	enable, _, err := key.GetIntegerValue("ProxyEnable")
	if err != nil || enable == 0 {
		key.Close()
		return
	}
	server, _, err := key.GetStringValue("ProxyServer")
	key.Close()
	if err != nil {
		return
	}
	// Only touch a proxy that points at this machine: an enterprise or PAC proxy
	// must survive untouched.
	if !strings.Contains(server, "127.0.0.1") && !strings.Contains(server, "localhost") {
		return
	}
	_ = wininet.ClearSystemProxy()
}

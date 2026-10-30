//go:build windows

package hcore

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sagernet/sing/common/wininet"
	"golang.org/x/sys/windows/registry"
)

const internetSettingsKey = `Software\Microsoft\Windows\CurrentVersion\Internet Settings`

// systemProxyRecordPath is where the ProxyServer value this app wrote is kept.
// It lives next to the other runtime data so it is removed with the app data.
func systemProxyRecordPath() string {
	return filepath.Join(sWorkingPath, "data", "system_proxy.last")
}

// readSystemProxy returns the current ProxyEnable/ProxyServer from the registry.
func readSystemProxy() (enabled bool, server string) {
	key, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsKey, registry.QUERY_VALUE)
	if err != nil {
		return false, ""
	}
	defer key.Close()
	enable, _, err := key.GetIntegerValue("ProxyEnable")
	if err != nil || enable == 0 {
		return false, ""
	}
	value, _, err := key.GetStringValue("ProxyServer")
	if err != nil {
		return true, ""
	}
	return true, value
}

// clearStaleSystemProxy clears a system proxy left behind by an abnormal exit
// (crash, force-kill, shutdown), but only when the registry still holds exactly
// the value this app wrote last time. A proxy set by another program never
// matches, so it is left untouched.
func clearStaleSystemProxy() {
	recorded, err := os.ReadFile(systemProxyRecordPath())
	if err != nil {
		return
	}
	want := strings.TrimSpace(string(recorded))
	if want == "" {
		return
	}
	enabled, server := readSystemProxy()
	if !enabled || server != want {
		return
	}
	_ = wininet.ClearSystemProxy()
	_ = os.Remove(systemProxyRecordPath())
}

// rememberSystemProxy records the proxy this app just enabled so a later start
// can recognise - and only then clear - its own leftover. It records nothing
// unless this app actually asked for the system proxy, so a proxy another
// program set while the app runs in TUN/proxy mode is never mistaken for ours.
func rememberSystemProxy() {
	opts := static.HiddifyOptions
	if opts == nil || !opts.InboundOptions.SetSystemProxy {
		return
	}
	enabled, server := readSystemProxy()
	if !enabled || server == "" {
		return
	}
	if !strings.Contains(server, "127.0.0.1") && !strings.Contains(server, "localhost") {
		return
	}
	_ = os.MkdirAll(filepath.Dir(systemProxyRecordPath()), 0o755)
	_ = os.WriteFile(systemProxyRecordPath(), []byte(server), 0o644)
}

package web

import (
	"fmt"
	"slices"
)

// App is a RemoteApp published through the gateway: instead of a full desktop
// the client gets only the program's windows. Program is the value of
// remoteapplicationprogram, usually "||<alias>" for an alias registered in the
// host's TSAppAllowList.
type App struct {
	Id      string
	Name    string
	Host    string
	Program string
	CmdLine string
}

// validateApps checks that every app has the fields needed to build an RDP
// file, that ids are unique, and that the app's host is one the gateway is
// allowed to connect to.
func validateApps(apps []App, hosts []string) error {
	seen := make(map[string]bool, len(apps))
	for _, a := range apps {
		if a.Id == "" || a.Program == "" || a.Host == "" {
			return fmt.Errorf("app %q: id, host and program are required", a.Id)
		}
		if seen[a.Id] {
			return fmt.Errorf("app %q: duplicate id", a.Id)
		}
		seen[a.Id] = true
		if !slices.Contains(hosts, a.Host) {
			return fmt.Errorf("app %q: host %s is not in the gateway hosts list", a.Id, a.Host)
		}
	}
	return nil
}

func (h *Handler) findApp(id string) *App {
	for i := range h.apps {
		if h.apps[i].Id == id {
			return &h.apps[i]
		}
	}
	return nil
}

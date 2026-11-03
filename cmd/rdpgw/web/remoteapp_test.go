package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/bolkedebruin/rdpgw/cmd/rdpgw/identity"
	"github.com/bolkedebruin/rdpgw/cmd/rdpgw/rdp"
)

var testApps = []App{
	{Id: "edge", Name: "Microsoft Edge", Host: "10.0.0.1:3389", Program: "||edge"},
	{Id: "notes", Name: "Notes", Host: "remote.host.com", Program: "||notepad", CmdLine: "C:\\notes.txt"},
}

func remoteAppHandler() *Handler {
	u, _ := url.Parse(gateway)
	c := Config{
		HostSelection:     "unsigned",
		Hosts:             hosts,
		Apps:              testApps,
		PAATokenGenerator: paaTokenMock,
		GatewayAddress:    u,
	}
	return c.NewHandler()
}

func authenticatedRequest(t *testing.T, target string) *http.Request {
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		t.Fatal(err)
	}
	id := identity.NewUser()
	id.SetUserName(testuser)
	id.SetAuthenticated(true)
	return identity.AddToRequestCtx(id, req)
}

func TestHandleDownloadRemoteApp(t *testing.T) {
	h := remoteAppHandler()
	rr := httptest.NewRecorder()
	h.HandleDownload(rr, authenticatedRequest(t, "/connect?app=notes"))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %v want %v (%s)", rr.Code, http.StatusOK, rr.Body.String())
	}
	data := rdpToMap(strings.Split(rr.Body.String(), rdp.CRLF))
	want := map[string]string{
		"full address":             "remote.host.com",
		"remoteapplicationmode":    "1",
		"remoteapplicationprogram": "||notepad",
		"remoteapplicationname":    "Notes",
		"remoteapplicationcmdline": "C:\\notes.txt",
	}
	for k, v := range want {
		if data[k] != v {
			t.Errorf("%s: got %q want %q", k, data[k], v)
		}
	}
	token, _ := paaTokenMock(context.Background(), testuser, "remote.host.com")
	if data["gatewayaccesstoken"] != token {
		t.Errorf("gatewayaccesstoken not issued for the app host: got %q want %q", data["gatewayaccesstoken"], token)
	}
}

func TestHandleDownloadUnknownApp(t *testing.T) {
	h := remoteAppHandler()
	rr := httptest.NewRecorder()
	h.HandleDownload(rr, authenticatedRequest(t, "/connect?app=nope"))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %v want %v", rr.Code, http.StatusBadRequest)
	}
}

func TestHandleDownloadDesktopHasNoRemoteApp(t *testing.T) {
	h := remoteAppHandler()
	rr := httptest.NewRecorder()
	h.HandleDownload(rr, authenticatedRequest(t, "/connect?host=10.0.0.1:3389"))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %v want %v", rr.Code, http.StatusOK)
	}
	data := rdpToMap(strings.Split(rr.Body.String(), rdp.CRLF))
	if _, ok := data["remoteapplicationprogram"]; ok {
		t.Errorf("desktop connection must not carry remoteapplicationprogram")
	}
}

func TestHandleHostListIncludesApps(t *testing.T) {
	h := remoteAppHandler()
	rr := httptest.NewRecorder()
	h.HandleHostList(rr, authenticatedRequest(t, "/api/v1/hosts"))

	var got []Host
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(hosts)+len(testApps) {
		t.Fatalf("entries: got %d want %d", len(got), len(hosts)+len(testApps))
	}
	apps := got[len(hosts):]
	if apps[0].App != "edge" || apps[0].Name != "Microsoft Edge" || apps[0].IsDefault {
		t.Errorf("unexpected app entry: %+v", apps[0])
	}
	for _, e := range got[:len(hosts)] {
		if e.App != "" {
			t.Errorf("host entry carries app id: %+v", e)
		}
	}
}

func TestValidateAppsRejectsUnknownHost(t *testing.T) {
	err := validateApps([]App{{Id: "x", Name: "X", Host: "elsewhere:3389", Program: "||x"}}, hosts)
	if err == nil {
		t.Error("app on a host outside Server.Hosts must be rejected")
	}
	if err := validateApps(testApps, hosts); err != nil {
		t.Errorf("valid apps rejected: %v", err)
	}
	if err := validateApps([]App{testApps[0], testApps[0]}, hosts); err == nil {
		t.Error("duplicate app ids must be rejected")
	}
}

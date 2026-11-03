package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadApps(t *testing.T) {
	f := filepath.Join(t.TempDir(), "rdpgw.yaml")
	yaml := `Server:
  Hosts:
    - win.example:3389
  HostSelection: unsigned
  Tls: disable
  SessionKey: thisisasessionkeyreplacethisjetzt
  SessionEncryptionKey: thisisasessionkeyreplacethisnunu
  Apps:
    - Id: edge
      Name: Microsoft Edge
      Host: win.example:3389
      Program: "||edge"
      CmdLine: "--profile-directory=Default"
Authentication:
  - local
`
	if err := os.WriteFile(f, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	c := Load(f)
	if len(c.Server.Apps) != 1 {
		t.Fatalf("apps: got %d want 1", len(c.Server.Apps))
	}
	want := AppConfig{Id: "edge", Name: "Microsoft Edge", Host: "win.example:3389", Program: "||edge", CmdLine: "--profile-directory=Default"}
	if c.Server.Apps[0] != want {
		t.Errorf("app: got %+v want %+v", c.Server.Apps[0], want)
	}
}

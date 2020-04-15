package command

import (
	"testing"
)

func TestConfigGet(t *testing.T) {
	cfg := `---
hosts:
  github.com:
    user: OWNER
    oauth_token: MUSTBEHIGHCUZIMATOKEN
editor: ed
`
	initBlankContext(cfg, "OWNER/REPO", "master")

	output, err := RunCommand(configGetCmd, "config get editor")
	if err != nil {
		t.Fatalf("error running command `config get editor`: %v", err)
	}

	eq(t, output.String(), "ed\n")
}

func TestConfigGet_not_found(t *testing.T) {
	initBlankContext("", "OWNER/REPO", "master")

	output, err := RunCommand(configGetCmd, "config get editor")
	if err != nil {
		t.Fatalf("error running command `config get editor`: %v", err)
	}

	eq(t, output.String(), "")
}

func TestConfigSet(t *testing.T) {
	initBlankContext("", "OWNER/REPO", "master")
}

func TestConfigSet_update(t *testing.T) {
}

func TestConfigGetHost(t *testing.T) {
}

func TestConfigSetHost(t *testing.T) {
}

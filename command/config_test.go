package command

import (
	"testing"

	"github.com/cli/cli/context"
)

func TestConfigGet(t *testing.T) {
	initBlankContext("OWNER/REPO", "master")
	defer context.StubConfig(`---
hosts:
  github.com:
    user: OWNER
    oauth_token: MUSTBEHIGHCUZIMATOKEN
editor: ed
`)()

	output, err := RunCommand(configGetCmd, "config get editor")
	if err != nil {
		t.Fatalf("error running command `config get editor`: %v", err)
	}

	eq(t, output.String(), "ed\n")
}

func TestConfigGet_not_found(t *testing.T) {
}

func TestConfigSet(t *testing.T) {
}

func TestConfigSet_update(t *testing.T) {
}

func TestConfigGetHost(t *testing.T) {
}

func TestConfigSetHost(t *testing.T) {
}

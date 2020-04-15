package context

import (
	"errors"
	"reflect"
	"testing"
)

func stubbedReadConfig(content string) func(fn string) ([]byte, error) {
	return func(fn string) ([]byte, error) {
		return []byte(content), nil
	}
}

func stubConfig(content string) func() {
	orig := readConfig
	readConfig = stubbedReadConfig(content)
	return func() {
		readConfig = orig
	}
}

func eq(t *testing.T, got interface{}, expected interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func Test_parseConfig(t *testing.T) {
	defer stubConfig(`---
hosts:
  github.com:
    user: monalisa
    oauth_token: OTOKEN
`)()
	config, err := parseConfig("filename")
	eq(t, err, nil)
	user, err := config.Get("github.com", "user")
	eq(t, err, nil)
	eq(t, user, "monalisa")
	token, err := config.Get("github.com", "oauth_token")
	eq(t, err, nil)
	eq(t, token, "OTOKEN")
}

func Test_parseConfig_multipleHosts(t *testing.T) {
	defer stubConfig(`---
hosts:
  example.com:
    user: wronguser
    oauth_token: NOTTHIS
  github.com:
    user: monalisa
    oauth_token: OTOKEN
`)()
	config, err := parseConfig("filename")
	eq(t, err, nil)
	user, err := config.Get("github.com", "user")
	eq(t, err, nil)
	eq(t, user, "monalisa")
	token, err := config.Get("github.com", "oauth_token")
	eq(t, err, nil)
	eq(t, token, "OTOKEN")
}

func Test_parseConfig_notFound(t *testing.T) {
	defer stubConfig(`---
hosts:
  example.com:
    user: wronguser
    oauth_token: NOTTHIS
`)()
	config, err := parseConfig("filename")
	eq(t, err, nil)
	_, err = config.configForHost("github.com")
	eq(t, err, errors.New(`could not find config entry for "github.com"`))
}

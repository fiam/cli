package context

import (
	"reflect"
	"testing"
)

func stubbedReadConfig(content string) func(fn string) ([]byte, error) {
	return func(fn string) ([]byte, error) {
		return []byte(content), nil
	}
}

func StubConfig(content string) func() {
	orig := ReadConfigFile
	ReadConfigFile = stubbedReadConfig(content)
	return func() {
		ReadConfigFile = orig
	}
}

func eq(t *testing.T, got interface{}, expected interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

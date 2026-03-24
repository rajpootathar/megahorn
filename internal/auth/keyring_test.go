package auth

import "testing"

func TestKeyringServiceName(t *testing.T) {
	k := NewKeyring()
	if k.ServiceName() != "megahorn" {
		t.Errorf("expected megahorn, got %s", k.ServiceName())
	}
}

func TestKeyringKey(t *testing.T) {
	k := NewKeyring()
	key := k.Key("twitter", "access_token")
	if key != "megahorn:twitter:access_token" {
		t.Errorf("expected megahorn:twitter:access_token, got %s", key)
	}
}

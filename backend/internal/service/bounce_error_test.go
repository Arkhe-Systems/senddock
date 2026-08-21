package service

import (
	"errors"
	"fmt"
	"net/textproto"
	"testing"
)

func TestClassifyBounceKeepsErrorChain(t *testing.T) {
	orig := &textproto.Error{Code: 550, Msg: "user unknown"}
	be := classifyBounce(orig)
	if be == nil {
		t.Fatal("expected a BounceError for a 550")
	}
	if be.Code != 550 || be.Message != "user unknown" {
		t.Fatalf("unexpected fields: %+v", be)
	}

	var protoErr *textproto.Error
	if !errors.As(be, &protoErr) {
		t.Fatal("BounceError must unwrap to the original *textproto.Error")
	}
	if protoErr.Code != 550 {
		t.Fatalf("unwrapped code = %d, want 550", protoErr.Code)
	}
}

func TestClassifyBounceIgnoresNonBounce(t *testing.T) {
	cases := map[string]error{
		"nil":     nil,
		"4xx":     &textproto.Error{Code: 421, Msg: "try later"},
		"wrapped": fmt.Errorf("smtp auth failed: %w", errors.New("bad credentials")),
	}
	for name, err := range cases {
		if be := classifyBounce(err); be != nil {
			t.Errorf("%s: expected nil, got %+v", name, be)
		}
	}
}

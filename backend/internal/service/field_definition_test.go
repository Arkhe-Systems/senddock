package service

import (
	"encoding/json"
	"testing"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/sqlc-dev/pqtype"
)

func enumDef(key string, options []string) db.SubscriberFieldDefinition {
	raw, _ := json.Marshal(options)
	return db.SubscriberFieldDefinition{
		Key:       key,
		FieldType: string(FieldTypeEnum),
		Options:   pqtype.NullRawMessage{RawMessage: raw, Valid: true},
	}
}

func TestValidateFieldValue(t *testing.T) {
	cases := []struct {
		name    string
		def     db.SubscriberFieldDefinition
		value   any
		wantErr bool
	}{
		{"string ok", db.SubscriberFieldDefinition{Key: "n", FieldType: string(FieldTypeString)}, "hi", false},
		{"string wrong", db.SubscriberFieldDefinition{Key: "n", FieldType: string(FieldTypeString)}, 12.0, true},
		{"number ok", db.SubscriberFieldDefinition{Key: "n", FieldType: string(FieldTypeNumber)}, 12.0, false},
		{"number wrong", db.SubscriberFieldDefinition{Key: "n", FieldType: string(FieldTypeNumber)}, "12", true},
		{"boolean ok", db.SubscriberFieldDefinition{Key: "n", FieldType: string(FieldTypeBoolean)}, true, false},
		{"date ok", db.SubscriberFieldDefinition{Key: "n", FieldType: string(FieldTypeDate)}, "1990-05-12", false},
		{"date wrong", db.SubscriberFieldDefinition{Key: "n", FieldType: string(FieldTypeDate)}, "12/05/1990", true},
		{"enum ok", enumDef("plan", []string{"free", "pro"}), "pro", false},
		{"enum wrong", enumDef("plan", []string{"free", "pro"}), "team", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateFieldValue(tc.def, tc.value)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateFieldValue(%v) error = %v, wantErr = %v", tc.value, err, tc.wantErr)
			}
		})
	}
}

func TestOptionsRoundTrip(t *testing.T) {
	options := []string{"a", "b", "c"}
	raw, err := optionsToRaw(options)
	if err != nil {
		t.Fatalf("optionsToRaw: %v", err)
	}
	got := optionsFromRaw(raw)
	if len(got) != len(options) {
		t.Fatalf("round trip length mismatch: got %v", got)
	}
	for i := range options {
		if got[i] != options[i] {
			t.Fatalf("round trip mismatch at %d: got %s want %s", i, got[i], options[i])
		}
	}
	if optionsFromRaw(pqtype.NullRawMessage{}) != nil {
		t.Fatalf("expected nil options for null raw message")
	}
}

func TestValidateTypeRejectsUnknown(t *testing.T) {
	if err := validateType("uuid"); err == nil {
		t.Fatal("expected error for unknown field type")
	}
	for _, ok := range []FieldType{FieldTypeString, FieldTypeNumber, FieldTypeDate, FieldTypeBoolean, FieldTypeEnum} {
		if err := validateType(string(ok)); err != nil {
			t.Fatalf("expected %s to be valid, got %v", ok, err)
		}
	}
}

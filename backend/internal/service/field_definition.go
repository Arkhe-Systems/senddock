package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeNumber  FieldType = "number"
	FieldTypeDate    FieldType = "date"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeEnum    FieldType = "enum"
)

var (
	ErrFieldValidation  = errors.New("field validation failed")
	ErrUnknownField     = fmt.Errorf("%w: unknown field", ErrFieldValidation)
	ErrInvalidFieldKey  = errors.New("field key must start with a letter and contain only letters, numbers, or underscores")
	ErrInvalidFieldType = errors.New("field type must be one of string, number, date, boolean, enum")
	ErrEnumNeedsOptions = errors.New("enum fields require at least one option")
)

var fieldKeyPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

type FieldDefinitionService struct {
	queries *db.Queries
}

func NewFieldDefinitionService(queries *db.Queries) *FieldDefinitionService {
	return &FieldDefinitionService{queries: queries}
}

type FieldDefinitionInput struct {
	Key      string
	Label    string
	Type     string
	Options  []string
	Required bool
}

func validateType(t string) error {
	switch FieldType(t) {
	case FieldTypeString, FieldTypeNumber, FieldTypeDate, FieldTypeBoolean, FieldTypeEnum:
		return nil
	default:
		return ErrInvalidFieldType
	}
}

func optionsToRaw(options []string) (pqtype.NullRawMessage, error) {
	if len(options) == 0 {
		return pqtype.NullRawMessage{}, nil
	}
	raw, err := json.Marshal(options)
	if err != nil {
		return pqtype.NullRawMessage{}, err
	}
	return pqtype.NullRawMessage{RawMessage: raw, Valid: true}, nil
}

func optionsFromRaw(raw pqtype.NullRawMessage) []string {
	if !raw.Valid || len(raw.RawMessage) == 0 {
		return nil
	}
	var options []string
	if err := json.Unmarshal(raw.RawMessage, &options); err != nil {
		return nil
	}
	return options
}

func (s *FieldDefinitionService) Create(ctx context.Context, projectID string, in FieldDefinitionInput) (db.SubscriberFieldDefinition, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.SubscriberFieldDefinition{}, errors.New("invalid project id")
	}
	if !fieldKeyPattern.MatchString(in.Key) {
		return db.SubscriberFieldDefinition{}, ErrInvalidFieldKey
	}
	if err := validateType(in.Type); err != nil {
		return db.SubscriberFieldDefinition{}, err
	}
	if FieldType(in.Type) == FieldTypeEnum && len(in.Options) == 0 {
		return db.SubscriberFieldDefinition{}, ErrEnumNeedsOptions
	}
	if in.Label == "" {
		in.Label = in.Key
	}

	options, err := optionsToRaw(in.Options)
	if err != nil {
		return db.SubscriberFieldDefinition{}, err
	}

	return s.queries.CreateFieldDefinition(ctx, db.CreateFieldDefinitionParams{
		ProjectID: pid,
		Key:       in.Key,
		Label:     in.Label,
		FieldType: in.Type,
		Options:   options,
		Required:  in.Required,
	})
}

func (s *FieldDefinitionService) List(ctx context.Context, projectID string) ([]db.SubscriberFieldDefinition, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}
	return s.queries.ListFieldDefinitionsByProject(ctx, pid)
}

func (s *FieldDefinitionService) Update(ctx context.Context, projectID, fieldID string, label string, options []string, required bool) (db.SubscriberFieldDefinition, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return db.SubscriberFieldDefinition{}, errors.New("invalid project id")
	}
	fid, err := uuid.Parse(fieldID)
	if err != nil {
		return db.SubscriberFieldDefinition{}, errors.New("invalid field id")
	}

	existing, err := s.queries.GetFieldDefinition(ctx, db.GetFieldDefinitionParams{ID: fid, ProjectID: pid})
	if err != nil {
		return db.SubscriberFieldDefinition{}, err
	}
	if FieldType(existing.FieldType) == FieldTypeEnum && len(options) == 0 {
		return db.SubscriberFieldDefinition{}, ErrEnumNeedsOptions
	}
	if label == "" {
		label = existing.Label
	}

	rawOptions, err := optionsToRaw(options)
	if err != nil {
		return db.SubscriberFieldDefinition{}, err
	}

	return s.queries.UpdateFieldDefinition(ctx, db.UpdateFieldDefinitionParams{
		ID:        fid,
		ProjectID: pid,
		Label:     label,
		Options:   rawOptions,
		Required:  required,
	})
}

func (s *FieldDefinitionService) Delete(ctx context.Context, projectID, fieldID string) error {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return errors.New("invalid project id")
	}
	fid, err := uuid.Parse(fieldID)
	if err != nil {
		return errors.New("invalid field id")
	}
	return s.queries.DeleteFieldDefinition(ctx, db.DeleteFieldDefinitionParams{ID: fid, ProjectID: pid})
}

func (s *FieldDefinitionService) ValidateFields(ctx context.Context, projectID string, incoming map[string]any) (json.RawMessage, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project id")
	}

	defs, err := s.queries.ListFieldDefinitionsByProject(ctx, pid)
	if err != nil {
		return nil, err
	}

	byKey := make(map[string]db.SubscriberFieldDefinition, len(defs))
	for _, def := range defs {
		byKey[def.Key] = def
	}

	for key, value := range incoming {
		def, ok := byKey[key]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownField, key)
		}
		if err := validateFieldValue(def, value); err != nil {
			return nil, err
		}
	}

	for _, def := range defs {
		if !def.Required {
			continue
		}
		value, ok := incoming[def.Key]
		if !ok || isEmptyValue(value) {
			return nil, fmt.Errorf("%w: field %s is required", ErrFieldValidation, def.Key)
		}
	}

	if incoming == nil {
		incoming = map[string]any{}
	}
	return json.Marshal(incoming)
}

func validateFieldValue(def db.SubscriberFieldDefinition, value any) error {
	switch FieldType(def.FieldType) {
	case FieldTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%w: field %s must be a string", ErrFieldValidation, def.Key)
		}
	case FieldTypeNumber:
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("%w: field %s must be a number", ErrFieldValidation, def.Key)
		}
	case FieldTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%w: field %s must be a boolean", ErrFieldValidation, def.Key)
		}
	case FieldTypeDate:
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("%w: field %s must be a date string (YYYY-MM-DD)", ErrFieldValidation, def.Key)
		}
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return fmt.Errorf("%w: field %s must be a valid date (YYYY-MM-DD)", ErrFieldValidation, def.Key)
		}
	case FieldTypeEnum:
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("%w: field %s must be one of its allowed options", ErrFieldValidation, def.Key)
		}
		options := optionsFromRaw(def.Options)
		for _, opt := range options {
			if opt == s {
				return nil
			}
		}
		return fmt.Errorf("%w: field %s must be one of its allowed options", ErrFieldValidation, def.Key)
	}
	return nil
}

func isEmptyValue(value any) bool {
	if value == nil {
		return true
	}
	if s, ok := value.(string); ok {
		return s == ""
	}
	return false
}

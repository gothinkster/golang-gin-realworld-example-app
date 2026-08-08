package common

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"
)

func TestNullableUnmarshalJSON(t *testing.T) {
	asserts := assert.New(t)

	var payload struct {
		Name Nullable[string]   `json:"name"`
		Tags Nullable[[]string] `json:"tags"`
	}

	// Absent keys leave the zero value: not Set
	asserts.NoError(json.Unmarshal([]byte(`{}`), &payload))
	asserts.False(payload.Name.Set)
	asserts.False(payload.Tags.Set)
	asserts.False(payload.Name.IsInvalid())

	// Explicit null: Set, neither Valid nor Invalid
	asserts.NoError(json.Unmarshal([]byte(`{"name":null,"tags":null}`), &payload))
	asserts.True(payload.Name.Set)
	asserts.False(payload.Name.Valid)
	asserts.False(payload.Name.Invalid)
	asserts.True(payload.Tags.Set)
	asserts.False(payload.Tags.Valid)
	asserts.False(payload.Name.IsInvalid())

	// Decodable values: Set and Valid
	payload.Name = Nullable[string]{}
	payload.Tags = Nullable[[]string]{}
	asserts.NoError(json.Unmarshal([]byte(`{"name":"anna","tags":["a","b"]}`), &payload))
	asserts.True(payload.Name.Set)
	asserts.True(payload.Name.Valid)
	asserts.Equal("anna", payload.Name.Value)
	asserts.True(payload.Tags.Valid)
	asserts.Equal([]string{"a", "b"}, payload.Tags.Value)
	asserts.False(payload.Tags.IsInvalid())

	// Wrong JSON types: Set and Invalid, and the decode error is swallowed
	payload.Name = Nullable[string]{}
	payload.Tags = Nullable[[]string]{}
	asserts.NoError(json.Unmarshal([]byte(`{"name":123,"tags":"nope"}`), &payload))
	asserts.True(payload.Name.Set)
	asserts.True(payload.Name.Invalid)
	asserts.False(payload.Name.Valid)
	asserts.True(payload.Name.IsInvalid())
	asserts.True(payload.Tags.IsInvalid())

	// A mixed-type array is undecodable into []string
	payload.Tags = Nullable[[]string]{}
	asserts.NoError(json.Unmarshal([]byte(`{"tags":["ok",1]}`), &payload))
	asserts.True(payload.Tags.IsInvalid())
}

func TestNullableValuer(t *testing.T) {
	asserts := assert.New(t)

	value := func(v interface{}) interface{} {
		return nullableValuer(reflect.ValueOf(v))
	}

	// Strings: absent -> typed nil pointer, null/invalid -> "", value -> value
	asserts.Equal((*string)(nil), value(Nullable[string]{}))
	asserts.Equal("", value(Nullable[string]{Set: true}))
	asserts.Equal("", value(Nullable[string]{Set: true, Invalid: true}))
	asserts.Equal("hi", value(Nullable[string]{Set: true, Valid: true, Value: "hi"}))

	// Slices: absent -> typed nil pointer, null/invalid -> nullMarker, value -> value
	asserts.Equal((*[]string)(nil), value(Nullable[[]string]{}))
	asserts.Equal(nullMarker{}, value(Nullable[[]string]{Set: true}))
	asserts.Equal(nullMarker{}, value(Nullable[[]string]{Set: true, Invalid: true}))
	asserts.Equal([]string{"x"}, value(Nullable[[]string]{Set: true, Valid: true, Value: []string{"x"}}))

	// Unregistered instantiations project to nil
	asserts.Nil(value(Nullable[int]{Set: true, Valid: true, Value: 7}))
}

// Exercises the valuer and the "notnull" tag through gin's binding validator,
// the way request handlers actually consume Nullable fields.
func TestNullableBindingTags(t *testing.T) {
	asserts := assert.New(t)

	type form struct {
		Name Nullable[string]   `json:"name" binding:"omitnil,required,min=4"`
		Tags Nullable[[]string] `json:"tags" binding:"omitnil,notnull"`
	}

	bind := func(body string) error {
		var f form
		return binding.JSON.BindBody([]byte(body), &f)
	}

	// Absent fields are skipped by omitnil
	asserts.NoError(bind(`{}`))

	// Valid values pass their tags
	asserts.NoError(bind(`{"name":"anna","tags":["a"]}`))
	asserts.NoError(bind(`{"tags":[]}`))

	// Explicit null: strings fail "required", slices fail "notnull"
	asserts.Error(bind(`{"name":null}`))
	asserts.Error(bind(`{"tags":null}`))

	// Inner-value tags apply to decoded values
	asserts.Error(bind(`{"name":"abc"}`))

	// Wrong-typed values collapse to the null projection and fail too
	asserts.Error(bind(`{"name":123}`))
	asserts.Error(bind(`{"tags":"nope"}`))
}

func TestMarkInvalidFields(t *testing.T) {
	asserts := assert.New(t)

	errs := CommonError{Errors: map[string][]string{
		"title": {"can't be blank"},
		"body":  {"is too long"},
	}}
	errs.MarkInvalidFields([]string{"title", "tagList"})

	asserts.Equal([]string{"is invalid"}, errs.Errors["title"], "existing message should be overridden")
	asserts.Equal([]string{"is invalid"}, errs.Errors["tagList"], "missing field should be added")
	asserts.Equal([]string{"is too long"}, errs.Errors["body"], "unrelated field should be untouched")

	errs.MarkInvalidFields(nil)
	asserts.Len(errs.Errors, 3, "no-op on empty field list")
}

func TestNewValidatorErrorMessages(t *testing.T) {
	asserts := assert.New(t)

	// Non-validator errors collapse to a generic body error
	errs := NewValidatorError(errors.New("boom"))
	asserts.Equal([]string{"is invalid"}, errs.Errors["body"])

	type form struct {
		Name  Nullable[string]   `json:"name" binding:"omitnil,required,min=4,max=10"`
		Email string             `json:"email" binding:"omitempty,email"`
		Tags  Nullable[[]string] `json:"tags" binding:"omitnil,notnull"`
	}
	bindErrs := func(body string) CommonError {
		var f form
		return NewValidatorError(binding.JSON.BindBody([]byte(body), &f))
	}

	asserts.Equal([]string{"can't be blank"}, bindErrs(`{"name":null}`).Errors["name"])
	asserts.Equal([]string{"can't be null"}, bindErrs(`{"tags":null}`).Errors["tagList"],
		"notnull failures should read as null errors, on the renamed field")
	asserts.Equal([]string{"is too short (minimum is 4 characters)"}, bindErrs(`{"name":"abc"}`).Errors["name"])
	asserts.Equal([]string{"is too long (maximum is 10 characters)"}, bindErrs(`{"name":"aaaaaaaaaaaa"}`).Errors["name"])
	asserts.Equal([]string{"is invalid"}, bindErrs(`{"email":"nope"}`).Errors["email"],
		"tags without a specific message should fall back to is invalid")
}

package common

import (
	"encoding/json"
	"reflect"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Nullable is a tri-state JSON field: absent, null, or a value of type T.
// It exists because encoding/json cannot otherwise distinguish an absent key
// from an explicit null — both leave a *T nil. UnmarshalJSON is only invoked
// for keys present in the payload, so Set doubles as an unforgeable presence
// flag; no input can produce a Nullable that looks absent.
//
//	Set == false                  -> key absent (preserve current value)
//	Set && !Valid && !Invalid     -> explicit null
//	Set && Valid                  -> Value holds the decoded value
//	Set && Invalid                -> present but not decodable into T
type Nullable[T any] struct {
	Set     bool
	Valid   bool
	Invalid bool
	Value   T
}

// UnmarshalJSON records a decode failure in Invalid instead of returning an
// error, so handlers can report it as a field-level validation error rather
// than aborting the whole request decode.
func (n *Nullable[T]) UnmarshalJSON(b []byte) error {
	n.Set = true
	if string(b) == "null" {
		return nil
	}
	if err := json.Unmarshal(b, &n.Value); err != nil {
		n.Invalid = true
		return nil
	}
	n.Valid = true
	return nil
}

// IsInvalid reports whether the key was present but its value not decodable
// into T (wrong JSON type).
func (n Nullable[T]) IsInvalid() bool {
	return n.Set && n.Invalid
}

// nullMarker is what the valuer projects an explicit JSON null onto for
// slice-backed Nullable fields. A nil slice cannot represent null there:
// "omitnil" would skip it (nil slices count as absent) and an empty slice is
// a legitimate value. The marker is a private struct no JSON input can ever
// decode into, so only the registered "notnull" tag reacts to it.
type nullMarker struct{}

// nullableValuer teaches go-playground/validator to see through Nullable
// fields so binding tags apply to the inner value. The mapping encodes the
// tri-state semantics into values the tag vocabulary understands:
//
//	absent        -> typed nil pointer  ("omitnil" skips the field)
//	null/invalid  -> "" for strings     ("required" fails it)
//	                 nullMarker slices  ("notnull" fails it)
//	value         -> the value itself   (min/max/... apply)
func nullableValuer(field reflect.Value) interface{} {
	switch n := field.Interface().(type) {
	case Nullable[string]:
		if !n.Set {
			return (*string)(nil)
		}
		if !n.Valid {
			return ""
		}
		return n.Value
	case Nullable[[]string]:
		if !n.Set {
			return (*[]string)(nil)
		}
		if !n.Valid {
			return nullMarker{}
		}
		return n.Value
	}
	return nil
}

// notNullValidation implements the "notnull" binding tag: it fails only when
// the valuer projected an explicit null (or undecodable value). Note it is a
// no-op on Nullable[string] fields, whose null projects to "" for "required".
func notNullValidation(fl validator.FieldLevel) bool {
	_, isNull := fl.Field().Interface().(nullMarker)
	return !isNull
}

// Register the valuer and the "notnull" tag on gin's binding validator so
// `binding:` tags work on Nullable fields everywhere (server and tests alike).
func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterCustomTypeFunc(nullableValuer, Nullable[string]{}, Nullable[[]string]{})
		if err := v.RegisterValidation("notnull", notNullValidation); err != nil {
			panic(err)
		}
	}
}

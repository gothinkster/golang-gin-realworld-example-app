// Common tools and helper functions
package common

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// A helper function to generate random string
func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		randIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		b[i] = letters[randIdx.Int64()]
	}
	return string(b)
}

// A helper function to generate random int
func RandInt() int {
	randNum, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		panic(err)
	}
	return int(randNum.Int64())
}

// Keep this config private, it should not expose to open source
const JWTSecret = "A String Very Very Very Strong!!@##$!@#$" // #nosec G101

// A Util function to generate jwt_token which can be used in the request header
func GenToken(id uint) string {
	jwt_token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	// Sign and get the complete encoded token as a string
	token, err := jwt_token.SignedString([]byte(JWTSecret))
	if err != nil {
		fmt.Printf("failed to sign JWT token for id %d: %v\n", id, err)
		return ""
	}
	return token
}

// My own Error type that will help return my customized Error info
// following the RealWorld spec format:
//
//	{"errors": {"email": ["can't be blank"]}}
type CommonError struct {
	Errors map[string][]string `json:"errors"`
}

// Maps Go struct field names to their JSON counterparts for error responses.
var errorFieldNames = map[string]string{
	"Tags":    "tagList",
	"TagList": "tagList",
}

func errorFieldName(field string) string {
	if name, ok := errorFieldNames[field]; ok {
		return name
	}
	return strings.ToLower(field)
}

func errorMessageForTag(v validator.FieldError) string {
	switch v.Tag() {
	case "required":
		return "can't be blank"
	case "notnull":
		return "can't be null"
	case "min":
		return fmt.Sprintf("is too short (minimum is %v characters)", v.Param())
	case "max":
		return fmt.Sprintf("is too long (maximum is %v characters)", v.Param())
	default:
		return "is invalid"
	}
}

// To handle the error returned by c.Bind in gin framework
// https://github.com/go-playground/validator/blob/v9/_examples/translations/main.go
func NewValidatorError(err error) CommonError {
	res := CommonError{}
	res.Errors = make(map[string][]string)
	var errs validator.ValidationErrors
	if !errors.As(err, &errs) {
		res.Errors["body"] = []string{"is invalid"}
		return res
	}
	for _, v := range errs {
		field := errorFieldName(v.Field())
		res.Errors[field] = append(res.Errors[field], errorMessageForTag(v))
	}
	return res
}

// MarkInvalidFields overrides the messages of the given fields with
// "is invalid". Wrong-JSON-type values on Nullable fields are collapsed to
// blank/null by the valuer, so without this the binding tags would report
// them as "can't be blank" / "can't be null".
func (e CommonError) MarkInvalidFields(fields []string) {
	for _, field := range fields {
		e.Errors[field] = []string{"is invalid"}
	}
}

// Wrap the error info in an object
func NewError(key string, err error) CommonError {
	return NewErrorMessage(key, err.Error())
}

// Wrap a plain error message in the RealWorld errors format
func NewErrorMessage(key string, message string) CommonError {
	res := CommonError{}
	res.Errors = map[string][]string{key: {message}}
	return res
}

// Changed the c.MustBindWith() ->  c.ShouldBindWith().
// I don't want to auto return 400 when error happened.
// origin function is here: https://github.com/gin-gonic/gin/blob/master/context.go
func Bind(c *gin.Context, obj interface{}) error {
	b := binding.Default(c.Request.Method, c.ContentType())
	return c.ShouldBindWith(obj, b)
}

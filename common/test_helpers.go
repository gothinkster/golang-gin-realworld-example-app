package common

import (
	"fmt"
	"net/http"
)

// HeaderTokenMock sets the Authorization header with a mock token for the given user ID.
// This is shared across package tests to avoid duplication.
func HeaderTokenMock(req *http.Request, u uint) {
	req.Header.Set("Authorization", fmt.Sprintf("Token %v", GenToken(u)))
}

package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegister(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Register(tt.args.w, tt.args.r)
		})
	}

	req, err := http.NewRequest("GET", "/kurjun/rest/auth/register", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Register)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	expected := "Not allowed"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got '%v' want '%v'", rr.Body.String(), expected)
	}

}

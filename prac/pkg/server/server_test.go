package server

import (
	"prac/pkg/api"
	"testing"
)

func Test_server_loginUser(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		req  api.Request
		want api.Response
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var s server
			got := s.loginUser(tt.req)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("loginUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

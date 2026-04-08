package unit_tests

import (
	"errors"
	"testing"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/service"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		pw      string
		wantErr *apperror.AppError
	}{
		{
			name:    "empty password returns ErrPasswordTooShort",
			pw:      "",
			wantErr: apperror.ErrPasswordTooShort,
		},
		{
			name:    "password too short (11 chars) returns ErrPasswordTooShort",
			pw:      "Abcdefgh1!x",
			wantErr: apperror.ErrPasswordTooShort,
		},
		{
			name:    "password exactly 12 chars with 3 classes succeeds",
			pw:      "Abcdefghij1k",
			wantErr: nil,
		},
		{
			name:    "password with only 2 classes returns ErrPasswordComplexity",
			pw:      "abcdefghijklmn",
			wantErr: apperror.ErrPasswordComplexity,
		},
		{
			name:    "password with all 4 classes succeeds",
			pw:      "Abcdefgh1!zz",
			wantErr: nil,
		},
		{
			name:    "password with lowercase + uppercase + digits succeeds (3 of 4)",
			pw:      "Abcdefghij12",
			wantErr: nil,
		},
		{
			name:    "password with lowercase + uppercase only returns ErrPasswordComplexity (2 of 4)",
			pw:      "Abcdefghijkl",
			wantErr: apperror.ErrPasswordComplexity,
		},
		{
			name:    "very long password with 3 classes succeeds",
			pw:      "Abcdefghijklmnopqrstuvwxyz12345678901234567890",
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := service.ValidatePassword(tc.pw)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %v, got nil", tc.wantErr)
			}
			var appErr *apperror.AppError
			if !errors.As(err, &appErr) {
				t.Fatalf("expected *apperror.AppError, got %T: %v", err, err)
			}
			if appErr.Code != tc.wantErr.Code || appErr.Message != tc.wantErr.Message {
				t.Fatalf("expected error {Code:%d, Message:%q}, got {Code:%d, Message:%q}",
					tc.wantErr.Code, tc.wantErr.Message, appErr.Code, appErr.Message)
			}
		})
	}
}

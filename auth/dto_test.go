package auth

import (
	"strings"
	"testing"
)

func TestSignUpDTOIsValid(t *testing.T) {
	tests := []struct {
		name       string
		dto        SignUpDTO
		wantValid  bool
		wantErrors []string
	}{
		{
			name: "valid",
			dto: SignUpDTO{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			wantValid: true,
		},
		{
			name: "missing name",
			dto: SignUpDTO{
				Email:    "john@example.com",
				Password: "password123",
			},
			wantValid:  false,
			wantErrors: []string{"name"},
		},
		{
			name: "invalid email",
			dto: SignUpDTO{
				Name:     "John Doe",
				Email:    "not-an-email",
				Password: "password123",
			},
			wantValid:  false,
			wantErrors: []string{"email"},
		},
		{
			name: "password too short",
			dto: SignUpDTO{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "short",
			},
			wantValid:  false,
			wantErrors: []string{"password"},
		},
		{
			name: "empty image",
			dto: SignUpDTO{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Image:    strPtr(""),
			},
			wantValid:  false,
			wantErrors: []string{"image"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, validationErrors := tt.dto.IsValid()
			if valid != tt.wantValid {
				t.Fatalf("IsValid() valid = %v, want %v (errors: %v)", valid, tt.wantValid, validationErrors)
			}
			for _, key := range tt.wantErrors {
				if _, ok := validationErrors[key]; !ok {
					t.Errorf("IsValid() missing expected error for key %q, got %v", key, validationErrors)
				}
			}
		})
	}
}

func TestSignInDTOIsValid(t *testing.T) {
	valid := SignInDTO{Email: "john@example.com", Password: "password123"}
	ok, errors := valid.IsValid()
	if !ok || len(errors) != 0 {
		t.Fatalf("valid SignInDTO rejected: %v", errors)
	}

	invalid := SignInDTO{Email: "bad", Password: ""}
	ok, errors = invalid.IsValid()
	if ok {
		t.Fatalf("invalid SignInDTO accepted")
	}
	if _, ok := errors["email"]; !ok {
		t.Errorf("expected email error, got %v", errors)
	}
	if _, ok := errors["password"]; !ok {
		t.Errorf("expected password error, got %v", errors)
	}
}

func TestUpdateUserDTOIsValid(t *testing.T) {
	valid := UpdateUserDTO{Name: strPtr("Jane")}
	ok, errors := valid.IsValid()
	if !ok || len(errors) != 0 {
		t.Fatalf("valid UpdateUserDTO rejected: %v", errors)
	}

	invalid := UpdateUserDTO{Name: strPtr(strings.Repeat("a", nameMaxLen+1))}
	ok, errors = invalid.IsValid()
	if ok {
		t.Fatalf("invalid UpdateUserDTO accepted")
	}
	if _, ok := errors["name"]; !ok {
		t.Errorf("expected name error, got %v", errors)
	}
}

func TestUpdateSessionDTOIsValid(t *testing.T) {
	valid := UpdateSessionDTO{IPAddress: strPtr("203.0.113.10"), UserAgent: strPtr("curl/8.0")}
	ok, errors := valid.IsValid()
	if !ok || len(errors) != 0 {
		t.Fatalf("valid UpdateSessionDTO rejected: %v", errors)
	}

	nothing := UpdateSessionDTO{}
	ok, errors = nothing.IsValid()
	if ok {
		t.Fatalf("empty UpdateSessionDTO accepted")
	}
	if _, ok := errors["session"]; !ok {
		t.Errorf("expected session error, got %v", errors)
	}

	tooLong := UpdateSessionDTO{UserAgent: strPtr(strings.Repeat("a", 513))}
	ok, errors = tooLong.IsValid()
	if ok {
		t.Fatalf("invalid UpdateSessionDTO accepted")
	}
	if _, ok := errors["userAgent"]; !ok {
		t.Errorf("expected userAgent error, got %v", errors)
	}
}

func TestResetPasswordDTOIsValid(t *testing.T) {
	valid := ResetPasswordDTO{Token: "tok", NewPassword: "password123"}
	ok, errors := valid.IsValid()
	if !ok || len(errors) != 0 {
		t.Fatalf("valid ResetPasswordDTO rejected: %v", errors)
	}

	invalid := ResetPasswordDTO{Token: "", NewPassword: "short"}
	ok, errors = invalid.IsValid()
	if ok {
		t.Fatalf("invalid ResetPasswordDTO accepted")
	}
	if _, ok := errors["token"]; !ok {
		t.Errorf("expected token error, got %v", errors)
	}
	if _, ok := errors["password"]; !ok {
		t.Errorf("expected password error, got %v", errors)
	}
}

func TestChangeEmailDTOIsValid(t *testing.T) {
	valid := ChangeEmailDTO{NewEmail: "new@example.com"}
	ok, errors := valid.IsValid()
	if !ok || len(errors) != 0 {
		t.Fatalf("valid ChangeEmailDTO rejected: %v", errors)
	}

	invalid := ChangeEmailDTO{NewEmail: "not-an-email", CallbackURL: "not-a-url"}
	ok, errors = invalid.IsValid()
	if ok {
		t.Fatalf("invalid ChangeEmailDTO accepted")
	}
	if _, ok := errors["newEmail"]; !ok {
		t.Errorf("expected newEmail error, got %v", errors)
	}
	if _, ok := errors["callbackURL"]; !ok {
		t.Errorf("expected callbackURL error, got %v", errors)
	}
}

func TestVerifyPasswordDTOIsValid(t *testing.T) {
	valid := VerifyPasswordDTO{Password: "password123"}
	ok, errors := valid.IsValid()
	if !ok || len(errors) != 0 {
		t.Fatalf("valid VerifyPasswordDTO rejected: %v", errors)
	}

	invalid := VerifyPasswordDTO{}
	ok, errors = invalid.IsValid()
	if ok {
		t.Fatalf("empty VerifyPasswordDTO accepted")
	}
	if _, ok := errors["password"]; !ok {
		t.Errorf("expected password error, got %v", errors)
	}
}

func TestUsernameAvailableDTOIsValid(t *testing.T) {
	valid := UsernameAvailableDTO{Username: "john"}
	ok, errors := valid.IsValid()
	if !ok || len(errors) != 0 {
		t.Fatalf("valid UsernameAvailableDTO rejected: %v", errors)
	}

	invalid := UsernameAvailableDTO{Username: strings.Repeat("a", usernameMaxLen+1)}
	ok, errors = invalid.IsValid()
	if ok {
		t.Fatalf("invalid UsernameAvailableDTO accepted")
	}
	if _, ok := errors["username"]; !ok {
		t.Errorf("expected username error, got %v", errors)
	}
}

func TestSignInUsernameDTOIsValid(t *testing.T) {
	valid := SignInUsernameDTO{Username: "john", Password: "password123"}
	ok, errors := valid.IsValid()
	if !ok || len(errors) != 0 {
		t.Fatalf("valid SignInUsernameDTO rejected: %v", errors)
	}

	invalid := SignInUsernameDTO{Username: "", Password: ""}
	ok, errors = invalid.IsValid()
	if ok {
		t.Fatalf("invalid SignInUsernameDTO accepted")
	}
	if _, ok := errors["username"]; !ok {
		t.Errorf("expected username error, got %v", errors)
	}
	if _, ok := errors["password"]; !ok {
		t.Errorf("expected password error, got %v", errors)
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got := normalizeEmail("  John@Example.COM "); got != "john@example.com" {
		t.Errorf("normalizeEmail = %q, want %q", got, "john@example.com")
	}
}

func strPtr(s string) *string {
	return &s
}

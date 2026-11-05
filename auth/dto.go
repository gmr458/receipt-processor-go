package auth

import (
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/gmr458/receipt-processor/validator"
)

const (
	passwordMinLen = 8
	passwordMaxLen = 72
	nameMaxLen     = 100
	emailMaxLen    = 254
	imageMaxLen    = 2048
	usernameMaxLen = 50
)

type SignUpDTO struct {
	Name            string  `json:"name"`
	Email           string  `json:"email"`
	Password        string  `json:"password"`
	Image           *string `json:"image"`
	RememberMe      bool    `json:"rememberMe"`
	Username        *string `json:"username"`
	DisplayUsername *string `json:"displayUsername"`
}

type SignInDTO struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

type UpdateUserDTO struct {
	Name            *string `json:"name"`
	Image           *string `json:"image"`
	Username        *string `json:"username"`
	DisplayUsername *string `json:"displayUsername"`
}

type ChangePasswordDTO struct {
	CurrentPassword     string `json:"currentPassword"`
	NewPassword         string `json:"newPassword"`
	RevokeOtherSessions bool   `json:"revokeOtherSessions"`
}

type RequestPasswordResetDTO struct {
	Email      string `json:"email"`
	RedirectTo string `json:"redirectTo"`
}

type ResetPasswordDTO struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

type UpdateSessionDTO struct {
	IPAddress *string `json:"ipAddress"`
	UserAgent *string `json:"userAgent"`
}

type ChangeEmailDTO struct {
	NewEmail    string `json:"newEmail"`
	CallbackURL string `json:"callbackURL"`
}

type VerifyPasswordDTO struct {
	Password string `json:"password"`
}

type UsernameAvailableDTO struct {
	Username string `json:"username"`
}

type SignInUsernameDTO struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

func (dto SignUpDTO) IsValid() (bool, map[string]string) {
	v := validator.New()

	validateName(v, dto.Name, true)
	dto.validateEmail(v)
	validatePassword(v, dto.Password)
	if dto.Image != nil {
		v.Check(*dto.Image != "" && utf8.RuneCountInString(*dto.Image) <= imageMaxLen, "image", fmt.Sprintf("image must be at most %d characters", imageMaxLen))
	}
	if dto.Username != nil {
		v.Check(*dto.Username != "" && utf8.RuneCountInString(*dto.Username) <= usernameMaxLen, "username", fmt.Sprintf("username must be at most %d characters", usernameMaxLen))
	}
	if dto.DisplayUsername != nil {
		v.Check(*dto.DisplayUsername != "" && utf8.RuneCountInString(*dto.DisplayUsername) <= usernameMaxLen, "displayUsername", fmt.Sprintf("displayUsername must be at most %d characters", usernameMaxLen))
	}

	return v.Ok(), v.Errors
}

func (dto SignInDTO) IsValid() (bool, map[string]string) {
	v := validator.New()

	dto.validateEmail(v)
	v.Check(dto.Password != "", "password", "password cannot be empty")

	return v.Ok(), v.Errors
}

func (dto UpdateUserDTO) IsValid() (bool, map[string]string) {
	v := validator.New()

	if dto.Name != nil {
		validateName(v, *dto.Name, false)
	}
	if dto.Image != nil {
		v.Check(*dto.Image == "" || utf8.RuneCountInString(*dto.Image) <= imageMaxLen, "image", fmt.Sprintf("image must be at most %d characters", imageMaxLen))
	}
	if dto.Username != nil {
		v.Check(*dto.Username != "" && utf8.RuneCountInString(*dto.Username) <= usernameMaxLen, "username", fmt.Sprintf("username must be at most %d characters", usernameMaxLen))
	}
	if dto.DisplayUsername != nil {
		v.Check(*dto.DisplayUsername != "" && utf8.RuneCountInString(*dto.DisplayUsername) <= usernameMaxLen, "displayUsername", fmt.Sprintf("displayUsername must be at most %d characters", usernameMaxLen))
	}

	return v.Ok(), v.Errors
}

func (dto ChangePasswordDTO) IsValid() (bool, map[string]string) {
	v := validator.New()

	v.Check(dto.CurrentPassword != "", "currentPassword", "currentPassword cannot be empty")
	validatePassword(v, dto.NewPassword)
	v.Check(dto.NewPassword != dto.CurrentPassword, "newPassword", "newPassword must be different from currentPassword")

	return v.Ok(), v.Errors
}

func (dto RequestPasswordResetDTO) IsValid() (bool, map[string]string) {
	v := validator.New()

	v.Check(isValidEmail(dto.Email), "email", "invalid email address")
	v.Check(dto.RedirectTo == "" || isValidCallbackURL(dto.RedirectTo), "redirectTo", "redirectTo must be a valid URL")

	return v.Ok(), v.Errors
}

func (dto ResetPasswordDTO) IsValid() (bool, map[string]string) {
	v := validator.New()

	v.Check(dto.Token != "", "token", "token cannot be empty")
	validatePassword(v, dto.NewPassword)

	return v.Ok(), v.Errors
}

func (dto UpdateSessionDTO) IsValid() (bool, map[string]string) {
	v := validator.New()

	if dto.IPAddress == nil && dto.UserAgent == nil {
		v.AddError("session", "nothing to update")
	}
	if dto.IPAddress != nil && len(*dto.IPAddress) > 45 {
		v.AddError("ipAddress", "ipAddress must be at most 45 characters")
	}
	if dto.UserAgent != nil && len(*dto.UserAgent) > 512 {
		v.AddError("userAgent", "userAgent must be at most 512 characters")
	}

	return v.Ok(), v.Errors
}

func (dto ChangeEmailDTO) IsValid() (bool, map[string]string) {
	v := validator.New()

	v.Check(isValidEmail(dto.NewEmail), "newEmail", "invalid email address")
	v.Check(dto.CallbackURL == "" || isValidCallbackURL(dto.CallbackURL), "callbackURL", "callbackURL must be a valid URL")

	return v.Ok(), v.Errors
}

func (dto VerifyPasswordDTO) IsValid() (bool, map[string]string) {
	v := validator.New()

	v.Check(dto.Password != "", "password", "password cannot be empty")

	return v.Ok(), v.Errors
}

func (dto UsernameAvailableDTO) IsValid() (bool, map[string]string) {
	v := validator.New()

	v.Check(dto.Username != "" && utf8.RuneCountInString(dto.Username) <= usernameMaxLen, "username", fmt.Sprintf("username must be at most %d characters", usernameMaxLen))

	return v.Ok(), v.Errors
}

func (dto SignInUsernameDTO) IsValid() (bool, map[string]string) {
	v := validator.New()

	v.Check(dto.Username != "" && utf8.RuneCountInString(dto.Username) <= usernameMaxLen, "username", fmt.Sprintf("username must be at most %d characters", usernameMaxLen))
	v.Check(dto.Password != "", "password", "password cannot be empty")

	return v.Ok(), v.Errors
}

func validateName(v *validator.Validator, name string, required bool) {
	const key = "name"

	if name == "" {
		if required {
			v.AddError(key, "name cannot be empty")
		}
		return
	}
	if utf8.RuneCountInString(name) > nameMaxLen {
		v.AddError(key, fmt.Sprintf("name must be at most %d characters", nameMaxLen))
	}
}

func (dto SignUpDTO) validateEmail(v *validator.Validator) {
	v.Check(isValidEmail(dto.Email), "email", "invalid email address")
}

func validatePassword(v *validator.Validator, password string) {
	const key = "password"

	v.Check(len(password) >= passwordMinLen, key, fmt.Sprintf("password must be at least %d characters", passwordMinLen))
	v.Check(len(password) <= passwordMaxLen, key, fmt.Sprintf("password must be at most %d characters", passwordMaxLen))
}

func (dto SignInDTO) validateEmail(v *validator.Validator) {
	v.Check(isValidEmail(dto.Email), "email", "invalid email address")
}

func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" || len(email) > emailMaxLen {
		return false
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	return addr.Address == email
}

func isValidCallbackURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}

	return u.IsAbs()
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/gmr458/receipt-processor/errs"
)

const (
	passwordBcryptCost = 12
	sessionCacheTTL    = time.Minute
	verificationTTL    = time.Hour
)

type Config struct {
	SessionDuration    time.Duration
	RememberMeDuration time.Duration
	SessionCacheTTL    time.Duration
	VerificationExpiry time.Duration
	DefaultCallbackURL string
}

type Service struct {
	users         UserRepository
	sessions      SessionRepository
	accounts      AccountRepository
	verifications VerificationRepository
	cache         SessionCache
	mailer        Mailer
	config        Config
}

func NewService(
	users UserRepository,
	sessions SessionRepository,
	accounts AccountRepository,
	verifications VerificationRepository,
	cache SessionCache,
	mailer Mailer,
	config Config,
) Service {
	if config.SessionDuration <= 0 {
		config.SessionDuration = 7 * 24 * time.Hour
	}
	if config.RememberMeDuration <= 0 {
		config.RememberMeDuration = 30 * 24 * time.Hour
	}
	if config.SessionCacheTTL <= 0 {
		config.SessionCacheTTL = sessionCacheTTL
	}
	if config.VerificationExpiry <= 0 {
		config.VerificationExpiry = verificationTTL
	}

	return Service{
		users:         users,
		sessions:      sessions,
		accounts:      accounts,
		verifications: verifications,
		cache:         cache,
		mailer:        mailer,
		config:        config,
	}
}

func (s *Service) SignUpEmail(
	ctx context.Context,
	dto SignUpDTO,
	ipAddress string,
	userAgent string,
) (*User, *Session, error) {
	isValid, validationErrors := dto.IsValid()
	if !isValid {
		return nil, nil, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field/s",
			Details: validationErrors,
		}
	}

	email := normalizeEmail(dto.Email)

	_, err := s.users.FindByEmail(ctx, email)
	if nil == err {
		return nil, nil, &errs.Error{
			Code:    errs.EUNPROCESSABLECONTENT,
			Message: "User already exists. Try signing in instead.",
		}
	}
	if errs.ErrorCode(err) != errs.ENOTFOUND {
		return nil, nil, err
	}

	now := time.Now()

	user := &User{
		ID:              uuid.Must(uuid.NewV7()).String(),
		Name:            dto.Name,
		Email:           email,
		EmailVerified:   false,
		Image:           dto.Image,
		CreatedAt:       now,
		UpdatedAt:       now,
		Username:        dto.Username,
		DisplayUsername: dto.DisplayUsername,
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(dto.Password),
		passwordBcryptCost,
	)
	if err != nil {
		return nil, nil, errs.Errorf(errs.EINTERNAL, "failed to hash password: %v", err)
	}
	hash := string(passwordHash)

	account := &Account{
		ID:         uuid.Must(uuid.NewV7()).String(),
		AccountID:  user.ID,
		ProviderID: CredentialProviderID,
		UserID:     user.ID,
		Password:   &hash,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	session, err := s.newSession(user.ID, ipAddress, userAgent, dto.RememberMe)
	if err != nil {
		return nil, nil, err
	}

	if err := s.users.CreateWithCredential(ctx, user, account, session); err != nil {
		return nil, nil, err
	}

	_ = s.cache.Set(ctx, session, s.config.SessionCacheTTL)

	return user, session, nil
}

func (s *Service) SignInEmail(
	ctx context.Context,
	dto SignInDTO,
	ipAddress string,
	userAgent string,
) (*User, *Session, error) {
	isValid, validationErrors := dto.IsValid()
	if !isValid {
		return nil, nil, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field/s",
			Details: validationErrors,
		}
	}

	email := normalizeEmail(dto.Email)

	account, user, err := s.accounts.FindCredentialByEmail(ctx, email)
	if err != nil {
		if errs.ErrorCode(err) == errs.ENOTFOUND {
			return nil, nil, &errs.Error{Code: errs.EUNAUTHORIZED, Message: "Invalid email or password"}
		}
		return nil, nil, err
	}
	if account.Password == nil {
		return nil, nil, &errs.Error{Code: errs.EUNAUTHORIZED, Message: "Invalid email or password"}
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(*account.Password),
		[]byte(dto.Password),
	)
	if err != nil {
		return nil, nil, &errs.Error{Code: errs.EUNAUTHORIZED, Message: "Invalid email or password"}
	}

	session, err := s.newSession(user.ID, ipAddress, userAgent, dto.RememberMe)
	if err != nil {
		return nil, nil, err
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, nil, err
	}

	_ = s.cache.Set(ctx, session, s.config.SessionCacheTTL)

	return user, session, nil
}

func (s *Service) SignInUsername(
	ctx context.Context,
	dto SignInUsernameDTO,
	ipAddress string,
	userAgent string,
) (*User, *Session, error) {
	isValid, validationErrors := dto.IsValid()
	if !isValid {
		return nil, nil, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field/s",
			Details: validationErrors,
		}
	}

	username := normalizeUsername(dto.Username)

	user, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		if errs.ErrorCode(err) == errs.ENOTFOUND {
			return nil, nil, &errs.Error{Code: errs.EUNAUTHORIZED, Message: "Invalid username or password"}
		}
		return nil, nil, err
	}

	account, _, err := s.accounts.FindCredentialByEmail(ctx, user.Email)
	if err != nil {
		if errs.ErrorCode(err) == errs.ENOTFOUND {
			return nil, nil, &errs.Error{Code: errs.EUNAUTHORIZED, Message: "Invalid username or password"}
		}
		return nil, nil, err
	}
	if account.Password == nil {
		return nil, nil, &errs.Error{Code: errs.EUNAUTHORIZED, Message: "Invalid username or password"}
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(*account.Password),
		[]byte(dto.Password),
	)
	if err != nil {
		return nil, nil, &errs.Error{Code: errs.EUNAUTHORIZED, Message: "Invalid username or password"}
	}

	session, err := s.newSession(user.ID, ipAddress, userAgent, dto.RememberMe)
	if err != nil {
		return nil, nil, err
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, nil, err
	}

	_ = s.cache.Set(ctx, session, s.config.SessionCacheTTL)

	return user, session, nil
}

func (s *Service) ChangeEmail(
	ctx context.Context,
	userID string,
	dto ChangeEmailDTO,
) (*User, error) {
	isValid, validationErrors := dto.IsValid()
	if !isValid {
		return nil, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field/s",
			Details: validationErrors,
		}
	}

	user, err := s.users.FindById(ctx, userID)
	if err != nil {
		return nil, err
	}

	newEmail := normalizeEmail(dto.NewEmail)
	if newEmail == user.Email {
		return nil, &errs.Error{
			Code:    errs.EUNPROCESSABLECONTENT,
			Message: "New email is the same as the current email",
		}
	}

	_, err = s.users.FindByEmail(ctx, newEmail)
	if nil == err {
		return nil, &errs.Error{Code: errs.ECONFLICT, Message: "Email already in use"}
	}
	if errs.ErrorCode(err) != errs.ENOTFOUND {
		return nil, err
	}

	if err := s.users.UpdateEmail(ctx, user.ID, newEmail); err != nil {
		return nil, err
	}

	user.Email = newEmail
	user.EmailVerified = false
	user.UpdatedAt = time.Now()

	return user, nil
}

func (s *Service) VerifyPassword(
	ctx context.Context,
	userID string,
	dto VerifyPasswordDTO,
) (bool, error) {
	isValid, validationErrors := dto.IsValid()
	if !isValid {
		return false, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field/s",
			Details: validationErrors,
		}
	}

	user, err := s.users.FindById(ctx, userID)
	if err != nil {
		return false, err
	}

	account, _, err := s.accounts.FindCredentialByEmail(ctx, user.Email)
	if err != nil {
		if errs.ErrorCode(err) == errs.ENOTFOUND {
			return false, nil
		}
		return false, err
	}
	if account.Password == nil {
		return false, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(*account.Password), []byte(dto.Password))

	return err == nil, nil
}

func (s *Service) IsUsernameAvailable(
	ctx context.Context,
	dto UsernameAvailableDTO,
) (bool, error) {
	isValid, validationErrors := dto.IsValid()
	if !isValid {
		return false, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field/s",
			Details: validationErrors,
		}
	}

	username := normalizeUsername(dto.Username)

	_, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		if errs.ErrorCode(err) == errs.ENOTFOUND {
			return true, nil
		}
		return false, err
	}

	return false, nil
}

func (s *Service) GetSession(
	ctx context.Context,
	token string,
) (*User, *Session, error) {
	if token == "" {
		return nil, nil, nil
	}

	cached, err := s.cache.Get(ctx, token)
	if nil == err && cached != nil && !cached.IsExpired() {
		user, err := s.users.FindById(ctx, cached.UserID)
		if err != nil {
			return nil, nil, err
		}
		return user, cached, nil
	}

	session, err := s.sessions.FindByToken(ctx, token)
	if err != nil {
		if errs.ErrorCode(err) == errs.ENOTFOUND {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	if session.IsExpired() {
		return nil, nil, nil
	}

	user, err := s.users.FindById(ctx, session.UserID)
	if err != nil {
		return nil, nil, err
	}

	_ = s.cache.Set(ctx, session, s.config.SessionCacheTTL)

	return user, session, nil
}

func (s *Service) SignOut(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}

	_ = s.cache.Delete(ctx, token)

	err := s.sessions.DeleteByToken(ctx, token)
	if err != nil && errs.ErrorCode(err) != errs.ENOTFOUND {
		return err
	}

	return nil
}

func (s *Service) ListSessions(ctx context.Context, userID string) ([]Session, error) {
	sessions, err := s.sessions.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	active := make([]Session, 0, len(sessions))
	for _, session := range sessions {
		if !session.IsExpired() {
			active = append(active, session)
		}
	}

	return active, nil
}

func (s *Service) RevokeSession(ctx context.Context, userID string, token string) error {
	if token == "" {
		return errs.Errorf(errs.EINVALID, "token cannot be empty")
	}

	session, err := s.sessions.FindByToken(ctx, token)
	if err != nil {
		return err
	}
	if session.UserID != userID {
		return &errs.Error{Code: errs.ENOTFOUND, Message: "Session not found"}
	}

	_ = s.cache.Delete(ctx, token)

	return s.sessions.DeleteByToken(ctx, token)
}

func (s *Service) UpdateSession(
	ctx context.Context,
	token string,
	dto UpdateSessionDTO,
) (*Session, error) {
	isValid, validationErrors := dto.IsValid()
	if !isValid {
		return nil, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field/s",
			Details: validationErrors,
		}
	}

	session, err := s.sessions.FindByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if session.IsExpired() {
		return nil, &errs.Error{Code: errs.EUNAUTHORIZED, Message: "Session expired"}
	}

	if dto.IPAddress != nil {
		session.IPAddress = dto.IPAddress
	}
	if dto.UserAgent != nil {
		session.UserAgent = dto.UserAgent
	}
	session.UpdatedAt = time.Now()

	if err := s.sessions.Update(ctx, session); err != nil {
		return nil, err
	}

	_ = s.cache.Delete(ctx, token)

	return session, nil
}

func (s *Service) RevokeAllSessions(ctx context.Context, userID string) error {
	_ = s.cache.DeleteByUserID(ctx, userID)

	if err := s.sessions.DeleteByUserID(ctx, userID); err != nil {
		return err
	}

	return nil
}

func (s *Service) RevokeOtherSessions(ctx context.Context, userID string, currentToken string) error {
	return s.revokeSessionsExcept(ctx, userID, currentToken)
}

func (s *Service) UpdateUser(
	ctx context.Context,
	userID string,
	dto UpdateUserDTO,
) (*User, error) {
	isValid, validationErrors := dto.IsValid()
	if !isValid {
		return nil, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field/s",
			Details: validationErrors,
		}
	}

	user, err := s.users.FindById(ctx, userID)
	if err != nil {
		return nil, err
	}

	if dto.Name != nil {
		user.Name = *dto.Name
	}
	if dto.Image != nil {
		if *dto.Image == "" {
			user.Image = nil
		} else {
			user.Image = dto.Image
		}
	}
	if dto.Username != nil {
		user.Username = dto.Username
	}
	if dto.DisplayUsername != nil {
		user.DisplayUsername = dto.DisplayUsername
	}
	user.UpdatedAt = time.Now()

	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) DeleteUser(ctx context.Context, userID string, password *string) error {
	user, err := s.users.FindById(ctx, userID)
	if err != nil {
		return err
	}

	if password != nil {
		account, _, err := s.accounts.FindCredentialByEmail(ctx, user.Email)
		if err != nil {
			if errs.ErrorCode(err) == errs.ENOTFOUND {
				return &errs.Error{Code: errs.EUNAUTHORIZED, Message: "Invalid password"}
			}
			return err
		}
		if account.Password == nil {
			return &errs.Error{Code: errs.EUNAUTHORIZED, Message: "Invalid password"}
		}
		if err := bcrypt.CompareHashAndPassword(
			[]byte(*account.Password),
			[]byte(*password),
		); err != nil {
			return &errs.Error{Code: errs.EUNAUTHORIZED, Message: "Invalid password"}
		}
	}

	_ = s.cache.DeleteByUserID(ctx, userID)

	return s.users.DeleteById(ctx, userID)
}

func (s *Service) ChangePassword(
	ctx context.Context,
	userID string,
	currentToken string,
	dto ChangePasswordDTO,
) (*User, error) {
	isValid, validationErrors := dto.IsValid()
	if !isValid {
		return nil, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field/s",
			Details: validationErrors,
		}
	}

	user, err := s.users.FindById(ctx, userID)
	if err != nil {
		return nil, err
	}

	account, _, err := s.accounts.FindCredentialByEmail(ctx, user.Email)
	if err != nil {
		if errs.ErrorCode(err) == errs.ENOTFOUND {
			return nil, &errs.Error{Code: errs.EUNAUTHORIZED, Message: "Invalid current password"}
		}
		return nil, err
	}
	if account.Password == nil {
		return nil, &errs.Error{Code: errs.EUNAUTHORIZED, Message: "Invalid current password"}
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(*account.Password),
		[]byte(dto.CurrentPassword),
	); err != nil {
		return nil, &errs.Error{Code: errs.EUNAUTHORIZED, Message: "Invalid current password"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(dto.NewPassword), passwordBcryptCost)
	if err != nil {
		return nil, errs.Errorf(errs.EINTERNAL, "failed to hash password: %v", err)
	}

	if err := s.accounts.UpdatePasswordByUserID(ctx, user.ID, string(hash)); err != nil {
		return nil, err
	}

	if dto.RevokeOtherSessions {
		if err := s.revokeSessionsExcept(ctx, user.ID, currentToken); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, dto RequestPasswordResetDTO) error {
	isValid, validationErrors := dto.IsValid()
	if !isValid {
		return &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field/s",
			Details: validationErrors,
		}
	}

	email := normalizeEmail(dto.Email)

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errs.ErrorCode(err) == errs.ENOTFOUND {
			return nil
		}
		return err
	}

	token := rand.Text()

	identifier := resetPasswordIdentifier(user.ID)
	now := time.Now()

	_ = s.verifications.DeleteByIdentifier(ctx, identifier)

	verification := &Verification{
		ID:         uuid.Must(uuid.NewV7()).String(),
		Identifier: identifier,
		Value:      token,
		ExpiresAt:  now.Add(s.config.VerificationExpiry),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.verifications.Create(ctx, verification); err != nil {
		return err
	}

	callbackURL := dto.RedirectTo
	if callbackURL == "" {
		callbackURL = s.config.DefaultCallbackURL
	}
	if callbackURL == "" {
		return &errs.Error{Code: errs.EINVALID, Message: "redirectTo is required"}
	}
	link := fmt.Sprintf("%s?token=%s", callbackURL, token)

	return s.mailer.SendPasswordReset(ctx, user.Email, link)
}

func (s *Service) ValidateResetToken(ctx context.Context, token string) (bool, error) {
	verification, err := s.verifications.FindByValue(ctx, token)
	if err != nil {
		if errs.ErrorCode(err) == errs.ENOTFOUND {
			return false, nil
		}
		return false, err
	}

	return !verification.IsExpired() &&
		strings.HasPrefix(verification.Identifier, resetPasswordIdentifierPrefix), nil
}

func (s *Service) ResetPassword(ctx context.Context, dto ResetPasswordDTO) error {
	isValid, validationErrors := dto.IsValid()
	if !isValid {
		return &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field/s",
			Details: validationErrors,
		}
	}

	verification, err := s.verifications.FindByValue(ctx, dto.Token)
	if err != nil {
		if errs.ErrorCode(err) == errs.ENOTFOUND {
			return &errs.Error{Code: errs.EINVALID, Message: "Invalid or expired token"}
		}
		return err
	}
	if verification.IsExpired() ||
		!strings.HasPrefix(verification.Identifier, resetPasswordIdentifierPrefix) {
		return &errs.Error{Code: errs.EINVALID, Message: "Invalid or expired token"}
	}

	userID := strings.TrimPrefix(verification.Identifier, resetPasswordIdentifierPrefix)

	user, err := s.users.FindById(ctx, userID)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(dto.NewPassword), passwordBcryptCost)
	if err != nil {
		return errs.Errorf(errs.EINTERNAL, "failed to hash password: %v", err)
	}

	if err := s.accounts.UpdatePasswordByUserID(ctx, user.ID, string(hash)); err != nil {
		return err
	}

	_ = s.cache.DeleteByUserID(ctx, user.ID)

	if err := s.sessions.DeleteByUserID(ctx, user.ID); err != nil {
		return err
	}

	return s.verifications.DeleteByIdentifier(ctx, verification.Identifier)
}

func (s *Service) SendVerificationEmail(ctx context.Context, userID string, callbackURL string) error {
	user, err := s.users.FindById(ctx, userID)
	if err != nil {
		return err
	}
	if user.EmailVerified {
		return nil
	}

	token := rand.Text()

	identifier := emailVerificationIdentifier(user.ID)
	now := time.Now()

	_ = s.verifications.DeleteByIdentifier(ctx, identifier)

	verification := &Verification{
		ID:         uuid.Must(uuid.NewV7()).String(),
		Identifier: identifier,
		Value:      token,
		ExpiresAt:  now.Add(s.config.VerificationExpiry),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.verifications.Create(ctx, verification); err != nil {
		return err
	}

	if callbackURL == "" {
		callbackURL = s.config.DefaultCallbackURL
	}
	if callbackURL == "" {
		return &errs.Error{Code: errs.EINVALID, Message: "callbackURL is required"}
	}
	link := fmt.Sprintf("%s?token=%s", callbackURL, token)

	return s.mailer.SendVerificationEmail(ctx, user.Email, link)
}

func (s *Service) VerifyEmail(ctx context.Context, token string) (*User, error) {
	verification, err := s.verifications.FindByValue(ctx, token)
	if err != nil {
		if errs.ErrorCode(err) == errs.ENOTFOUND {
			return nil, &errs.Error{Code: errs.EINVALID, Message: "Invalid or expired token"}
		}
		return nil, err
	}
	if verification.IsExpired() ||
		!strings.HasPrefix(verification.Identifier, emailVerificationIdentifierPrefix) {
		return nil, &errs.Error{Code: errs.EINVALID, Message: "Invalid or expired token"}
	}

	userID := strings.TrimPrefix(verification.Identifier, emailVerificationIdentifierPrefix)

	user, err := s.users.FindById(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.EmailVerified {
		if err := s.users.SetEmailVerified(ctx, user.ID, true); err != nil {
			return nil, err
		}
		user.EmailVerified = true
		user.UpdatedAt = time.Now()
	}

	if err := s.verifications.DeleteByIdentifier(ctx, verification.Identifier); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) revokeSessionsExcept(ctx context.Context, userID string, keepToken string) error {
	sessions, err := s.sessions.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		if session.Token == keepToken {
			continue
		}
		_ = s.cache.Delete(ctx, session.Token)
		if err := s.sessions.DeleteByToken(ctx, session.Token); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) newSession(
	userID string,
	ipAddress string,
	userAgent string,
	rememberMe bool,
) (*Session, error) {
	token := rand.Text()

	duration := s.config.SessionDuration
	if rememberMe {
		duration = s.config.RememberMeDuration
	}

	now := time.Now()

	var ip *string
	if ipAddress != "" {
		ip = &ipAddress
	}
	var ua *string
	if userAgent != "" {
		ua = &userAgent
	}

	return &Session{
		ID:        uuid.Must(uuid.NewV7()).String(),
		ExpiresAt: now.Add(duration),
		Token:     token,
		CreatedAt: now,
		UpdatedAt: now,
		IPAddress: ip,
		UserAgent: ua,
		UserID:    userID,
	}, nil
}

package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/gmr458/receipt-processor/auth"
	"github.com/gmr458/receipt-processor/errs"
)

type UserRepository struct {
	conn *Conn
}

type SessionRepository struct {
	conn *Conn
}

type AccountRepository struct {
	conn *Conn
}

type VerificationRepository struct {
	conn *Conn
}

func scanUser(row pgx.Row) (*auth.User, error) {
	var u auth.User
	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.EmailVerified,
		&u.Image,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.Role,
		&u.Banned,
		&u.BanReason,
		&u.BanExpires,
		&u.Username,
		&u.DisplayUsername,
		&u.TwoFactorEnabled,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r UserRepository) CreateWithCredential(
	ctx context.Context,
	user *auth.User,
	account *auth.Account,
	session *auth.Session,
) error {
	tx, err := r.conn.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
        INSERT INTO "user" (
            id,
            name,
            email,
            email_verified,
            image,
            created_at,
            updated_at,
            role,
            banned,
            ban_reason,
            ban_expires,
            username,
            display_username,
            two_factor_enabled
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
    `
	_, err = tx.Exec(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.EmailVerified,
		user.Image,
		user.CreatedAt,
		user.UpdatedAt,
		user.Role,
		user.Banned,
		user.BanReason,
		user.BanExpires,
		user.Username,
		user.DisplayUsername,
		user.TwoFactorEnabled,
	)
	if err != nil {
		return mapUniqueViolation(err, "email already in use")
	}

	query = `
        INSERT INTO "account" (
            id,
            issuer,
            account_id,
            provider_id,
            user_id,
            access_token,
            refresh_token,
            id_token,
            access_token_expires_at,
            refresh_token_expires_at,
            scope,
            password,
            created_at,
            updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
    `
	_, err = tx.Exec(ctx, query,
		account.ID,
		account.Issuer,
		account.AccountID,
		account.ProviderID,
		account.UserID,
		account.AccessToken,
		account.RefreshToken,
		account.IDToken,
		account.AccessTokenExpiresAt,
		account.RefreshTokenExpiresAt,
		account.Scope,
		account.Password,
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		return err
	}

	query = `
        INSERT INTO "session" (
            id,
            expires_at,
            token,
            created_at,
            updated_at,
            ip_address,
            user_agent,
            user_id,
            impersonated_by
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
	_, err = tx.Exec(ctx, query,
		session.ID,
		session.ExpiresAt,
		session.Token,
		session.CreatedAt,
		session.UpdatedAt,
		session.IPAddress,
		session.UserAgent,
		session.UserID,
		session.ImpersonatedBy,
	)
	if err != nil {
		return mapUniqueViolation(err, "session token already in use")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r UserRepository) FindByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `
        SELECT
            id,
            name,
            email,
            email_verified,
            image,
            created_at,
            updated_at,
            role,
            banned,
            ban_reason,
            ban_expires,
            username,
            display_username,
            two_factor_enabled
        FROM "user"
        WHERE email = $1
    `
	row := r.conn.Pool.QueryRow(ctx, query, email)
	user, err := scanUser(row)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, &errs.Error{Code: errs.ENOTFOUND, Message: "User not found"}
		default:
			return nil, err
		}
	}
	return user, nil
}

func (r UserRepository) FindByUsername(ctx context.Context, username string) (*auth.User, error) {
	query := `
        SELECT
            id,
            name,
            email,
            email_verified,
            image,
            created_at,
            updated_at,
            role,
            banned,
            ban_reason,
            ban_expires,
            username,
            display_username,
            two_factor_enabled
        FROM "user"
        WHERE username = $1
    `
	row := r.conn.Pool.QueryRow(ctx, query, username)
	user, err := scanUser(row)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, &errs.Error{Code: errs.ENOTFOUND, Message: "User not found"}
		default:
			return nil, err
		}
	}
	return user, nil
}

func (r UserRepository) FindById(ctx context.Context, id string) (*auth.User, error) {
	query := `
        SELECT
            id,
            name,
            email,
            email_verified,
            image,
            created_at,
            updated_at,
            role,
            banned,
            ban_reason,
            ban_expires,
            username,
            display_username,
            two_factor_enabled
        FROM "user"
        WHERE id = $1
    `
	row := r.conn.Pool.QueryRow(ctx, query, id)
	user, err := scanUser(row)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, &errs.Error{Code: errs.ENOTFOUND, Message: "User not found"}
		default:
			return nil, err
		}
	}
	return user, nil
}

func (r UserRepository) Update(ctx context.Context, user *auth.User) error {
	query := `
        UPDATE "user"
        SET
            name = $1,
            image = $2,
            username = $3,
            display_username = $4,
            updated_at = $5
        WHERE id = $6
    `
	_, err := r.conn.Pool.Exec(ctx, query,
		user.Name,
		user.Image,
		user.Username,
		user.DisplayUsername,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return mapUniqueViolation(err, "username already in use")
	}
	return nil
}

func (r UserRepository) DeleteById(ctx context.Context, id string) error {
	_, err := r.conn.Pool.Exec(ctx, `DELETE FROM "user" WHERE id = $1`, id)
	if err != nil {
		return err
	}
	return nil
}

func scanSession(row pgx.Row) (*auth.Session, error) {
	var s auth.Session
	err := row.Scan(
		&s.ID,
		&s.ExpiresAt,
		&s.Token,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.IPAddress,
		&s.UserAgent,
		&s.UserID,
		&s.ImpersonatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r SessionRepository) Create(ctx context.Context, session *auth.Session) error {
	query := `
        INSERT INTO "session" (
            id,
            expires_at,
            token,
            created_at,
            updated_at,
            ip_address,
            user_agent,
            user_id,
            impersonated_by
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
	_, err := r.conn.Pool.Exec(ctx, query,
		session.ID,
		session.ExpiresAt,
		session.Token,
		session.CreatedAt,
		session.UpdatedAt,
		session.IPAddress,
		session.UserAgent,
		session.UserID,
		session.ImpersonatedBy,
	)
	if err != nil {
		return mapUniqueViolation(err, "session token already in use")
	}
	return nil
}

func (r SessionRepository) FindByToken(ctx context.Context, token string) (*auth.Session, error) {
	query := `
        SELECT
            id,
            expires_at,
            token,
            created_at,
            updated_at,
            ip_address,
            user_agent,
            user_id,
            impersonated_by
        FROM "session"
        WHERE token = $1
    `
	row := r.conn.Pool.QueryRow(ctx, query, token)
	session, err := scanSession(row)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, &errs.Error{Code: errs.ENOTFOUND, Message: "Session not found"}
		default:
			return nil, err
		}
	}
	return session, nil
}

func (r SessionRepository) FindByUserID(ctx context.Context, userID string) ([]auth.Session, error) {
	query := `
        SELECT
            id,
            expires_at,
            token,
            created_at,
            updated_at,
            ip_address,
            user_agent,
            user_id,
            impersonated_by
        FROM "session"
        WHERE user_id = $1
        ORDER BY created_at DESC
    `
	rows, err := r.conn.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := []auth.Session{}
	for rows.Next() {
		var s auth.Session
		err = rows.Scan(
			&s.ID,
			&s.ExpiresAt,
			&s.Token,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.IPAddress,
			&s.UserAgent,
			&s.UserID,
			&s.ImpersonatedBy,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r SessionRepository) DeleteByToken(ctx context.Context, token string) error {
	_, err := r.conn.Pool.Exec(ctx, `DELETE FROM "session" WHERE token = $1`, token)
	if err != nil {
		return err
	}
	return nil
}

func (r SessionRepository) Update(ctx context.Context, session *auth.Session) error {
	query := `
        UPDATE "session"
        SET
            ip_address = $1,
            user_agent = $2,
            updated_at = $3
        WHERE id = $4
    `
	_, err := r.conn.Pool.Exec(ctx, query,
		session.IPAddress,
		session.UserAgent,
		session.UpdatedAt,
		session.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r SessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.conn.Pool.Exec(ctx, `DELETE FROM "session" WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	return nil
}

func (r AccountRepository) FindCredentialByEmail(
	ctx context.Context,
	email string,
) (*auth.Account, *auth.User, error) {
	query := `
        SELECT
            a.id,
            a.issuer,
            a.account_id,
            a.provider_id,
            a.user_id,
            a.access_token,
            a.refresh_token,
            a.id_token,
            a.access_token_expires_at,
            a.refresh_token_expires_at,
            a.scope,
            a.password,
            a.created_at,
            a.updated_at,
            u.id,
            u.name,
            u.email,
            u.email_verified,
            u.image,
            u.created_at,
            u.updated_at,
            u.role,
            u.banned,
            u.ban_reason,
            u.ban_expires,
            u.username,
            u.display_username,
            u.two_factor_enabled
        FROM "account" a
        JOIN "user" u ON u.id = a.user_id
        WHERE a.provider_id = $1 AND u.email = $2
    `

	row := r.conn.Pool.QueryRow(ctx, query, auth.CredentialProviderID, email)

	var a auth.Account
	var u auth.User
	err := row.Scan(
		&a.ID,
		&a.Issuer,
		&a.AccountID,
		&a.ProviderID,
		&a.UserID,
		&a.AccessToken,
		&a.RefreshToken,
		&a.IDToken,
		&a.AccessTokenExpiresAt,
		&a.RefreshTokenExpiresAt,
		&a.Scope,
		&a.Password,
		&a.CreatedAt,
		&a.UpdatedAt,
		&u.ID,
		&u.Name,
		&u.Email,
		&u.EmailVerified,
		&u.Image,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.Role,
		&u.Banned,
		&u.BanReason,
		&u.BanExpires,
		&u.Username,
		&u.DisplayUsername,
		&u.TwoFactorEnabled,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, nil, &errs.Error{Code: errs.ENOTFOUND, Message: "Account not found"}
		default:
			return nil, nil, err
		}
	}
	return &a, &u, nil
}

func (r AccountRepository) UpdatePasswordByUserID(
	ctx context.Context,
	userID string,
	passwordHash string,
) error {
	query := `
        UPDATE "account"
        SET
            password = $1,
            updated_at = $2
        WHERE user_id = $3 AND provider_id = $4
    `
	_, err := r.conn.Pool.Exec(ctx, query,
		passwordHash,
		time.Now(),
		userID,
		auth.CredentialProviderID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r UserRepository) SetEmailVerified(ctx context.Context, id string, verified bool) error {
	query := `
        UPDATE "user"
        SET
            email_verified = $1,
            updated_at = $2
        WHERE id = $3
    `
	_, err := r.conn.Pool.Exec(ctx, query, verified, time.Now(), id)
	if err != nil {
		return err
	}
	return nil
}

func (r UserRepository) UpdateEmail(ctx context.Context, id string, email string) error {
	query := `
        UPDATE "user"
        SET
            email = $1,
            email_verified = $2,
            updated_at = $3
        WHERE id = $4
    `
	_, err := r.conn.Pool.Exec(ctx, query, email, false, time.Now(), id)
	if err != nil {
		return mapUniqueViolation(err, "email already in use")
	}
	return nil
}

func scanVerification(row pgx.Row) (*auth.Verification, error) {
	var v auth.Verification
	err := row.Scan(
		&v.ID,
		&v.Identifier,
		&v.Value,
		&v.ExpiresAt,
		&v.CreatedAt,
		&v.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r VerificationRepository) Create(ctx context.Context, verification *auth.Verification) error {
	query := `
        INSERT INTO "verification" (
            id,
            identifier,
            value,
            expires_at,
            created_at,
            updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := r.conn.Pool.Exec(ctx, query,
		verification.ID,
		verification.Identifier,
		verification.Value,
		verification.ExpiresAt,
		verification.CreatedAt,
		verification.UpdatedAt,
	)
	if err != nil {
		return mapUniqueViolation(err, "verification already in use")
	}
	return nil
}

func (r VerificationRepository) FindByValue(ctx context.Context, value string) (*auth.Verification, error) {
	query := `
        SELECT
            id,
            identifier,
            value,
            expires_at,
            created_at,
            updated_at
        FROM "verification"
        WHERE value = $1
    `
	row := r.conn.Pool.QueryRow(ctx, query, value)
	verification, err := scanVerification(row)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, &errs.Error{Code: errs.ENOTFOUND, Message: "Verification not found"}
		default:
			return nil, err
		}
	}
	return verification, nil
}

func (r VerificationRepository) DeleteByIdentifier(ctx context.Context, identifier string) error {
	_, err := r.conn.Pool.Exec(ctx, `DELETE FROM "verification" WHERE identifier = $1`, identifier)
	if err != nil {
		return err
	}
	return nil
}

func mapUniqueViolation(err error, message string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return &errs.Error{Code: errs.ECONFLICT, Message: message}
	}
	return err
}

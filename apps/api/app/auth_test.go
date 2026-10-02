package app

import (
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5"
	"github.com/samuelsih/golib/assert"
	"github.com/samuelsih/nibiru/api/app/auth"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionTTL   = time.Hour
	testPassword = "Password.1"
)

func login(t *testing.T, handler auth.Handler, email string) auth.Session {
	t.Helper()

	_, session, err := handler.Login(t.Context(), auth.LoginRequest{
		Email:    email,
		Password: testPassword,
	})
	assert.NoError(t, err)

	return session
}

func findUser(t *testing.T, email string) auth.User {
	t.Helper()

	rows, err := db.Query(t.Context(),
		`SELECT id, email, password, first_name, last_name, created_at FROM users WHERE email = $1`, email)
	assert.NoError(t, err)

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[auth.User])
	assert.NoError(t, err)

	return user
}

func userExists(t *testing.T, email string) bool {
	t.Helper()

	var exists bool
	err := db.QueryRow(t.Context(),
		`SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	assert.NoError(t, err)

	return exists
}

func countSessions(t *testing.T, token string) int {
	t.Helper()

	var count int
	err := db.QueryRow(t.Context(),
		`SELECT count(*) FROM sessions WHERE token = $1`, token).Scan(&count)
	assert.NoError(t, err)

	return count
}

func TestHandlerRegister(t *testing.T) {
	handler := auth.NewHandler(db, sessionTTL)

	tests := []struct {
		name      string
		firstName string
		lastName  null.String
		password  string
		wantErr   error
	}{
		{
			name:      "with last name",
			firstName: "Yanto",
			lastName:  null.StringFrom("Asep"),
			password:  testPassword,
		},
		{
			name:      "without last name",
			firstName: "Budi",
			password:  testPassword,
		},
		{
			name:      "long password",
			firstName: "Jajang",
			password:  strings.Repeat("a", 73),
			wantErr:   bcrypt.ErrPasswordTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email := uuid.New().String() + "@example.com"

			err := handler.Register(t.Context(), auth.RegisterRequest{
				Email:     email,
				Password:  tt.password,
				FirstName: tt.firstName,
				LastName:  tt.lastName,
			})

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.False(t, userExists(t, email))

				return
			}

			assert.NoError(t, err)

			user := findUser(t, email)
			assert.Equal(t, user.Email, email)
			assert.Equal(t, user.FirstName, tt.firstName)
			assert.Equal(t, user.LastName, tt.lastName)
			assert.NotEqual(t, user.PasswordHash, tt.password)
			assert.True(t, user.CorrectPassword(tt.password))
		})
	}
}

func TestHandlerRegisterDuplicateEmail(t *testing.T) {
	handler := auth.NewHandler(db, sessionTTL)
	email := uuid.New().String() + "@example.com"

	r := auth.RegisterRequest{
		Email:     email,
		Password:  testPassword,
		FirstName: "Kasep",
	}

	assert.NoError(t, handler.Register(t.Context(), r))

	err := handler.Register(t.Context(), r)
	assert.ErrorIs(t, err, auth.ErrEmailDuplicate)
}

func TestHandlerLogin(t *testing.T) {
	handler := auth.NewHandler(db, sessionTTL)
	email := uuid.New().String() + "@example.com"

	assert.NoError(t, handler.Register(t.Context(), auth.RegisterRequest{
		Email:     email,
		Password:  testPassword,
		FirstName: "Ceurik",
	}))

	unknownEmail := uuid.New().String() + "@example.com"

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  error
	}{
		{name: "valid credentials", email: email, password: testPassword},
		{name: "unknown email", email: unknownEmail, password: testPassword, wantErr: auth.ErrInvalidCredentials},
		{name: "wrong password", email: email, password: "not-the-password", wantErr: auth.ErrInvalidCredentials},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, session, err := handler.Login(t.Context(), auth.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			})

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, user, auth.User{})
				assert.Equal(t, session, auth.Session{})

				return
			}

			assert.NoError(t, err)

			stored := findUser(t, email)
			assert.Equal(t, user, stored)
			assert.Equal(t, session.UserID, stored.ID)
			assert.NotEqual(t, session.Token, "")
			assert.Equal(t, session.ExpiresAt.Sub(session.CreatedAt), sessionTTL)
			assert.False(t, session.Expired())
			assert.Equal(t, countSessions(t, session.Token), 1)
		})
	}
}

func TestHandlerLogout(t *testing.T) {
	handler := auth.NewHandler(db, sessionTTL)

	tests := []struct {
		name        string
		unknown     bool
		logoutFirst bool
		wantFirst   int
		wantSecond  int
	}{
		{name: "unknown token", unknown: true, wantFirst: 1, wantSecond: 1},
		{name: "first session", logoutFirst: true, wantFirst: 0, wantSecond: 1},
		{name: "second session", wantFirst: 1, wantSecond: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email := uuid.New().String() + "@example.com"

			assert.NoError(t, handler.Register(t.Context(), auth.RegisterRequest{
				Email:     email,
				Password:  testPassword,
				FirstName: "Herman",
			}))

			first := login(t, handler, email)
			second := login(t, handler, email)

			token := second.Token
			switch {
			case tt.unknown:
				token = "unknown-token"
			case tt.logoutFirst:
				token = first.Token
			}

			assert.NoError(t, handler.Logout(t.Context(), token))
			assert.Equal(t, countSessions(t, first.Token), tt.wantFirst)
			assert.Equal(t, countSessions(t, second.Token), tt.wantSecond)
		})
	}
}

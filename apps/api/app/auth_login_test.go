package app

import (
	"testing"

	"github.com/samuelsih/golib/assert"
)

func loginUser(t *testing.T, email string) Session {
	t.Helper()

	_, session, err := Login(t.Context(), db, sessionTTL, LoginRequest{
		Email:    email,
		Password: testPassword,
	})
	assert.NoError(t, err)

	return session
}

func TestLogin(t *testing.T) {
	email := newEmail()
	registerUser(t, email)

	unknownEmail := newEmail()

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  error
	}{
		{name: "valid credentials", email: email, password: testPassword},
		{name: "unknown email", email: unknownEmail, password: testPassword, wantErr: ErrInvalidCredentials},
		{name: "wrong password", email: email, password: "not-the-password", wantErr: ErrInvalidCredentials},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, session, err := Login(t.Context(), db, sessionTTL, LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			})

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, user, User{})
				assert.Equal(t, session, Session{})

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

func TestLoginRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     LoginRequest
		wantErr bool
	}{
		{
			name: "valid",
			req:  LoginRequest{Email: "admin@gmail.com", Password: "Password.1"},
		},
		{
			name:    "invalid email",
			req:     LoginRequest{Email: "not-an-email", Password: "Password.1"},
			wantErr: true,
		},
		{
			name:    "missing password",
			req:     LoginRequest{Email: "admin@gmail.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.wantErr {
				assert.NotNil(t, err)

				return
			}

			assert.NoError(t, err)
		})
	}
}

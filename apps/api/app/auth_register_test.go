package app

import (
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5"
	"github.com/samuelsih/golib/assert"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionTTL   = time.Hour
	testPassword = "Password.1"
)

func newEmail() string {
	return uuid.NewV7().String() + "@example.com"
}

func findUser(t *testing.T, email string) User {
	t.Helper()

	rows, err := db.Query(t.Context(),
		`SELECT id, email, password, first_name, last_name, created_at FROM users WHERE email = $1`, email)
	assert.NoError(t, err)

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
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

func registerUser(t *testing.T, email string) User {
	t.Helper()

	err := Register(t.Context(), db, RegisterRequest{
		Email:     email,
		Password:  testPassword,
		FirstName: "Sandi",
	})
	assert.NoError(t, err)

	return findUser(t, email)
}

func TestRegister(t *testing.T) {
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
			email := newEmail()

			err := Register(t.Context(), db, RegisterRequest{
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
			assert.True(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(tt.password)) == nil)
		})
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	email := newEmail()

	r := RegisterRequest{
		Email:     email,
		Password:  testPassword,
		FirstName: "Kasep",
	}

	assert.NoError(t, Register(t.Context(), db, r))

	err := Register(t.Context(), db, r)
	assert.ErrorIs(t, err, ErrEmailDuplicate)
}

func TestRegisterRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     RegisterRequest
		wantErr bool
	}{
		{
			name: "valid",
			req:  RegisterRequest{Email: "admin@gmail.com", Password: "Password.1", FirstName: "Admin"},
		},
		{
			name:    "invalid email",
			req:     RegisterRequest{Email: "not-an-email", Password: "Password.1", FirstName: "Admin"},
			wantErr: true,
		},
		{
			name:    "missing password",
			req:     RegisterRequest{Email: "admin@gmail.com", FirstName: "Admin"},
			wantErr: true,
		},
		{
			name:    "short password",
			req:     RegisterRequest{Email: "admin@gmail.com", Password: "short", FirstName: "Admin"},
			wantErr: true,
		},
		{
			name:    "missing first name",
			req:     RegisterRequest{Email: "admin@gmail.com", Password: "Password.1"},
			wantErr: true,
		},
		{
			name: "last name too long",
			req: RegisterRequest{
				Email:     "admin@gmail.com",
				Password:  "Password.1",
				FirstName: "Admin",
				LastName:  null.StringFrom(strings.Repeat("a", 101)),
			},
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

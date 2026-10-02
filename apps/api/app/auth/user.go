package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null/v6"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uuid.UUID   `db:"id"`
	Email        string      `db:"email"`
	PasswordHash string      `db:"password"`
	FirstName    string      `db:"first_name"`
	LastName     null.String `db:"last_name"`
	CreatedAt    time.Time   `db:"created_at"`
}

func NewUser(email, password, firstName string, lastName null.String) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		FirstName:    firstName,
		LastName:     lastName,
		CreatedAt:    time.Now(),
	}, nil
}

func (u User) CorrectPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

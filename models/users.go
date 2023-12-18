package models

import (
	"errors"
	"time"

	upperDB "github.com/upper/db/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	passwordCost    = 12
	usersEmailIndex = "users_email_key"
)

var (
	// ErrNoMoreRows ...
	ErrNoMoreRows = errors.New("No record found")
	// ErrDuplicateEmail ...
	ErrDuplicateEmail = errors.New("Email already exists")
	// ErrUserNotActive ...
	ErrUserNotActive = errors.New("User account is inactive")
	// ErrInvalidLogin ...
	ErrInvalidLogin = errors.New("Invalid login")
)

// UsersModel ...
type UsersModel struct {
	db upperDB.Session
}

// Users is the users table in postgres
type Users struct {
	ID        int       `db:"id,omitempty"`
	Username  string    `db:"username"`
	Password  string    `db:"password_hash"`
	Email     string    `db:"email"`
	Activated bool      `db:"activated"`
	CreatedAt time.Time `db:"created_at"`
}

// Table returns the table names
func (um UsersModel) Table() string {
	return "users"
}

// GetByID ...
func (um UsersModel) GetByID(id int) (*Users, error) {
	var user Users
	err := um.db.Collection(um.Table()).Find(upperDB.Cond{"id": id}).One(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail ...
func (um UsersModel) GetByEmail(email string) (*Users, error) {
	var user Users
	err := um.db.Collection(um.Table()).Find(upperDB.Cond{"email": email}).One(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Insert ...
func (um UsersModel) Insert(user *Users) error {
	newhash, err := bcrypt.GenerateFromPassword([]byte(user.Password), passwordCost)
	if err != nil {
		return err
	}
	user.Password = string(newhash)
	user.CreatedAt = time.Now()
	col := um.db.Collection(um.Table())
	res, err := col.Insert(user)
	if err != nil {
		switch {
		case errHasDuplicate(err, usersEmailIndex):
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	user.ID = convertUpperIDToInt(res.ID())

	return nil
}

func (u *Users) comparePassword(plainPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// AuthenticateUser logs in the user
func (um *UsersModel) AuthenticateUser(email, password string) (*Users, error) {
	user, err := um.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if !user.Activated {
		return nil, ErrUserNotActive
	}
	match, err := user.comparePassword(password)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, ErrInvalidLogin
	}
	return user, nil
}

package inmem

import (
	"strconv"

	"github.com/titan-x/titan/data"
	"github.com/titan-x/titan/models"
)

// DB is an in-memory database.
type DB struct {
	UserDB
}

// UserDB is in-memory user database.
type UserDB struct {
	ids    map[string]*models.User
	emails map[string]*models.User
}

// NewDB creates a new in-memory database.
func NewDB() *DB {
	return &DB{
		UserDB: UserDB{
			ids:    make(map[string]*models.User),
			emails: make(map[string]*models.User),
		},
	}
}

// Seed seeds database with essential data.
func (db UserDB) Seed(overwrite bool, jwtPass string) error {
	if err := data.SeedInit(jwtPass); err != nil {
		return err
	}

	for _, u := range data.SeedUsers {
		if err := db.SaveUser(&u); err != nil {
			return err
		}
	}

	return nil
}

// GetByID retrieves a user by ID.
func (db UserDB) GetByID(id string) (u *models.User, ok bool) {
	u, ok = db.ids[id]
	return
}

// GetByEmail retrieves a user by e-mail address.
func (db UserDB) GetByEmail(email string) (u *models.User, ok bool) {
	u, ok = db.emails[email]
	return
}

// SaveUser save or updates a user object in the database.
func (db UserDB) SaveUser(u *models.User) error {
	if u.ID == "" {
		u.ID = strconv.Itoa(len(db.ids) + 1)
	}

	db.ids[u.ID] = u
	db.emails[u.Email] = u
	return nil
}

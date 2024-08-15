package storage

import (
	"context"
	"errors"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	_ "github.com/lib/pq"
	"rest-authentication/config"
	"rest-authentication/model"
)

// Storage represents the storage service
type Storage struct {
	cfg config.DB
	db  *pg.DB
}

// New creates a new Storage instance
func New(ctx context.Context, cfg config.DB) (*Storage, error) {
	opt, err := pg.ParseURL(cfg.Addr)
	if err != nil {
		return nil, err
	}

	db := pg.Connect(opt)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err = db.Ping(ctx); err != nil {
		return nil, err
	}

	s := &Storage{db: db, cfg: cfg}

	if err = s.create(); err != nil {
		s.Close()
		return nil, err
	}

	return s, nil
}

// Close closes the database connection
func (s *Storage) Close() {
	_ = s.db.Close()
}

// create creates the tables
func (s *Storage) create() error {
	models := []interface{}{
		(*model.User)(nil),
		(*model.RefreshToken)(nil),
	}

	for _, m := range models {
		if err := s.db.Model(m).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		}); err != nil {
			return err
		}
	}

	return nil
}

// InsertRefreshToken inserts a refresh token
func (s *Storage) InsertRefreshToken(UUID, refresh string) error {
	_, err := s.db.Model(&model.RefreshToken{
		UUID:    UUID,
		Refresh: refresh,
	}).Insert()
	return err
}

// DeleteRefreshToken deletes a refresh token
func (s *Storage) DeleteRefreshToken(UUID string) error {
	_, err := s.db.Model(&model.RefreshToken{}).Where("uuid = ?", UUID).Delete()
	return err
}

// HasUUIDToken checks if a UUID token exists
func (s *Storage) HasUUIDToken(UUID string) (bool, error) {
	return s.db.Model(&model.RefreshToken{}).Where("uuid = ?", UUID).Exists()
}

// HasRefreshToken checks if a refresh token exists
func (s *Storage) HasRefreshToken(UUID, refresh string) (bool, error) {
	return s.db.Model(&model.RefreshToken{}).Where("uuid = ? AND refresh = ?", UUID, refresh).Exists()
}

// InsertUser inserts a new user into the database
func (s *Storage) InsertUser(user model.User) (string, error) {
	res, err := s.db.Model(&user).OnConflict("DO NOTHING").Insert()
	if err != nil || res.RowsAffected() == 0 {
		return "", err
	}
	return user.GUID, err
}

// SelectUserByGUIDAndPass selects a user by GUID and password
func (s *Storage) SelectUserByGUIDAndPass(GUID, password string) (model.User, error) {
	var user model.User
	err := s.db.Model(&user).
		Where("guid = ? and password = ?", GUID, password).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return model.User{}, nil
	}
	return user, err
}

// SelectUserEmailByGUID selects a user email by GUID
func (s *Storage) SelectUserEmailByGUID(GUID string) (string, error) {
	var user model.User
	err := s.db.Model(&user).
		Where("guid = ?", GUID).
		Select()
	return user.Email, err
}

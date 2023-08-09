package spass

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// FileStore implements Store
type FileStore struct {
	env *Env
}

// NewFileStore creates a new FileStore
func NewFileStore(env *Env) *FileStore {
	return &FileStore{
		env: env,
	}
}

// List the secrets in the store
func (s *FileStore) List(ctx context.Context, namespace string) ([]*SecretFile, error) {
	res := []*SecretFile{}

	err := filepath.Walk(s.env.PASSWORD_STORE_DIR, func(pth string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == ".gpg-id" {
			return nil
		}

		res = append(res, &SecretFile{
			env:      s.env,
			filename: pth,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Get a secret by name.
func (s *FileStore) Secret(ctx context.Context, name string) (*SecretFile, error) {
	filename := filepath.Join(s.env.PASSWORD_STORE_DIR, name) + ".gpg"

	notfound := fmt.Errorf("no secret found named '%s'", name)

	f, err := os.Open(filename)
	if err != nil {
		return nil, notfound
	}

	info, err := f.Stat()
	if err != nil {
		return nil, notfound
	}

	if info.IsDir() {
		return nil, notfound
	}

	secret := &SecretFile{
		env:      s.env,
		filename: filename,
	}

	return secret, nil
}

// NewSecret returns a new secret.
func (s *FileStore) NewSecret(ctx context.Context, name string) (*SecretFile, error) {
	filename := filepath.Join(s.env.PASSWORD_STORE_DIR, name) + ".gpg"

	// TODO: check if file exists?

	secret := &SecretFile{
		env:      s.env,
		filename: filename,
	}

	return secret, nil
}

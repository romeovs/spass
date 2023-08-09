package spass

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SecretFile implements Secret
type SecretFile struct {
	env      *Env
	filename string
}

// strip the gpg suffix of a secret
func strip(filename string) string {
	return strings.TrimSuffix(filename, ".gpg")
}

// Name gets the name of secret
func (s *SecretFile) FullName() string {
	return strings.TrimPrefix(strip(s.filename), s.env.PASSWORD_STORE_DIR+"/")
}

// Name gets the name of secret
func (s *SecretFile) Name() string {
	return filepath.Base(s.FullName())
}

// Name gets the name of secret
func (s *SecretFile) Namespace() string {
	dir := filepath.Dir(s.FullName())
	if dir == "." {
		return ""
	}
	return dir
}

func (s *SecretFile) decrypt(ctx context.Context) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gpg", "--decrypt", s.filename)
	return cmd.Output()
}

func (s *SecretFile) encrypt(ctx context.Context, keyid string, content string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gpg", "--recipient", keyid, "--encrypt")
	cmd.Stdin = strings.NewReader(content)
	return cmd.Output()
}

func (s *SecretFile) keyid() (string, error) {
	dir := filepath.Dir(s.filename)
	idFile := filepath.Join(dir, ".gpg-id")
	buf, err := os.ReadFile(idFile)
	if err != nil {
		return "", fmt.Errorf("cannot read .gpg-id file for secret")
	}

	return strings.TrimRight(string(buf), "\n"), nil
}

func (s *SecretFile) Write(ctx context.Context, content string) error {
	keyid, err := s.keyid()
	if err != nil {
		return err
	}

	buf, err := s.encrypt(ctx, keyid, content)
	if err != nil {
		return err
	}

	err = os.WriteFile(s.filename, buf, 0644)
	if err != nil {
		return fmt.Errorf("could not write secret '%s'", s.FullName())
	}

	return nil
}

// Name gets the name of secret
func (s *SecretFile) Body(ctx context.Context) (string, error) {
	buf, err := s.decrypt(ctx)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

// Name gets the name of secret
func (s *SecretFile) Password(ctx context.Context) (string, error) {
	body, err := s.Body(ctx)
	if err != nil {
		return "", nil
	}

	lines := strings.SplitN(body, "\n", 2)
	pass := lines[0]

	if pass == "" {
		return "", fmt.Errorf("no password set for secret '%s'", s.FullName())
	}

	if strings.HasPrefix(pass, "otpauth://totp/") {
		return "", fmt.Errorf("no password set for secret '%s'", s.FullName())
	}

	return pass, nil
}

// Set password
func (s *SecretFile) SetPassword(ctx context.Context, password string) error {
	body := ""
	if _, err := os.Stat(s.filename); err == nil {
		body, err = s.Body(ctx)
		if err != nil {
			return err
		}
	}

	lines := strings.SplitN(body, "\n", 2)

	lines[0] = password
	body = strings.Join(lines, "\n")

	return s.Write(ctx, body)
}

// Remove wipes and removes the secret
func (s *SecretFile) Remove() error {
	f, err := os.OpenFile(s.filename, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("could not open secret '%s'", s.FullName())
	}

	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("could not stat secret '%s'", s.FullName())
	}

	size := info.Size()
	zeroes := make([]byte, size)
	copy(zeroes[:], "0")
	n, err := f.Write(zeroes)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("could not wipe secret '%s'", s.FullName())
	}

	if int64(n) != size {
		fmt.Println(err)
		return fmt.Errorf("could not fully wipe secret '%s'", s.FullName())
	}

	err = os.Remove(s.filename)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("could not remove secret '%s'", s.FullName())
	}

	return nil
}

// Pairs gets the pairs in the secret file
func (s *SecretFile) Pairs(ctx context.Context) ([]*Pair, error) {
	body, err := s.Body(ctx)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(body, "\n")
	res := make([]*Pair, 0, len(lines))
	for i, line := range lines {
		if i == 0 {
			// skip password
			continue
		}
		if line == "" {
			continue
		}
		res = append(res, parse(line))
	}

	return res, nil
}

func parse(line string) *Pair {
	parts := strings.SplitN(line, ": ", 2)
	if len(parts) <= 1 {
		return &Pair{
			Value: line,
		}
	}

	key := parts[0]
	value := parts[1]

	if strings.HasPrefix(value, "//") {
		key = ""
		value = line
	}

	return &Pair{
		Key:   key,
		Value: value,
	}
}

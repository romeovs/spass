package spass

import (
	"fmt"
	"os"
	"path/filepath"
)

type Env struct {
	PASSWORD_STORE_DIR string
	EDITOR string
}

func ReadEnv() *Env {
	dir := filepath.Join(os.Getenv("HOME"), ".password-store")
	if env := os.Getenv("PASSWORD_STORE_DIR"); env != "" {
		dir = env
	}

	editor := "vim"
	if env := os.Getenv("EDITOR"); env != "" {
		editor = env
	}

	return &Env {
		PASSWORD_STORE_DIR: dir,
		EDITOR: editor,
	}
}

func (env *Env) Print() {
	fmt.Printf("PASSWORD_STORE_DIR=%s\n", env.PASSWORD_STORE_DIR)
	fmt.Printf("EDITOR=%s\n", env.EDITOR)
}

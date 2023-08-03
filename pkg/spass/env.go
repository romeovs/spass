package spass

import (
	"fmt"
	"os"
	"path/filepath"
)

type Env struct {
	PASSWORD_STORE_DIR string
	EDITOR string
	HAVEIBEENPWND_API_KEY string
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

	pnwd := ""
	if env := os.Getenv("HAVEIBEENPWND_API_KEY"); env != "" {
		editor = env
	}

	return &Env {
		PASSWORD_STORE_DIR: dir,
		EDITOR: editor,
		HAVEIBEENPWND_API_KEY: pwnd,
	}
}

func (env *Env) Print() {
	fmt.Printf("PASSWORD_STORE_DIR=%s\n", env.PASSWORD_STORE_DIR)
	fmt.Printf("EDITOR=%s\n", env.EDITOR)
	fmt.Printf("HAVEIBEENPWND_API_KEY=%s\n", env.HAVEIBEENPWND_API_KEY)
}

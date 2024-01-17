package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/romeovs/spass/pkg/clipboard"
	"github.com/romeovs/spass/pkg/editor"
	"github.com/romeovs/spass/pkg/generate"
	"github.com/romeovs/spass/pkg/pwnd"
	"github.com/romeovs/spass/pkg/spass"
	"github.com/urfave/cli/v2"
)

func main() {
	env := spass.ReadEnv()
	store := spass.NewFileStore(env)
	ctx := context.Background()

	app := &cli.App{
		Name:                   "spass",
		Usage:                  "a fun password manager, compatible with pass.",
		Suggest:                true,
		UseShortOptionHandling: true,
		ExitErrHandler: func(cli *cli.Context, err error) {
			if err == nil {
				os.Exit(0)
				return
			}

			fmt.Println(err)
			os.Exit(1)
		},
		Commands: []*cli.Command{
			{
				Name:  "env",
				Usage: "print the relevant environment variables or defaults",
				Action: func(cli *cli.Context) error {
					env.Print()
					return nil
				},
			},
			{
				Name:      "list",
				Aliases:   []string{"ls"},
				ArgsUsage: "[namespace]",
				Usage:     "list the secrets in the password store",
				Action: func(cli *cli.Context) error {
					namespace := cli.Args().Get(0)

					secrets, err := store.List(ctx, namespace)
					if err != nil {
						return err
					}

					for _, secret := range secrets {
						fmt.Println(secret.FullName())
					}

					return nil
				},
			},
			{
				Name:      "pass",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "copy",
						Aliases: []string{"c"},
						Usage:   "copy value to the clipboard",
						Value:   false,
					},
				},
				Usage: "show the password for the specified secret",
				Action: func(cli *cli.Context) error {
					name := cli.Args().Get(0)
					if name == "" {
						return errors.New("no name provided")
					}

					secret, err := store.Secret(ctx, name)
					if err != nil {
						return err
					}

					if secret == nil {
						return errors.New("no secret found")
					}

					pass, err := secret.Password(ctx)
					if err != nil {
						return err
					}

					if cli.Bool("copy") {
						fmt.Println("password copied!")
						clipboard.Write(pass)
					} else {
						fmt.Println(pass)
					}

					// Show otp if found
					body, err := secret.Body(ctx)
					if err != nil {
						return err
					}

					lines := strings.Split(body, "\n")
					otpauth := ""
					for _, line := range lines {
						if strings.HasPrefix(line, "otpauth://totp") {
							otpauth = line
						}
					}

					if otpauth == "" {
						return nil
					}

					key, err := otp.NewKeyFromURL(otpauth)
					if err != nil {
						return fmt.Errorf("invalid otp set up in secret '%s'", secret.FullName())
					}

					now := time.Now()
					epoch := now.Unix()
					period := int64(key.Period())

					elapsed := epoch % period
					left := period - elapsed

					if cli.Bool("wait") && left < 3 {
						fmt.Println("waiting for new token...")
						time.Sleep(time.Duration(left+1) * time.Second)
					}

					code, err := totp.GenerateCode(key.Secret(), time.Now())
					if err != nil {
						return fmt.Errorf("failed to generate code: %v", err)
					}

					now = time.Now()
					epoch = now.Unix()
					period = int64(key.Period())

					elapsed = epoch % period
					left = period - elapsed

					fmt.Printf("%-10s valid for another %ds\n", code, left)

					return nil
				},
			},
			{
				Name:      "show",
				ArgsUsage: "[name]",
				Usage:     "show all the info for the specified secret",
				Flags:     []cli.Flag{},
				Action: func(cli *cli.Context) error {
					name := cli.Args().Get(0)
					if name == "" {
						return errors.New("no name provided")
					}

					secret, err := store.Secret(ctx, name)
					if err != nil {
						return err
					}

					if secret == nil {
						return errors.New("unreachable")
					}

					body, err := secret.Body(ctx)
					if err != nil {
						return err
					}

					fmt.Printf("%s", body)
					return nil
				},
			},
			{
				Name:      "generate",
				ArgsUsage: "[name]",
				Usage:     "generate a new password and store as a secret under the provided name",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "lowercase",
						Aliases: []string{"l"},
						Value:   false,
						Usage:   "use only lowercase characters",
					},
					&cli.BoolFlag{
						Name:    "no-numbers",
						Aliases: []string{"n"},
						Value:   false,
						Usage:   "do not use numbers",
					},
					&cli.BoolFlag{
						Name:    "no-symbols",
						Aliases: []string{"s"},
						Value:   false,
						Usage:   "do not use symbols",
					},
					&cli.BoolFlag{
						Name:    "overwrite",
						Aliases: []string{"o"},
						Value:   false,
						Usage:   "overwrite existing password if the secret already exists",
					},
				},
				Action: func(cli *cli.Context) error {
					name := cli.Args().Get(0)
					if name == "" {
						return errors.New("no name provided")
					}

					secret, _ := store.Secret(ctx, name)
					if secret != nil && !cli.Bool("overwrite") {
						return fmt.Errorf("a secret with that name already exists, pass --overwrite to overwrite the password")
					}

					generator := &generate.Generator{
						LowerCase: cli.Bool("lowercase"),
						NoDigits:  cli.Bool("no-numbers"),
						NoSymbols: cli.Bool("no-symbols"),
					}

					size := 18
					password, err := generator.Generate(size)
					if err != nil {
						return err
					}

					secret, err = store.NewSecret(ctx, name)
					if err != nil {
						return err
					}

					err = secret.SetPassword(ctx, password)
					if err != nil {
						return err
					}

					fmt.Println(password)
					return nil
				},
			},
			{
				Name:      "edit",
				ArgsUsage: "[name]",
				Usage:     "edit the contents of the specified secret",
				Flags:     []cli.Flag{},
				Action: func(cli *cli.Context) error {
					name := cli.Args().Get(0)
					if name == "" {
						return errors.New("no name provided")
					}

					secret, err := store.Secret(ctx, name)
					if err != nil {
						return err
					}

					if secret == nil {
						return errors.New("unreachable")
					}

					body, err := secret.Body(ctx)
					if err != nil {
						return err
					}

					b, err := editor.Edit(env.EDITOR, body)
					if err != nil {
						return err
					}

					err = secret.Write(ctx, string(b))
					if err != nil {
						return err
					}

					fmt.Printf("secret '%s' saved!\n", secret.FullName())

					return nil
				},
			},
			{
				Name:      "remove",
				Aliases:   []string{"rm"},
				ArgsUsage: "[name]",
				Usage:     "delete a secret in the store",
				Flags:     []cli.Flag{},
				Action: func(cli *cli.Context) error {
					name := cli.Args().Get(0)
					if name == "" {
						return errors.New("no name provided")
					}

					secret, err := store.Secret(ctx, name)
					if err != nil {
						return err
					}

					if secret == nil {
						return errors.New("unreachable")
					}

					return secret.Remove()
				},
			},
			{
				Name:      "get",
				ArgsUsage: "[name] [key]",
				Usage:     "get the value of the key in the specified secret",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "case-insensitive",
						Aliases: []string{"i"},
						Value:   false,
						Usage:   "make the key match case insensitive",
					},
				},
				Action: func(cli *cli.Context) error {
					name := cli.Args().Get(0)
					if name == "" {
						return errors.New("no name provided")
					}

					caseInsensitive := cli.Bool("case-insensitive")

					key := cli.Args().Get(1)
					if name == "" {
						return errors.New("no key provided")
					}

					secret, err := store.Secret(ctx, name)
					if err != nil {
						return err
					}

					if secret == nil {
						return errors.New("unreachable")
					}

					pairs, err := secret.Pairs(ctx)
					if err != nil {
						return err
					}

					fmt.Println(caseInsensitive)

					ok := false
					for _, pair := range pairs {
						if caseInsensitive {
							if strings.EqualFold(pair.Key, key) {
								ok = true
								fmt.Println(pair.Value)
							}
						} else {
							if pair.Key == key {
								ok = true
								fmt.Println(pair.Value)
							}
						}
					}

					if !ok {
						return fmt.Errorf("key '%s' not found in secret '%s'", key, name)
					}

					return nil
				},
			},
			{
				Name:      "otp",
				ArgsUsage: "[name]",
				Usage:     "get an one time password from the specified secret",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "copy",
						Aliases: []string{"c"},
						Usage:   "copy value to the clipboard",
						Value:   false,
					},
					&cli.BoolFlag{
						Name:    "wait",
						Aliases: []string{"w"},
						Usage:   "wait for a new token if the current one is about to expire",
						Value:   false,
					},
				},
				Action: func(cli *cli.Context) error {
					name := cli.Args().Get(0)
					if name == "" {
						return errors.New("no name provided")
					}

					secret, err := store.Secret(ctx, name)
					if err != nil {
						return err
					}

					if secret == nil {
						return errors.New("unreachable")
					}

					body, err := secret.Body(ctx)
					if err != nil {
						return err
					}

					lines := strings.Split(body, "\n")
					otpauth := ""
					for _, line := range lines {
						if strings.HasPrefix(line, "otpauth://totp") {
							otpauth = line
						}
					}

					if otpauth == "" {
						return fmt.Errorf("no otp set up in secret '%s'", secret.FullName())
					}

					key, err := otp.NewKeyFromURL(otpauth)
					if err != nil {
						return fmt.Errorf("invalid otp set up in secret '%s'", secret.FullName())
					}

					now := time.Now()
					epoch := now.Unix()
					period := int64(key.Period())

					elapsed := epoch % period
					left := period - elapsed

					if cli.Bool("wait") && left < 3 {
						fmt.Println("waiting for new token...")
						time.Sleep(time.Duration(left+1) * time.Second)
					}

					code, err := totp.GenerateCode(key.Secret(), time.Now())
					if err != nil {
						return fmt.Errorf("failed to generate code: %v", err)
					}

					now = time.Now()
					epoch = now.Unix()
					period = int64(key.Period())

					elapsed = epoch % period
					left = period - elapsed

					fmt.Printf("%-10s valid for another %ds\n", code, left)

					if cli.Bool("copy") {
						clipboard.Write(code)
						fmt.Printf("%10s copied to clipboard!\n", " ")
					}

					return nil
				},
			},
			{
				Name:      "pwnd",
				ArgsUsage: "[name]",
				Usage:     "check if the password in the specified secret was pwnd",
				Flags:     []cli.Flag{},
				Action: func(cli *cli.Context) error {
					name := cli.Args().Get(0)
					if name == "" {
						return errors.New("no name provided")
					}

					secret, err := store.Secret(ctx, name)
					if err != nil {
						return err
					}

					if secret == nil {
						return errors.New("unreachable")
					}

					password, err := secret.Password(ctx)
					if err != nil {
						return err
					}

					client := pwnd.NewClient(env.HAVEIBEENPWND_API_KEY)
					ispwnd, err := client.Check(password)
					if err != nil {
						return err
					}

					if ispwnd {
						fmt.Println("this password has been pwnd, generate a new one")
					} else {
						fmt.Println("this password has not been pwnd!")
					}

					return nil
				},
			},
			{
				Name:      "search",
				ArgsUsage: "[query]",
				Usage:     "search for a secret containg the query",
				Flags:     []cli.Flag{},
				Action: func(cli *cli.Context) error {
					arg := cli.Args().Get(0)

					parts := strings.Split(arg, ":")
					key := parts[0]
					val := parts[1]

					secrets, err := store.List(ctx, "")
					if err != nil {
						return err
					}

					ok := false
					for _, secret := range secrets {
						pairs, err := secret.Pairs(ctx)
						if err != nil {
							return err
						}

						for _, pair := range pairs {
							if strings.ToLower(key) == strings.ToLower(pair.Key) {
								if strings.Contains(strings.ToLower(pair.Value), strings.ToLower(val)) {
									ok = true
									fmt.Printf("match found in secret '%s':\n", secret.FullName())
									fmt.Printf("%s: %s\n", pair.Key, pair.Value)
								}
							}
						}
					}
					if ok {
						return nil
					} else {
						return fmt.Errorf("no match found")
					}
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

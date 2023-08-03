package main

import (
	"context"
	"os"
	"log"
	"fmt"
	"errors"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/romeovs/spass/pkg/spass"
	"github.com/romeovs/spass/pkg/editor"
	"github.com/romeovs/spass/pkg/generate"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func main() {
	// search [query]
	// copy [name]
	// copy [name] [key]
	// pwnd [namespace]
	// completions

	env := spass.ReadEnv()
	store := spass.NewFileStore(env)
	ctx := context.Background()

	app := &cli.App{
		Name: "spass",
		Usage: "a fun password manager, compatible with pass.",
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
				Name: "env",
				Usage: "print the relevant environment variables or defaults",
				Action: func(cli *cli.Context) error {
					env.Print()
					return nil
				},
			},
			{
				Name: "list",
				Aliases: []string{"ls"},
				ArgsUsage: "[namespace]",
				Usage: "list the secrets in the password store",
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
				Name: "pass",
				ArgsUsage: "[name]",
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

					pass, err := secret.Password()
					if err != nil {
						return err
					}

					fmt.Println(pass)
					return nil
				},
			},
			{
				Name: "show",
				ArgsUsage: "[name]",
				Usage: "show all the info for the specified secret",
				Flags: []cli.Flag{},
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

					body, err := secret.Body()
					if err != nil {
						return err
					}

					fmt.Printf("%s", body)
					return nil
				},
			},
			{
				Name: "generate",
				ArgsUsage: "[name]",
				Usage: "generate a new password and store as a secret under the provided name",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "lowercase",
						Aliases: []string{"l"},
						Value: false,
						Usage: "use only lowercase characters",
					},
					&cli.BoolFlag{
						Name: "no-numbers",
						Aliases: []string{"n"},
						Value: false,
						Usage: "do not use numbers",
					},
					&cli.BoolFlag{
						Name: "no-symbols",
						Aliases: []string{"s"},
						Value: false,
						Usage: "do not use symbols",
					},
					&cli.BoolFlag{
						Name: "overwrite",
						Aliases: []string{"o"},
						Value: false,
						Usage: "overwrite existing password if the secret already exists",
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
						NoDigits: cli.Bool("no-numbers"),
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


					err = secret.SetPassword(password)
					if err != nil {
						return err
					}

					fmt.Println(password)
					return nil
				},
			},
			{
				Name: "edit",
				ArgsUsage: "[name]",
				Usage: "edit the contents of the specified secret",
				Flags: []cli.Flag{},
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

					body, err := secret.Body()
					if err != nil {
						return err
					}

					b, err := editor.Edit(env.EDITOR, body)
					if err != nil {
						return err
					}

					err = secret.Write(string(b))
					if err != nil {
						return err
					}

					fmt.Printf("secret '%s' saved!\n", secret.FullName())

					return nil
				},
			},
			{
				Name: "remove",
				Aliases: []string{"rm"},
				ArgsUsage: "[name]",
				Usage: "delete a secret in the store",
				Flags: []cli.Flag{},
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
				Name: "get",
				ArgsUsage: "[name] [key]",
				Usage: "get the value of the key in the specified secret",
				Flags: []cli.Flag{
					&cli.BoolFlag {
						Name: "case-insensitive",
						Aliases: []string {
							"i",
						},
						Value: false,
						Usage: "make the key match case insensitive",
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

					pairs, err := secret.Pairs()
					if err != nil {
						return err
					}

					fmt.Println(caseInsensitive)

					ok := false
					for _, pair := range pairs {
						if caseInsensitive {
							if strings.ToLower(pair.Key) == strings.ToLower(key) {
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
				Name: "otp",
				ArgsUsage: "[name]",
				Usage: "get an one time password from the specified secret",
				Flags: []cli.Flag{},
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

					body, err := secret.Body()
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

					if left < 3 {
						time.Sleep(time.Duration(left + 1) * time.Second)
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

					fmt.Printf("%s     valid for another %ds\n", code, left)

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

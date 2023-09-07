# spaß

A [`pass`](https://www.passwordstore.org/)-compatible password manager for the cli.

## Usage

```
NAME:
   spass - a fun password manager, compatible with pass.

USAGE:
   spass [global options] command [command options] [arguments...]

COMMANDS:
   env         print the relevant environment variables or defaults
   list, ls    list the secrets in the password store
   pass        show the password for the specified secret
   show        show all the info for the specified secret
   generate    generate a new password and store as a secret under the provided name
   edit        edit the contents of the specified secret
   remove, rm  delete a secret in the store
   get         get the value of the key in the specified secret
   otp         get an one time password from the specified secret
   pwnd        check if the password in the specified secret was pwnd
   search      get an one time password from the specified secret
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

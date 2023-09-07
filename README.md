# spaß

A [`pass`](https://www.passwordstore.org/)-compatible password manager for the cli.

`spass` is fully backwards compatible with [GNU `pass`](https://www.passwordstore.org/),
but adds a couple of extra features which are nice.

Where `pass`' secret format is almost completely free-form ,
`spass` adds the concept of "named fields" which are a bit like HTTP-headers
and can help you organise the data in your secrets.

A secret has the following structure:
```
<password>
<field name>: <field value>
<unstructured data>
```
For example:
```
pa$$w0rd
domain: example.com
This is some unstructured data that does not have a field name.
issuer: google.com
username: john-doe
```

`spass` also adds support for generating One-Time Passwords (OTPs).
When one of the fields in the secret is a valid
[`otpauth://` uri](https://github.com/google/google-authenticator/wiki/Key-Uri-Format)
`spass` will be able to generate an OTP for it.

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
   search      search for a secret containg the query
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

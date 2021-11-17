module github.com/synfinatic/aws-sso-cli

go 1.16

// pin this version (or later) until 99designs/keyring updates.
replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4

require (
	github.com/99designs/keyring v1.1.6
	github.com/alecthomas/kong v0.2.18
	github.com/atotto/clipboard v0.1.4
	github.com/aws/aws-sdk-go v1.38.40
	github.com/c-bata/go-prompt v0.2.5 // 0.2.6 is broken
	github.com/davecgh/go-spew v1.1.1
	github.com/goccy/go-yaml v1.9.4
	github.com/knadh/koanf v0.16.0
	github.com/manifoldco/promptui v0.9.0
	github.com/posener/complete v1.2.3
	github.com/sirupsen/logrus v1.7.0
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966
	github.com/stretchr/testify v1.7.0
	github.com/synfinatic/gotable v0.0.1
	github.com/willabides/kongplete v0.2.0
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a

	// see: https://github.com/sirupsen/logrus/issues/1275
	golang.org/x/sys v0.0.0-20210817190340-bfb29a6856f2 // indirect
)

module github.com/redhajuanda/fayl

go 1.24.2

require (
	github.com/VauntDev/tqla v0.0.3
	github.com/georgysavva/scany/v2 v2.1.4
	github.com/jmoiron/sqlx v1.4.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pkg/errors v0.9.1
	github.com/redhajuanda/kuysor v0.0.0-20250708072051-8185612563b8
	github.com/redhajuanda/perkakas v0.0.0-20250721042902-fcb1bf9edba4
	github.com/stretchr/testify v1.10.0
	go.uber.org/mock v0.5.2
)

// replace github.com/redhajuanda/kuysor => ../../redhajuanda/kuysor
// replace github.com/redhajuanda/perkakas => ../silib

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

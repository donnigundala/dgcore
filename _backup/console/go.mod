module github.com/donnigundala/dgcore/console

go 1.25.0

require (
	github.com/donnigundala/dgcore/config v0.0.0
	github.com/donnigundala/dgcore/contracts v0.0.0
	github.com/donnigundala/dgcore/database v0.0.0-00010101000000-000000000000
	github.com/donnigundala/dgcore/foundation v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.8.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/donnigundala/dgcore/container v0.0.0 // indirect
	github.com/donnigundala/dgcore/ctxutil v0.0.0 // indirect
	github.com/donnigundala/dgcore/http v0.0.0-00010101000000-000000000000 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.mongodb.org/mongo-driver v1.11.4 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	gorm.io/driver/mysql v1.6.0 // indirect
	gorm.io/driver/postgres v1.6.0 // indirect
	gorm.io/driver/sqlite v1.6.0 // indirect
	gorm.io/gorm v1.31.0 // indirect
	gorm.io/plugin/dbresolver v1.6.2 // indirect
)

replace github.com/donnigundala/dgcore/contracts => ../contracts

replace github.com/donnigundala/dgcore/foundation => ../foundation

replace github.com/donnigundala/dgcore/config => ../config

replace github.com/donnigundala/dgcore/database => ../database

replace github.com/donnigundala/dgcore/ctxutil => ../ctxutil

replace github.com/donnigundala/dgcore/http => ../http

replace github.com/donnigundala/dgcore/container => ../container

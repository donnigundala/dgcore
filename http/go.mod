module github.com/donnigundala/dgcore/http

go 1.25.0

require (
	github.com/donnigundala/dgcore/contracts v0.0.0
	github.com/donnigundala/dgcore/ctxutil v0.0.0
	github.com/donnigundala/dgcore/errors v0.0.0
	github.com/donnigundala/dgcore/logging v0.0.0
	github.com/gin-gonic/gin v1.10.0
	github.com/google/uuid v1.6.0
	golang.org/x/time v0.5.0
)

replace github.com/donnigundala/dgcore/ctxutil => ../ctxutil

replace github.com/donnigundala/dgcore/contracts => ../contracts

replace github.com/donnigundala/dgcore/errors => ../errors

replace github.com/donnigundala/dgcore/logging => ../logging

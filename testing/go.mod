module github.com/donnigundala/dgcore/testing

go 1.25.0

require (
	github.com/donnigundala/dgcore/contracts v0.0.0
	github.com/donnigundala/dgcore/logging v0.0.0
)

replace github.com/donnigundala/dgcore/contracts => ../contracts

replace github.com/donnigundala/dgcore/logging => ../logging

// Package tools pins the controller-gen version this repository builds.
// tools.go carries the import behind the tools build tag; this untagged file
// gives `go test ./...` and `go vet ./...` a package to find, which the
// nested Go modules workflow needs.
package tools

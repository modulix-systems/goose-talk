// Package repo implements application outer layer logic. Each logic group in own file.
package gateways

//go:generate mockgen -source=contracts.go -destination=../services/mocks_repo_test.go -package=services_test
type (
	UsersRepo interface {
	}
)

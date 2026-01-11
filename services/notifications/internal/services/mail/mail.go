package mail

import "github.com/modulix-systems/goose-talk/logger"

type Service struct {
	log logger.Interface
}

func New() *Service {
	return &Service{}
}

package client

import (
	"context"

	"github.com/henrique-avice/goexpert-desafio-multithreading/internal/models"
)

type CEPClient interface {
	Query(ctx context.Context, cep string) (*models.Address, error)
}

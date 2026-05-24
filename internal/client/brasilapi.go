package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/henrique-avice/goexpert-desafio-multithreading/internal/models"
)

const maxResponseSize = 10 * 1024

const BrasilAPIURL = "https://brasilapi.com.br/api/cep/v1"

type BrasilAPI struct {
	client  *http.Client
	baseURL string
	logger  *slog.Logger
}

type brasilAPIResponse struct {
	CEP          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
}

func NewBrasilAPI(client *http.Client) *BrasilAPI {
	if client == nil {
		client = &http.Client{
			Timeout: 2 * time.Second,
		}
	}
	return &BrasilAPI{
		client:  client,
		baseURL: BrasilAPIURL,
		logger:  logger,
	}
}

func (b *BrasilAPI) Query(ctx context.Context, cep string) (*models.Address, error) {
	if len(cep) != 8 {
		return nil, fmt.Errorf("CEP inválido: esperado 8 dígitos, recebido %d", len(cep))
	}

	url := fmt.Sprintf("%s/%s", b.baseURL, cep)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		b.logger.Error("erro ao criar request BrasilAPI",
			slog.String("cep", cep),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("erro ao criar request: %w", err)
	}

	req.Header.Set("User-Agent", "FullCycleGoLang/1.0")

	resp, err := b.client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			b.logger.Warn("timeout em BrasilAPI",
				slog.String("cep", cep),
			)
			return nil, fmt.Errorf("timeout: BrasilAPI não respondeu")
		}
		b.logger.Error("erro na requisição BrasilAPI",
			slog.String("cep", cep),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("erro na requisição: %w", err)
	}
	defer resp.Body.Close()

	limitedBody := io.LimitReader(resp.Body, maxResponseSize)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(limitedBody)
		b.logger.Warn("status code não OK em BrasilAPI",
			slog.String("cep", cep),
			slog.Int("status", resp.StatusCode),
		)
		return nil, fmt.Errorf(
			"BrasilAPI retornou %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var apiResp brasilAPIResponse
	if err := json.NewDecoder(limitedBody).Decode(&apiResp); err != nil {
		b.logger.Error("erro ao parsear resposta BrasilAPI",
			slog.String("cep", cep),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("erro ao parsear resposta: %w", err)
	}

	if apiResp.CEP == "" || apiResp.State == "" || apiResp.City == "" || apiResp.Street == "" {
		b.logger.Warn("resposta incompleta da BrasilAPI",
			slog.String("cep", cep),
			slog.String("api_cep", apiResp.CEP),
			slog.String("state", apiResp.State),
			slog.String("city", apiResp.City),
			slog.String("street", apiResp.Street),
		)
		return nil, fmt.Errorf("resposta incompleta da BrasilAPI")
	}

	b.logger.Info("sucesso em BrasilAPI",
		slog.String("cep", cep),
		slog.String("city", apiResp.City),
	)

	return &models.Address{
		CEP:        apiResp.CEP,
		Logradouro: apiResp.Street,
		Bairro:     apiResp.Neighborhood,
		Cidade:     apiResp.City,
		Estado:     apiResp.State,
	}, nil
}

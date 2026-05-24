package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/henrique-avice/goexpert-desafio-multithreading/internal/models"
)

const ViaCEPURL = "https://viacep.com.br/ws"

type ViaCEP struct {
	client  *http.Client
	baseURL string
	logger  *slog.Logger
}

type viaCEPResponse struct {
	CEP        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
}

func NewViaCEP(client *http.Client) *ViaCEP {
	if client == nil {
		client = &http.Client{
			Timeout: 2 * time.Second,
		}
	}
	return &ViaCEP{
		client:  client,
		baseURL: ViaCEPURL,
		logger:  logger,
	}
}

func (v *ViaCEP) Query(ctx context.Context, cep string) (*models.Address, error) {
	if len(cep) != 8 {
		return nil, fmt.Errorf("CEP inválido: %s", cep)
	}

	formattedCEP := fmt.Sprintf("%s-%s", cep[:5], cep[5:])
	url := fmt.Sprintf("%s/%s/json/", v.baseURL, formattedCEP)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		v.logger.Error("erro ao criar request ViaCEP",
			slog.String("cep", cep),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("erro ao criar request: %w", err)
	}

	req.Header.Set("User-Agent", "FullCycleGoLang/1.0")

	resp, err := v.client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			v.logger.Warn("timeout em ViaCEP",
				slog.String("cep", cep),
			)
			return nil, fmt.Errorf("timeout: ViaCEP não respondeu")
		}
		v.logger.Error("erro na requisição ViaCEP",
			slog.String("cep", cep),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("erro na requisição: %w", err)
	}
	defer resp.Body.Close()

	limitedBody := io.LimitReader(resp.Body, maxResponseSize)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(limitedBody)
		v.logger.Warn("status code não OK em ViaCEP",
			slog.String("cep", cep),
			slog.Int("status", resp.StatusCode),
		)
		return nil, fmt.Errorf(
			"ViaCEP retornou %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var apiResp viaCEPResponse
	if err := json.NewDecoder(limitedBody).Decode(&apiResp); err != nil {
		v.logger.Error("erro ao parsear resposta ViaCEP",
			slog.String("cep", cep),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("erro ao parsear resposta: %w", err)
	}

	if apiResp.CEP == "" || apiResp.Logradouro == "" || apiResp.Localidade == "" {
		v.logger.Warn("resposta incompleta de ViaCEP",
			slog.String("cep", cep),
			slog.String("api_cep", apiResp.CEP),
			slog.String("logradouro", apiResp.Logradouro),
			slog.String("localidade", apiResp.Localidade),
		)
		return nil, fmt.Errorf("CEP não encontrado ou resposta incompleta: %s", cep)
	}

	v.logger.Info("sucesso em ViaCEP",
		slog.String("cep", cep),
		slog.String("city", apiResp.Localidade),
	)

	return &models.Address{
		CEP:        strings.ReplaceAll(apiResp.CEP, "-", ""),
		Logradouro: apiResp.Logradouro,
		Bairro:     apiResp.Bairro,
		Cidade:     apiResp.Localidade,
		Estado:     apiResp.UF,
	}, nil
}

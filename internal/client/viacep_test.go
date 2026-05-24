package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newViaCEPWithBase(baseURL string) *ViaCEP {
	return &ViaCEP{
		client:  &http.Client{Timeout: 2 * time.Second},
		baseURL: baseURL,
		logger:  logger,
	}
}

func TestViaCEP_Query_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(viaCEPResponse{
			CEP:        "01310-100",
			Logradouro: "Avenida Paulista",
			Bairro:     "Bela Vista",
			Localidade: "São Paulo",
			UF:         "SP",
		})
	}))
	defer srv.Close()

	api := newViaCEPWithBase(srv.URL)
	addr, err := api.Query(context.Background(), "01310100")
	if err != nil {
		t.Fatalf("esperado sucesso, got: %v", err)
	}
	if addr.Cidade != "São Paulo" {
		t.Errorf("Cidade esperada 'São Paulo', got: %s", addr.Cidade)
	}
	if addr.Estado != "SP" {
		t.Errorf("Estado esperado 'SP', got: %s", addr.Estado)
	}
	if addr.Logradouro != "Avenida Paulista" {
		t.Errorf("Logradouro esperado 'Avenida Paulista', got: %s", addr.Logradouro)
	}
}

func TestViaCEP_Query_CEPHyphenRemoved(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(viaCEPResponse{
			CEP:        "01310-100",
			Logradouro: "Av Paulista",
			Localidade: "São Paulo",
			UF:         "SP",
		})
	}))
	defer srv.Close()

	api := newViaCEPWithBase(srv.URL)
	addr, err := api.Query(context.Background(), "01310100")
	if err != nil {
		t.Fatalf("esperado sucesso, got: %v", err)
	}
	if addr.CEP != "01310100" {
		t.Errorf("CEP deve ser sem hífen, got: %s", addr.CEP)
	}
}

func TestViaCEP_Query_InvalidCEP(t *testing.T) {
	api := newViaCEPWithBase("http://localhost")
	_, err := api.Query(context.Background(), "123")
	if err == nil {
		t.Fatal("esperado erro para CEP inválido")
	}
}

func TestViaCEP_Query_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()

	api := newViaCEPWithBase(srv.URL)
	_, err := api.Query(context.Background(), "00000000")
	if err == nil {
		t.Fatal("esperado erro para HTTP 400")
	}
}

func TestViaCEP_Query_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("nao e json"))
	}))
	defer srv.Close()

	api := newViaCEPWithBase(srv.URL)
	_, err := api.Query(context.Background(), "01310100")
	if err == nil {
		t.Fatal("esperado erro para JSON inválido")
	}
}

func TestViaCEP_Query_IncompleteResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(viaCEPResponse{
			CEP: "01310-100",
			// Logradouro e Localidade ausentes
		})
	}))
	defer srv.Close()

	api := newViaCEPWithBase(srv.URL)
	_, err := api.Query(context.Background(), "01310100")
	if err == nil {
		t.Fatal("esperado erro para resposta incompleta")
	}
}

func TestViaCEP_Query_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer srv.Close()

	api := newViaCEPWithBase(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := api.Query(ctx, "01310100")
	if err == nil {
		t.Fatal("esperado erro para contexto cancelado")
	}
}

package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/henrique-avice/goexpert-desafio-multithreading/internal/models"
)

// mockClient implementa CEPClient para testes
type mockClient struct {
	addr  *models.Address
	err   error
	delay time.Duration
}

func (m *mockClient) Query(ctx context.Context, cep string) (*models.Address, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return m.addr, m.err
}

var successAddr = &models.Address{
	CEP:        "01310100",
	Logradouro: "Avenida Paulista",
	Bairro:     "Bela Vista",
	Cidade:     "São Paulo",
	Estado:     "SP",
}

// validateCEP

func TestValidateCEP_Valid(t *testing.T) {
	cep, err := validateCEP("01310100")
	if err != nil {
		t.Fatalf("esperado sucesso, got: %v", err)
	}
	if cep != "01310100" {
		t.Errorf("esperado '01310100', got: %s", cep)
	}
}

func TestValidateCEP_WithHyphen(t *testing.T) {
	cep, err := validateCEP("01310-100")
	if err != nil {
		t.Fatalf("esperado sucesso com hífen, got: %v", err)
	}
	if cep != "01310100" {
		t.Errorf("esperado '01310100' sem hífen, got: %s", cep)
	}
}

func TestValidateCEP_TooShort(t *testing.T) {
	_, err := validateCEP("1234567")
	if err == nil {
		t.Fatal("esperado erro para CEP curto")
	}
}

func TestValidateCEP_TooLong(t *testing.T) {
	_, err := validateCEP("123456789")
	if err == nil {
		t.Fatal("esperado erro para CEP longo")
	}
}

func TestValidateCEP_NonNumeric(t *testing.T) {
	_, err := validateCEP("0131010A")
	if err == nil {
		t.Fatal("esperado erro para CEP com letras")
	}
}

func TestValidateCEP_Empty(t *testing.T) {
	_, err := validateCEP("")
	if err == nil {
		t.Fatal("esperado erro para CEP vazio")
	}
}

func TestValidateCEP_WithSpaces(t *testing.T) {
	// Espaços não são removidos — deve falhar por tamanho ou caractere inválido
	_, err := validateCEP("0131 0100")
	if err == nil {
		t.Fatal("esperado erro para CEP com espaço")
	}
}

// formatCEP

func TestFormatCEP_Valid(t *testing.T) {
	result := formatCEP("01310100")
	if result != "01310-100" {
		t.Errorf("esperado '01310-100', got: %s", result)
	}
}

func TestFormatCEP_TooShort(t *testing.T) {
	result := formatCEP("0131")
	if result != "0131" {
		t.Errorf("esperado retorno sem modificação '0131', got: %s", result)
	}
}

func TestFormatCEP_TooLong(t *testing.T) {
	result := formatCEP("013101001")
	if result != "013101001" {
		t.Errorf("esperado retorno sem modificação, got: %s", result)
	}
}

// getCEPFast

func TestGetCEPFast_FirstAPIWins(t *testing.T) {
	fast := &mockClient{addr: successAddr, delay: 0}
	slow := &mockClient{addr: successAddr, delay: 200 * time.Millisecond}

	// Substituir clients diretamente no getCEPFast via helper
	addr, source, err := getCEPFastWithClients("01310100", time.Second, fast, "BrasilAPI", slow, "ViaCEP")
	if err != nil {
		t.Fatalf("esperado sucesso, got: %v", err)
	}
	_ = addr
	_ = source
}

func TestGetCEPFast_SecondAPIWins(t *testing.T) {
	slow := &mockClient{addr: successAddr, delay: 200 * time.Millisecond}
	fast := &mockClient{addr: successAddr, delay: 0}

	addr, _, err := getCEPFastWithClients("01310100", time.Second, slow, "BrasilAPI", fast, "ViaCEP")
	if err != nil {
		t.Fatalf("esperado sucesso, got: %v", err)
	}
	if addr.Cidade != "São Paulo" {
		t.Errorf("Cidade esperada, got: %s", addr.Cidade)
	}
}

func TestGetCEPFast_BothFail(t *testing.T) {
	fail1 := &mockClient{err: errors.New("API 1 falhou")}
	fail2 := &mockClient{err: errors.New("API 2 falhou")}

	_, _, err := getCEPFastWithClients("01310100", time.Second, fail1, "BrasilAPI", fail2, "ViaCEP")
	if err == nil {
		t.Fatal("esperado erro quando ambas as APIs falham")
	}
}

func TestGetCEPFast_Timeout(t *testing.T) {
	slow1 := &mockClient{addr: successAddr, delay: 500 * time.Millisecond}
	slow2 := &mockClient{addr: successAddr, delay: 500 * time.Millisecond}

	_, _, err := getCEPFastWithClients("01310100", 50*time.Millisecond, slow1, "BrasilAPI", slow2, "ViaCEP")
	if err == nil {
		t.Fatal("esperado erro de timeout")
	}
}

func TestGetCEPFast_OneFailsOtherSucceeds(t *testing.T) {
	fail := &mockClient{err: errors.New("API falhou")}
	ok := &mockClient{addr: successAddr, delay: 10 * time.Millisecond}

	addr, _, err := getCEPFastWithClients("01310100", time.Second, fail, "BrasilAPI", ok, "ViaCEP")
	if err != nil {
		t.Fatalf("esperado sucesso quando uma API funciona, got: %v", err)
	}
	if addr.Cidade != "São Paulo" {
		t.Errorf("Cidade esperada 'São Paulo', got: %s", addr.Cidade)
	}
}

func TestGetCEPFast_InvalidCEP(t *testing.T) {
	_, _, err := getCEPFastWithClients("abc", time.Second, &mockClient{}, "A", &mockClient{}, "B")
	if err == nil {
		t.Fatal("esperado erro para CEP inválido")
	}
}

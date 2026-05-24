package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/henrique-avice/goexpert-desafio-multithreading/internal/client"
	"github.com/henrique-avice/goexpert-desafio-multithreading/internal/models"
)

func main() {
	cepFlag := flag.String("cep", "", "CEP a consultar (com ou sem hífen)")
	timeoutFlag := flag.Float64("timeout", 1.0, "Timeout em segundos")
	flag.Parse()

	if *cepFlag == "" {
		fmt.Println("Uso: cepfinder -cep=29902555 [-timeout=1.0]")
		return
	}

	timeout := time.Duration(*timeoutFlag * float64(time.Second))
	addr, source, err := getCEPFast(*cepFlag, timeout)

	if err != nil {
		printError(err)
		return
	}

	printSuccess(addr, source)
}

func getCEPFast(cepInput string, timeout time.Duration) (*models.Address, string, error) {
	httpClient := &http.Client{Timeout: 2 * time.Second}
	return getCEPFastWithClients(
		cepInput, timeout,
		client.NewBrasilAPI(httpClient), "BrasilAPI",
		client.NewViaCEP(httpClient), "ViaCEP",
	)
}

func getCEPFastWithClients(
	cepInput string,
	timeout time.Duration,
	c1 client.CEPClient, source1 string,
	c2 client.CEPClient, source2 string,
) (*models.Address, string, error) {
	cep, err := validateCEP(cepInput)
	if err != nil {
		return nil, "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type result struct {
		addr   *models.Address
		source string
		err    error
	}

	entries := []struct {
		c      client.CEPClient
		source string
	}{
		{c1, source1},
		{c2, source2},
	}

	ch := make(chan result, len(entries))

	for _, entry := range entries {
		entry := entry
		go func() {
			addr, err := entry.c.Query(ctx, cep)
			ch <- result{addr: addr, source: entry.source, err: err}
		}()
	}

	failureCount := 0
	var lastErr error

	for range entries {
		res := <-ch
		if res.err == nil && res.addr != nil {
			return res.addr, res.source, nil
		}
		lastErr = res.err
		failureCount++
		if failureCount >= len(entries) {
			return nil, "", fmt.Errorf("ambas as APIs falharam: %w", lastErr)
		}
	}

	return nil, "", fmt.Errorf("timeout: nenhuma API respondeu em %v", timeout)
}

func validateCEP(cepInput string) (string, error) {
	cep := strings.ReplaceAll(cepInput, "-", "")

	if len(cep) != 8 {
		return "", fmt.Errorf("CEP inválido: deve conter 8 dígitos (recebido: %s)", cepInput)
	}

	for _, ch := range cep {
		if ch < '0' || ch > '9' {
			return "", fmt.Errorf("CEP inválido: deve conter apenas dígitos (recebido: %s)", cepInput)
		}
	}

	return cep, nil
}

func printSuccess(addr *models.Address, source string) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("CEP Consultado: %s\n", formatCEP(addr.CEP))
	fmt.Println("\nEndereço:")
	fmt.Printf("  Logradouro: %s\n", addr.Logradouro)
	fmt.Printf("  Bairro:     %s\n", addr.Bairro)
	fmt.Printf("  Cidade:     %s\n", addr.Cidade)
	fmt.Printf("  Estado:     %s\n", addr.Estado)
	fmt.Printf("\nAPI Vencedora: %s\n", source)
	fmt.Println(strings.Repeat("=", 60) + "\n")
}

func printError(err error) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("Erro: %v\n", err)
	fmt.Println(strings.Repeat("=", 60) + "\n")
}

func formatCEP(cep string) string {
	if len(cep) != 8 {
		return cep
	}
	return cep[:5] + "-" + cep[5:]
}

# goexpert-desafio-multithreading

> Consulta simultânea de CEP em duas APIs externas; a primeira a responder vence.

## Índice

- [Visão Geral](#visão-geral)
- [Funcionalidades](#funcionalidades)
- [Requisitos](#requisitos)
- [Execução](#execução)
- [Arquitetura](#arquitetura)
- [Testes](#testes)
- [Como Utilizar](#como-utilizar)

## Visão Geral

Ferramenta CLI que dispara consultas simultâneas à **BrasilAPI** e à **ViaCEP** usando goroutines. A primeira resposta válida recebida é exibida junto ao nome da API vencedora. Se nenhuma responder dentro do timeout, o programa encerra com erro.

## Funcionalidades

### Requisitos do Desafio

- [x] Consulta simultânea à BrasilAPI e à ViaCEP
- [x] Exibe o resultado da API que responder primeiro
- [x] Descarta a resposta mais lenta
- [x] Exibe qual API respondeu primeiro
- [x] Timeout de 1 segundo (configurável via flag)
- [x] Exibe o endereço completo (logradouro, bairro, cidade, estado)

### Extras Implementados

- Flag `--timeout` para ajustar o limite de tempo (padrão: 1.0s)
- Validação do CEP antes da consulta (8 dígitos numéricos; aceita formato com ou sem hífen)

## Requisitos

- Go 1.26.2+
- Docker e Docker Compose

## Execução

### Docker Compose (Recomendado)

```bash
docker-compose up --build
```

Executa a consulta com o CEP padrão configurado (`29902555`).

### Docker

```bash
docker build -t cepfinder .
docker run cepfinder -cep=01310100
```

### Local

```bash
go run ./cmd/cepfinder -cep=01310100
go run ./cmd/cepfinder -cep=01310-100 -timeout=2.0
```

## Arquitetura

```
[CLI -cep=xxxxx] ──► [getCEPFastWithClients]
                        ┌──────┴──────┐
                   [BrasilAPI]    [ViaCEP]
                   goroutine 1   goroutine 2
                        └──────┬──────┘
                         [channel (cap 2)]
                     Primeira resposta válida vence
```

| Componente | Responsabilidade |
|---|---|
| `cmd/cepfinder` | Parsing de flags, orquestração e exibição do resultado |
| `internal/client` | Interface `CEPClient` e implementações BrasilAPI e ViaCEP |
| `internal/models` | Estrutura `Address` unificada |

## Testes

```bash
go test -v -race ./...
```

---

## Como Utilizar

### 1. Iniciando o Sistema

```bash
docker-compose up --build
```

Executa a consulta com o CEP padrão configurado (`29902555`) e encerra automaticamente após exibir o resultado.

### 2. Testando com CEP Personalizado

Via Docker (é necessário fazer o build antes do primeiro uso):

```bash
docker build -t cepfinder .
docker run --rm cepfinder -cep=01310100
docker run --rm cepfinder -cep=01310-100 -timeout=2.0
```

Localmente:

```bash
go run ./cmd/cepfinder -cep=01310100
go run ./cmd/cepfinder -cep=01310-100 -timeout=2.0
```

### 3. Resultado Esperado

```
============================================================
CEP Consultado: 01310-100

Endereço:
  Logradouro: Avenida Paulista
  Bairro:     Bela Vista
  Cidade:     São Paulo
  Estado:     SP

API Vencedora: BrasilAPI
============================================================
```

A **API Vencedora** varia — pode ser `BrasilAPI` ou `ViaCEP`, dependendo de qual responder primeiro. Se nenhuma responder dentro do timeout (padrão: 1s), o programa exibe uma mensagem de erro e encerra com código de saída não-zero.

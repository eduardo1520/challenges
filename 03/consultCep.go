package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ViaCEP struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Unidade     string `json:"unidade"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Estado      string `json:"estado"`
	Regiao      string `json:"regiao"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type BrasilAPI struct {
	Cep     string `json:"cep"`
	State   string `json:"state"`
	City    string `json:"city"`
	Bairro  string `json:"bairro"`
	Street  string `json:"street"`
	Service string `json:"service"`
}

func fetchCep(ctx context.Context, url string, endpointName string, ch chan<- string) {
	// Cria uma requisição HTTP com o contexto (suporta timeout)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- fmt.Sprintf("Erro ao criar requisição para %s: %v", endpointName, err)
		return
	}

	// Executa a requisição HTTP
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- fmt.Sprintf("Erro na requisição para %s: %v", endpointName, err)
		return
	}
	defer resp.Body.Close()

	// Lê o corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- fmt.Sprintf("Erro ao ler resposta de %s: %v", endpointName, err)
		return
	}

	if endpointName == "ViaCEP" {
		var data ViaCEP
		if err := json.Unmarshal(body, &data); err != nil {
			ch <- fmt.Sprintf("Erro ao parsear JSON de %s: %v", endpointName, err)
			return
		}
		// Envia o logradouro junto com o nome do endpoint para o canal
		ch <- fmt.Sprintf("Endpoint mais rápido: %s\nLogradouro: %s", endpointName, data)
	}

	if endpointName == "BrasilAPI" {
		var data BrasilAPI
		if err := json.Unmarshal(body, &data); err != nil {
			ch <- fmt.Sprintf("Erro ao parsear JSON de %s: %v", endpointName, err)
			return
		}
		// Envia o logradouro junto com o nome do endpoint para o canal
		ch <- fmt.Sprintf("Endpoint mais rápido: %s\nLogradouro: %s", endpointName, data)
	}

}

func main() {
	cep := "06342140"
	ch := make(chan string)

	// URLs dos endpoints
	endpoint1 := "https://brasilapi.com.br/api/cep/v1/"
	endpoint2 := "https://viacep.com.br/ws/"

	// Cria um contexto com timeout de 1 segundo
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Inicia goroutines para ambos os endpoints
	go fetchCep(ctx, endpoint1+cep, "BrasilAPI", ch)
	go fetchCep(ctx, endpoint2+cep+"/json/", "ViaCEP", ch)

	// Usa select para receber a primeira resposta ou um timeout
	select {
	case result := <-ch:
		fmt.Println(result)
	case <-ctx.Done():
		fmt.Println("Erro: timeout de 1 segundo excedido.")
	}
}

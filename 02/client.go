package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Os requisitos para cumprir este desafio são:
// O client.go deverá realizar uma requisição HTTP no server.go solicitando a cotação do dólar
// O client.go terá que salvar a cotação atual em um arquivo "cotacao.txt" no formato: Dólar: {valor}
// O endpoint necessário gerado pelo server.go para este desafio será: /cotacao e a porta a ser utilizada pelo servidor HTTP será a 8080..
// Definimos uma estrutura para receber o JSON da resposta
type Cotacao struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	url := "http://localhost:8080/cotacao"

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Erro ao fazer a requisição: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Erro ao ler a resposta: %v", err)
	}

	var cotacao Cotacao

	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		log.Fatalf("Erro ao desserializar o JSON: %v", err)
	}

	// Converte o valor de Bid de string para float64
	bidFloat, err := strconv.ParseFloat(cotacao.USDBRL.Bid, 64)
	if err != nil {
		log.Fatalf("Erro ao converter o Bid: %v", err)
	}

	fmt.Printf(" Dólar: %.2f\n", bidFloat)

	// Obtém a data e hora atuais
	dataHora := time.Now().Format("02-01-2006 15:04:05")

	fileContent := fmt.Sprintf("Dólar: %.2f -> consulta: %s\n", bidFloat, dataHora)

	file, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo: %v", err)
	}

	defer file.Close()

	if _, err := file.WriteString(fileContent); err != nil {
		log.Fatalf("Erro ao escrever no arquivo: %v", err)
	}

	fmt.Println("Cotação salva no arquivo cotacao.txt")

}

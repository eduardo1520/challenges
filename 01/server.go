package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// O server.go deverá consumir a API contendo o câmbio de Dólar e Real no endereço:
// https://economia.awesomeapi.com.br/json/last/USD-BRL e em seguida deverá retornar no formato JSON o resultado para o cliente.

// Usando o package "context",
// o server.go deverá registrar no banco de dados SQLite cada cotação recebida,
// sendo que o timeout máximo para chamar a API de cotação do dólar deverá ser de 200ms
// e o timeout máximo para conseguir persistir os dados no banco deverá ser de 10ms

// Estrutura para fazer o unmarshal da resposta da API

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
	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable(db)

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, db)
	})

	log.Println("Servidor iniciado na porta 8080")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()
	log.Println("Chamando API de cotação do dólar")

	defer log.Println("Cotação do Dólar obtida")

	select {
	case <-time.After(200 * time.Millisecond):
		cotacao, err := consultaCotacaoDolar()

		if err != nil {
			log.Println("Erro ao obter cotação:", err)
			http.Error(w, "Erro ao obter cotação", http.StatusInternalServerError)
			return
		}

		select {
		case <-time.After(10 * time.Millisecond):
			if err := registrarCotacao(db, cotacao); err != nil {
				log.Println("Erro ao registrar cotação:", err)
				http.Error(w, "Erro ao registrar cotação", http.StatusRequestTimeout)
				return
			}
		case <-ctx.Done():
			log.Println("Registro persistido com sucesso :)")
		}

		log.Println("Request processada com sucesso em 200ms")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(cotacao)

	case <-ctx.Done():
		log.Println("Request cancelada pelo cliente")
		http.Error(w, "Request cancelada pelo cliente", http.StatusRequestTimeout)
	}
}

func consultaCotacaoDolar() (*Cotacao, error) {
	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"

	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var cotacao Cotacao

	err = json.Unmarshal(body, &cotacao)

	if err != nil {
		return nil, err
	}
	return &cotacao, nil
}

func createTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT,
		bid TEXT,
		ask TEXT,
		timestamp TEXT
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Erro ao criar tabela cotacoes:", err)
	}
}

func registrarCotacao(db *sql.DB, cotacao *Cotacao) error {
	query := `INSERT INTO cotacoes(code, bid, ask, timestamp) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query, cotacao.USDBRL.Code, cotacao.USDBRL.Bid, cotacao.USDBRL.Ask, cotacao.USDBRL.Timestamp)
	return err
}

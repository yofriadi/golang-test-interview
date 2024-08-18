package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

var agreementLetterURL = "https://www.scribd.com/document/703635194/01HMR5RTKR2AZ2S37Z5GNZANDN"

func main() {
	ctx := context.Background()
	conn := connectDB(ctx)
	defer conn.Close()

	runMigration(ctx, conn)
	router := httprouter.New()
	router.POST("/user", handleCreateUser(ctx, conn))
	router.POST("/loan", handleCreateLoan(ctx, conn))
	router.POST("/loan/:id/approve", handleApproveLoan(ctx, conn))
	router.POST("/loan/:id/invest", handleInvestLoan(ctx, conn))
	router.POST("/loan/:id/disburse", handleDisburseLoan(ctx, conn))

	fmt.Println("Server is running at http://localhost:8123")
	log.Fatal(http.ListenAndServe(":8123", router))
}

func connectDB(ctx context.Context) *pgxpool.Pool {
	conn, err := pgxpool.New(
		ctx,
		"postgres://postgres:postgres@db:5432/postgres?sslmode=disable",
	)
	if err != nil {
		log.Fatal(err)
	}

	return conn
}

type response struct {
	Message string `json:"message"`
}

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func handleCreateUser(
	ctx context.Context,
	conn *pgxpool.Pool,
) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		type user struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}
		var u user
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			log.Fatal(err)
		}

		_, err = conn.Exec(
			ctx,
			`INSERT INTO users (name, type) VALUES ($1, $2) RETURNING id;`,
			u.Name,
			u.Type,
		)
		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response{Message: "success create user"})
	}
}

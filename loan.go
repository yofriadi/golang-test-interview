package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func handleCreateLoan(
	ctx context.Context,
	conn *pgxpool.Pool,
) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		type request struct {
			UserID int `json:"userId"`
			Amount int `json:"amount"`
		}
		var req request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Fatal(err)
		}

		rows, err := conn.Query(
			ctx,
			`INSERT INTO loans (user_id, amount) VALUES ($1, $2) RETURNING id;`,
			req.UserID,
			req.Amount,
		)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var id int
		for rows.Next() {
			err = rows.Scan(&id)
			if err != nil {
				log.Fatal(err)
			}
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		type response struct {
			ID     int `json:"id"`
			UserID int `json:"userId"`
			Amount int `json:"amount"`
		}
		var res response
		err = conn.QueryRow(ctx, `SELECT id, user_id, amount FROM loans WHERE id = $1;`, id).
			Scan(&res.ID, &res.UserID, &res.Amount)
		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

func handleApproveLoan(
	ctx context.Context,
	conn *pgxpool.Pool,
) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		type request struct {
			ImageURLBorrowerVisited string `json:"imageUrlBorrowerVisited"`
			EmployeeID              int    `json:"employeeId"`
		}
		var req request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Fatal(err)
		}

		_, err = conn.Query(
			ctx,
			`UPDATE loans SET image_url_borrower_visited = $1, approved_by = $2, status = 'approved' WHERE id = $3;`,
			req.ImageURLBorrowerVisited,
			req.EmployeeID,
			ps.ByName("id"),
		)
		if err != nil {
			log.Fatal(err)
		}

		type response struct {
			ID                      int    `json:"id"`
			UserID                  int    `json:"userId"`
			Amount                  int    `json:"amount"`
			Status                  string `json:"status"`
			ImageURLBorrowerVisited string `json:"imageUrlBorrowerVisited"`
			ApprovedBy              int    `json:"approvedBy"`
		}
		var res response
		err = conn.QueryRow(
			ctx,
			`SELECT id, user_id, amount, status, image_url_borrower_visited, approved_by FROM loans WHERE id = $1;`,
			ps.ByName("id"),
		).Scan(
			&res.ID,
			&res.UserID,
			&res.Amount,
			&res.Status,
			&res.ImageURLBorrowerVisited,
			&res.ApprovedBy,
		)
		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

func handleInvestLoan(
	ctx context.Context,
	conn *pgxpool.Pool,
) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var (
			principalAmount int
			status          string
		)
		err := conn.QueryRow(ctx, `SELECT amount, status FROM loans WHERE id = $1;`, ps.ByName("id")).
			Scan(&principalAmount, &status)
		if err != nil {
			log.Fatal(err)
		}

		if status == "proposed" {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response{Message: "loan is not approved yet"})
			return
		}

		if status == "disbursed" {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response{Message: "loan is already disbursed"})
			return
		}

		if status == "approved" {
			_, err = conn.Query(
				ctx,
				`UPDATE loans SET status = 'invested' WHERE id = $1;`,
				ps.ByName("id"),
			)
			if err != nil {
				log.Fatal(err)
			}
		}

		rows, err := conn.Query(
			ctx,
			`SELECT amount FROM loan_transactions WHERE loan_id = $1;`,
			ps.ByName("id"),
		)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var invested int
		for rows.Next() {
			var amount int
			err = rows.Scan(&amount)
			if err != nil {
				log.Fatal(err)
			}
			invested += amount
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		type request struct {
			Amount int `json:"amount"`
			UserID int `json:"user_id"`
		}
		var req request
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Fatal(err)
		}

		if invested+req.Amount > principalAmount {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response{Message: "invest amount exceeds loan amount"})

			return
		}

		rows, err = conn.Query(
			ctx,
			`INSERT INTO loan_transactions (loan_id, user_id, amount) VALUES ($3, $1, $2) RETURNING loan_id, amount`,
			req.UserID,
			req.Amount,
			ps.ByName("id"),
		)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		type response struct {
			LoanID             int    `json:"loanId"`
			Amount             int    `json:"amount"`
			AgreementLetterURL string `json:"agreementLetter"`
			Interests          string `json:"interests"`
			ProfitReturn       int    `json:"profitReturn"`
		}
		var res response
		for rows.Next() {
			err = rows.Scan(&res.LoanID, &res.Amount)
			if err != nil {
				log.Fatal(err)
			}
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		res.AgreementLetterURL = agreementLetterURL
		res.Interests = "10%"
		res.ProfitReturn = req.Amount * 10 / 100

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

func handleDisburseLoan(
	ctx context.Context,
	conn *pgxpool.Pool,
) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		type request struct {
			AgreementLetterURL string `json:"agreementLetterURL"`
			EmployeeID         int    `json:"employeeId"`
		}
		var req request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Fatal(err)
		}

		_, err = conn.Query(
			ctx,
			`UPDATE loans SET agreement_letter_url = $1, disbursed_by = $2, status = 'disbursed' WHERE id = $3;`,
			req.AgreementLetterURL,
			req.EmployeeID,
			ps.ByName("id"),
		)
		if err != nil {
			log.Fatal(err)
		}

		type response struct {
			ID                      int    `json:"id"`
			UserID                  int    `json:"userId"`
			Amount                  int    `json:"amount"`
			Status                  string `json:"status"`
			ImageURLBorrowerVisited string `json:"imageUrlBorrowerVisited"`
			ApprovedBy              int    `json:"approvedBy"`
			AgreementLetterURL      string `json:"agreementLetterURL"`
			DisbursedBy             int    `json:"disbursedBy"`
		}
		var res response
		err = conn.QueryRow(
			ctx,
			`SELECT id, user_id, amount, status, image_url_borrower_visited, approved_by, agreement_letter_url, disbursed_by
			FROM loans
			WHERE id = $1;`,
			ps.ByName("id"),
		).Scan(
			&res.ID,
			&res.UserID,
			&res.Amount,
			&res.Status,
			&res.ImageURLBorrowerVisited,
			&res.ApprovedBy,
			&res.AgreementLetterURL,
			&res.DisbursedBy,
		)
		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

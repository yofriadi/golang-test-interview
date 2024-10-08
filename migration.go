package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func runMigration(ctx context.Context, conn *pgxpool.Pool) {
	_, err := conn.Exec(ctx, `
		CREATE TYPE user_type AS ENUM ('borrower', 'investor', 'employee');
		CREATE TABLE IF NOT EXISTS users (
		    id int PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
		    name varchar(100) NOT NULL,
			type user_type NOT NULL
		);

		CREATE TYPE loan_status AS ENUM ('proposed', 'approved', 'invested', 'disbursed');
		CREATE TABLE IF NOT EXISTS loans (
	    	id int PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
	    	user_id int NOT NULL REFERENCES users (id),
	    	amount int NOT NULL,
	    	status loan_status NOT NULL DEFAULT 'proposed',
	    	image_url_borrower_visited text,
	    	approved_by int REFERENCES users(id),
	    	agreement_letter_url text,
	    	disbursed_by int REFERENCES users(id)
		);

	    CREATE TABLE IF NOT EXISTS loan_transactions (
	      	id int PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
	    	loan_id int REFERENCES loans(id),
	    	user_id int REFERENCES users(id),
	    	amount int
	    );
	`)
	if err != nil {
		fmt.Println(err)
	}
}

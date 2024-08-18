### How to run

make sure you have docker installed, then run

```bash
docker compose up --build
```

or just run `go run main.go` but setup Postgres database yourself

### How to test

creating user, there is 3 user role employee, borrower, investor; pick one.

```bash
curl --location 'http://localhost:8123/user' \
--header 'Content-Type: application/json' \
--data '{
 "name": "Borrower",
  "type": "borrower"
}'
```

create loan, for simplicity, we pass user id in payload

```bash
curl --location 'http://localhost:8123/loan' \
--header 'Content-Type: application/json' \
--data '{
 "user_id": 1,
  "amount": 1000000
}'
```

approve loan, create another user as employee

```bash
curl --location 'http://localhost:8123/loan/5/approve' \
--header 'Content-Type: application/json' \
--data '{
  "imageUrlBorrowerVisited": "https://amartha.com/_next/image/?url=https%3A%2F%2Faccess.amartha.com%2Fuploads%2FDukungan_UMKM_dengan_Pinjaman_Modal_Usaha_Amartha_com_3_c78f6fffdc.png&w=3840&q=75",
  "employeeId": 2
}'
```

invest loan, create another user as investor

```bash
curl --location 'http://localhost:8123/loan/1/invest' \
--header 'Content-Type: application/json' \
--data '{
  "user_id": 3,
  "amount": 100000
}
```

disburse loan

```bash
curl --location 'http://localhost:8123/loan/1/disburse' \
--header 'Content-Type: application/json' \
--data '{
  "agreementLetterUrl": "https://www.scribd.com/document/703635194/01HMR5RTKR2AZ2S37Z5GNZANDN",
  "employeeId": 2
}'
```

### Notes

it is a bare minimum mvp, needs to add locks, transactions, etc. it is far from perfect, but you get the idea.

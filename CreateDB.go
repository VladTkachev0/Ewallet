package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"math/rand"
)

func randAdress() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	n := new(big.Int)
	n.SetBytes(b)
	return n.Text(62)[:10]
}

func createTableWallet(db *sql.DB) {
	wallet_table := `CREATE TABLE IF NOT EXISTS wallet (
        "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        "adress" TEXT,
        "money" FLOAT
        );`
	query, err := db.Prepare(wallet_table)
	if err != nil {
		log.Fatal(err)
	}
	query.Exec()
	fmt.Println("Table wallet successfully!")
}

func createTableTransfer(db *sql.DB) {
	transfer_table := `CREATE TABLE IF NOT EXISTS transfer (
        "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        "adress_one" TEXT REFERENCES wallet (wallet_adress),
        "adress_two" TEXT REFERENCES wallet (wallet_adress),
        "sum" FLOAT
        );`
	query1, err1 := db.Prepare(transfer_table)
	if err1 != nil {
		log.Fatal(err1)
	}
	query1.Exec()
	fmt.Println("Table transfer successfully!")
}

func insertWallet(db *sql.DB) {
	statement, err := db.Prepare("INSERT INTO wallet(adress,  money) VALUES (?, ?)")
	checkErr(err)
	for i := 0; i < 10; i++ {
		_, err = statement.Exec(randAdress(), 100.00)
		checkErr(err)
	}

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}

}

func firstCreate() bool {
	row, err := database.Query("SELECT * FROM wallet")
	checkErr(err)
	defer row.Close()

	if !row.Next() {
		return true
	}
	return false
}

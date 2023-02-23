package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net/http"

	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Wallet struct {
	ID     int     `json:"id"`
	ADRESS string  `json:"adress"`
	MONEY  float32 `json:"money"`
}

type Transfer struct {
	ID         int     `json:"id"`
	ADRESS_ONE string  `json:"adress_one"`
	ADRESS_TWO string  `json:"adress_two"`
	SUM        float32 `json:"sum"`
}

var database *sql.DB

func main() {

	db, err := sql.Open("sqlite3", "./database.db")
	checkErr(err)
	defer db.Close()
	database = db

	createTableWallet(db)
	createTableTransfer(db)
	if firstCreate() {
		insertWallet(db)
	}

	r := mux.NewRouter()
	//getBalance
	r.HandleFunc("/api/wallet/{adress}/balance", getBalance).Methods("GET")
	//send
	r.HandleFunc("/api/send", send).Methods("POST")
	//getLast
	r.HandleFunc("/api/transactions", getLast).Methods("GET")

	log.Fatal(http.ListenAndServe(":8888", r))

}

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

func getBalance(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	rows, err := database.Query("select * from Wallet")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	wallet := []Wallet{}

	for rows.Next() {
		p := Wallet{}
		err := rows.Scan(&p.ID, &p.ADRESS, &p.MONEY)
		if err != nil {
			fmt.Println(err)
			continue
		}
		wallet = append(wallet, p)
	}

	params := mux.Vars(r)
	for _, item := range wallet {
		if item.ADRESS == params["adress"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(database)

}

func send(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var transfer Transfer

	err := json.NewDecoder(r.Body).Decode(&transfer)
	if err != nil {
		json.NewEncoder(w).Encode("Incorrect amount")
		return
	}

	if transfer.SUM < 0 {
		json.NewEncoder(w).Encode("Negative amount entered")
		return
	}

	fmt.Println("Insert and Update database")

	statement1, err1 := database.Prepare("INSERT INTO transfer (adress_one, adress_two, sum) values (?,?,?)")
	checkErr(err1)
	statement2, err2 := database.Prepare("UPDATE wallet SET money = money - ? where adress = ?")
	checkErr(err2)
	statement3, err3 := database.Prepare("UPDATE wallet SET money = money + ? where adress = ?")
	checkErr(err3)

	_, err1 = statement1.Exec(transfer.ADRESS_ONE, transfer.ADRESS_TWO, transfer.SUM)
	checkErr(err1)
	_, err2 = statement2.Exec(transfer.SUM, transfer.ADRESS_ONE)
	//checkErr(err2)
	if err2 != nil {
		json.NewEncoder(w).Encode("На кошельке недостаточно средств")
		return
	}
	_, err3 = statement3.Exec(transfer.SUM, transfer.ADRESS_TWO)
	checkErr(err3)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getLast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	count := query.Get("count")

	count2, _ := strconv.Atoi(count)

	rows, err := database.Query("SELECT * FROM transfer")

	checkErr(err)

	var transfers []Transfer

	for rows.Next() {
		var id int
		var adress_one string
		var adress_two string
		var sum float32

		err = rows.Scan(&id, &adress_one, &adress_two, &sum)

		checkErr(err)

		transfers = append(transfers, Transfer{ID: id, ADRESS_ONE: adress_one, ADRESS_TWO: adress_two, SUM: sum})
	}

	err = json.NewEncoder(w).Encode(transfers[len(transfers)-count2:])
	if err != nil {
		json.NewEncoder(w).Encode("Such a number of transfers have not yet been made")
		return
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

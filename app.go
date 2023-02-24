package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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

	if transfer.ADRESS_ONE == transfer.ADRESS_TWO {
		json.NewEncoder(w).Encode("The same wallets")
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
		json.NewEncoder(w).Encode("There are not enough funds in the wallet")
		return
	}
	_, err3 = statement3.Exec(transfer.SUM, transfer.ADRESS_TWO)
	checkErr(err3)

	json.NewEncoder(w).Encode("Translation done")

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

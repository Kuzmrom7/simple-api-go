package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

//Users структура для парсинга json
type Users struct {
	Id        string `json:id`
	Name      string `json:"name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

//go get -u github.com/gorilla/mux
var (
	db *sql.DB
)

//PORT Порт
const PORT string = ":8080"
const DB_CONNECT_STRING =
	"host= 172.17.0.2 port=5432 user=postgres  password= docker dbname=clientserver sslmode=disable"

func init() {
	var err error
	db, err = sql.Open("postgres", DB_CONNECT_STRING)
	if err != nil {
		log.Panic(err)
	}

	if err = db.Ping(); err != nil {
		log.Panic(err)
	}
}

func main() {

	//Run server and routes
	r := mux.NewRouter()

	//Получить всех пользователей
	r.HandleFunc("/user", GetUsers).Methods("GET")
	//Создать пользователя
	r.HandleFunc("/user", CreateUsers).Methods("POST")
	//Удалить пользователя
	r.HandleFunc("/user/{id:[0-9]+}", DeleteUser).Methods("DELETE")
	//Получить пользователя по id
	r.HandleFunc("/user/{id:[0-9]+}", GetUserById).Methods("GET")
	//Обновить пользователя
	r.HandleFunc("/user/{id:[0-9]+}", UpdateUser).Methods("PUT")

	log.Println("Server up and run on port " + PORT)
	log.Fatal(http.ListenAndServe(PORT, r))

}

func GetUsers(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	defer rows.Close()

	users := make([]*Users, 0)

	for rows.Next() {
		us := new(Users)
		err = rows.Scan(&us.Id, &us.Name, &us.FirstName, &us.LastName)
		PanicOnErr(err)
		users = append(users, us)
	}
	PanicOnErr(err)
	productsJson, _ := json.Marshal(users)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(productsJson)
}

func CreateUsers(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	user := Users{}

	err := decoder.Decode(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO users (username, first_name, last_name) VALUES ($1, $2, $3)", user.Name, user.FirstName, user.LastName)
	PanicOnErr(err)

	lastInsertId,err := result.LastInsertId()
	fmt.Println(lastInsertId)
	w.WriteHeader(http.StatusOK)

}

func GetUserById(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/user/"):]
	index, _ := strconv.ParseInt(id, 10, 0)

	row := db.QueryRow("SELECT * FROM users WHERE id = $1", index)

	us := new(Users)

	err := row.Scan(&us.Id, &us.Name, &us.FirstName, &us.LastName)
	PanicOnErr(err)

	productsJson, _ := json.Marshal(us)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(productsJson)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	//deleteUsers
	id := r.URL.Path[len("/user/"):]
	index, _ := strconv.ParseInt(id, 10, 0)

	result, err := db.Exec("DELETE FROM users WHERE id = $1", index)
	PanicOnErr(err)

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "User %s delete successfully (%d row affected)\n", id, rowsAffected)
	w.WriteHeader(http.StatusOK)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	//update
	id := r.URL.Path[len("/user/"):]
	index, _ := strconv.ParseInt(id, 10, 0)

	decoder := json.NewDecoder(r.Body)
	user := Users{}
	err := decoder.Decode(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := db.Exec("UPDATE users SET username = $1, first_name = $2, last_name = $3  WHERE id = $4", user.Name, user.FirstName, user.LastName, index)
	PanicOnErr(err)

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "User %s update successfully (%d row affected)\n", id, rowsAffected)
	w.WriteHeader(http.StatusOK)

}

//PanicOnErr panics on error
func PanicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

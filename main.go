package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"log"
	"net/http"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var db *sql.DB

// define a user pulled from database (has Id)
type User struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
}

// define a request for an email based on ID
type EmailReq struct {
	Id int `json:"id"`
	Game string `json:"game"`
}

// runs automagically when the file is built and rain
func main() {
	config := buildDbConfig()
	connectToDb(config)

	// these are needed to make the DB stay connect
	// no idea why
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(0)
	handleRequests()
}

// returns env variable for local testing (will be commented out for deploy since
// Heroku env variables will be used)
// func goDotEnvVariable(key string) string {
// 	err := godotenv.Load(".env")

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	return os.Getenv(key)
// }

// builds DB config (will be commented out for deploy)
// same function using Heroku config vars will be built
// func buildDbConfig() mysql.Config {
// 	return mysql.Config{
// 		User: goDotEnvVariable("User"),
// 		Passwd: goDotEnvVariable("Password"),
// 		Net: "tcp",
// 		Addr: goDotEnvVariable("Address"),
// 		DBName: goDotEnvVariable("DB"),
// 		AllowNativePasswords: true,
// 	}
// }

func buildDbConfig() mysql.Config {
	return mysql.Config {
		User: os.Getenv("User"),
		Passwd: os.Getenv("Password"),
		Net: "tcp",
		Addr: os.Getenv("Address"),
		DBName: os.Getenv("DB"),
		AllowNativePasswords: true,
	}
}

// establishes connection to DB
func connectToDb(config mysql.Config) {
	var err error
	db, err = sql.Open("mysql", config.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("You connected boiiii")
}

// inserts new user into the DB
func insertDbRow(name string, email string) int64 {
	command := fmt.Sprintf("INSERT INTO users(name, email) VALUES('%s', '%s')", name, email)
	res, err := db.Exec(command)

	if err != nil {
		log.Fatal(err)
	}

	id, err := res.LastInsertId()

	if err != nil {
		panic(err.Error())
	}

	// returns id back to calling function so json can be assembled
	return id
}

// gets a user from the DB based on id and maps values to a user struct
func queryDbAndBuildUser(id int) User {
	var user User
	query := fmt.Sprintf("SELECT * FROM users WHERE id = %d", id)
	row := db.QueryRow(query)

	// map the columns to struct fields
	err := row.Scan(&user.Id, &user.Name, &user.Email)

	if err != nil {
		panic(err.Error())
	}

	return user
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeGet)
	router.HandleFunc("/user", createUser).Methods("POST")
	router.HandleFunc("/email", sendEmail).Methods("POST")
	log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), router))
}

// returns some dumb JSON when the base URL is hit
func homeGet(writer http.ResponseWriter, req *http.Request) {
	// build response
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	resp := make(map[string] string)
	resp["message"] = "This does nothing"
	jsonResp, err := json.Marshal(resp)

	if err != nil {
		log.Fatal("Could not make JSON in homeGet")
	}

	// write the JSON response
	writer.Write(jsonResp)
}

// creates a user in the DB based on JSON body of the request
func createUser(writer http.ResponseWriter, req *http.Request) {
	reqBody, _ := ioutil.ReadAll(req.Body)
	var user User

	// take the JSON and map to the user struct
	json.Unmarshal(reqBody, &user)

	// insert in to the DB
	id := insertDbRow(user.Name, user.Email)
	user.Id = int(id)
	
	writer.WriteHeader(http.StatusCreated)
	writer.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(user)

	if err != nil {
		log.Fatal("Could not make JSON in createUser")
	}

	// send the response
	writer.Write(jsonResp)
}

func sendEmail(writer http.ResponseWriter, req *http.Request) {
	reqBody, _ := ioutil.ReadAll(req.Body)

	var emailRequest EmailReq

	// map JSON to request ID
	json.Unmarshal(reqBody, &emailRequest)

	// get user from the database
	user := queryDbAndBuildUser(emailRequest.Id)

	if !buildandSendEmail(user, emailRequest.Game) {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusCreated)
	}

}

func buildandSendEmail(user User, game string) bool {
	from := mail.NewEmail("Patch Poro", "junkmand@oregonstate.edu")
	to := mail.NewEmail(user.Name, user.Email)
	subject := fmt.Sprintf("Patch Poro - %s has an update!", game)
	plainText := fmt.Sprintf("Hi %s!\nWe found a new update for %s - log in to check it out. https://patch-poro.herokuapp.com", user.Name, game)
	client := sendgrid.NewSendClient(os.Getenv("SendGridKey"))
	message := mail.NewSingleEmail(from, subject, to, plainText, "")

	_, err := client.Send(message)

	if err != nil {
		log.Fatal(err.Error())
		return false
	}

	return true
}
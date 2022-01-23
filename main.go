package main

import (
	"fmt"
	"log"
	"database/sql"
	"os"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

// define a user
type User struct {
	id int
	name string
	email string
}

// runs automagically when the file is built and rain
func main() {
	config := buildDbConfig()
	connectToDb(config)
}

// returns env variable for local testing (will be commented out for deploy since
// Heroku env variables will be used)
func goDotEnvVariable(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal(err)
	}

	return os.Getenv(key)
}

// builds DB config (will be commented out for deploy)
// same function using Heroku config vars will be built
func buildDbConfig() mysql.Config {
	return mysql.Config{
		User: goDotEnvVariable("User"),
		Passwd: goDotEnvVariable("Password"),
		Net: "tcp",
		Addr: goDotEnvVariable("Address"),
		DBName: goDotEnvVariable("DB"),
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
	err := row.Scan(&user.id, &user.name, &user.email)

	if err != nil {
		panic(err.Error())
	}

	return user
} 
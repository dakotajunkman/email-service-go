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

// define a game
type Game struct {
	id int
	name string
	year int
	month int
	day int
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

// updates date in DB based on game name
// could use game id as well
func updateDbRow(name string, year int, month int, day int) int64 {
	command := fmt.Sprintf("UPDATE games SET year = %d, month = %d, day = %d WHERE name = '%s';", year, month, day, name)
	res, err := db.Exec(command)

	if err != nil {
		log.Fatal(err)
	}

	rows, err := res.RowsAffected()

	if err != nil {
		panic(err.Error())
	}

	// not needed but go requires all variables to be used
	return rows
}

func queryDbAndBuildGame(name string) Game {
	var game Game
	query := fmt.Sprintf("SELECT * FROM games WHERE name = '%s'", name)
	row := db.QueryRow(query)

	// map the columns to struct fields
	err := row.Scan(&game.id, &game.name, &game.year, &game.month, &game.day)

	if err != nil {
		panic(err.Error())
	}

	return game
} 
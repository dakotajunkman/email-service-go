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

func main() {
	config := connectToDb()
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

func goDotEnvVariable(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal(err)
	}

	return os.Getenv(key)
}
func connectToDb() mysql.Config {
	return mysql.Config{
		User: goDotEnvVariable("User"),
		Passwd: goDotEnvVariable("Password"),
		Net: "tcp",
		Addr: goDotEnvVariable("Address"),
		DBName: goDotEnvVariable("DB"),
		AllowNativePasswords: true,
	}
}

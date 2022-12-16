package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"strconv"
	"time"

	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/url"
)

func main() {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "mysql"
	}

	// mysql
	mysqlName := os.Getenv("MYSQL_NAME")
	if mysqlName == "" {
		mysqlName = "test"
	}
	mysqlUser := os.Getenv("MYSQL_USER")
	if mysqlUser == "" {
		mysqlUser = "root"
	}
	mysqlPass := os.Getenv("MYSQL_PASS")
	if mysqlPass == "" {
		mysqlPass = ""
	}
	mysqlHost := os.Getenv("MYSQL_HOST")
	if mysqlHost == "" && dbType == "mysql" {
		fmt.Fprintf(os.Stderr, "\"MYSQL_HOST\" not set, but \"DB_TYPE\" is set \"mysql\"")
		os.Exit(1)
	}
	mysqlPort := os.Getenv("MYSQL_PORT")
	if mysqlPort == "" {
		mysqlPort = "3306"
	}

	// mongodb
	mongoUri := os.Getenv("MONGODB_URI")
	if mongoUri == "" && dbType == "mongodb" {
		fmt.Fprintf(os.Stderr, "\"MONGODB_URI\" not set, but \"DB_TYPE\" is set \"mongodb\"")
		os.Exit(1)
	}

	triesStr := os.Getenv("TRIES")
	tries, err := strconv.Atoi(triesStr)
	if err != nil {
		tries = 10
	}

	for i := 1; i < tries; i += 1 {
		sleepS := 3*i + 1
		sleep := time.Duration(sleepS) * time.Second

		if dbType == "mysql" {
			db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mysqlUser, mysqlPass, mysqlHost, mysqlPort, mysqlName))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Try (%d/%d) sleep %d seconds error connect to '%s@password@tcp(%s:%s)/%s': %v\n", i, tries, sleepS, mysqlUser, mysqlHost, mysqlPort, mysqlName, err)
				time.Sleep(sleep)
				continue
			}
			defer db.Close()

			_, err = GetSQLTables(db)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Try (%d/%d) sleep %d seconds error: %v\n", i, tries, sleepS, err)
				time.Sleep(sleep)
				continue
			}
			fmt.Println("Connect success")
			break
		}

		if dbType == "mongodb" {
			url, err := url.Parse(mongoUri)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: cannot get db from uri: %v\n", err)
				os.Exit(1)
			}

			dbName := url.Path
			if dbName[0] == '/' {
				dbName = url.Path[1:]
			}

			client, err := mongo.NewClient(options.Client().ApplyURI(mongoUri))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Try (%d/%d) sleep %d seconds error mongodb connect to '%s': %v\n", i, tries, sleepS, url.Host)
				time.Sleep(sleep)
				continue
			}

			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			err = client.Connect(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error cannot create context %v\n", err)
				os.Exit(1)
			}

			_, err = client.Database(dbName).ListCollectionNames(ctx, bson.D{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Try (%d/%d) sleep %d seconds error list collections:%v\n", i, tries, sleepS, err)
				time.Sleep(sleep)
				continue
			}

			fmt.Println("Connect success")
			break
		}

		if i == tries-1 {
			fmt.Fprintf(os.Stderr, "Connection attempts have failed")
			os.Exit(2)
		}
	}

}

func GetSQLTables(db *sql.DB) ([]string, error) {
	errorFuncName := "Func GetSQLTables() error"
	query := "SHOW TABLES"
	tableRows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("%s: query: '%s': %v", errorFuncName, query, err)
	}
	defer tableRows.Close()

	var tables []string
	for tableRows.Next() {
		var table string
		err = tableRows.Scan(&table)
		if err != nil {
			return nil, fmt.Errorf("%s: for query '%s', cannot read table. Error: %v", errorFuncName, query, err)
		}
		tables = append(tables, table)
	}
	return tables, nil
}

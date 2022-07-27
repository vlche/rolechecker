package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v7"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// vars
var (
	flagService  = ""
	flagURL      = ""
	flagUser     = ""
	flagPassword = ""
)

func init() {
	flag.StringVar(&flagService, "service", "mongo", "service type (mongo, redis, postgresql etc), e.g. service=mongo")
	flag.StringVar(&flagURL, "url", "mongodb.local", "url for the service, e.g. url=mongodb.local")
	flag.StringVar(&flagUser, "user", "root", "user for the service, e.g. user=root")
	flag.StringVar(&flagPassword, "password", "example", "password for the service, e.g. password=example")
}

func loadEnv() {
	if envService := os.Getenv("CHECK_SERVICE"); envService != "" {
		flagService = envService
	}
	if envURL := os.Getenv("CHECK_URL"); envURL != "" {
		flagURL = envURL
	}
	if envUser := os.Getenv("CHECK_USER"); envUser != "" {
		flagUser = envUser
	}
	if envPassword := os.Getenv("CHECK_PASSWORD"); envPassword != "" {
		flagPassword = envPassword
	}
	return
}

func main() {
	flag.Parse()
	loadEnv()
	switch flagService {
	case "mongo":
		roleCheckMongo()
	case "redis":
		roleCheckRedis()
	case "postgres":
		roleCheckPostgres()
	}
}

func roleCheckMongo() {
	type mongoHelloStruct struct {
		SetName           string
		IsWritablePrimary bool
		Secondary         bool
		Primary           string
		Me                string
	}
	credential := options.Credential{
		AuthSource: "admin",
		Username:   flagUser,
		Password:   flagPassword,
	}
	clientOpts := options.Client().
		ApplyURI("mongodb://" + flagURL).
		SetAuth(credential).
		SetDirect(true)
	client, err := mongo.Connect(context.TODO(), clientOpts)

	if err != nil {
		panic(err)
	}
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}
	admin := client.Database(credential.AuthSource)
	var mongoHello mongoHelloStruct
	command := bson.M{"hello": 1}
	err = admin.RunCommand(context.TODO(), command).Decode(&mongoHello)
	if err != nil {
		log.Fatal(err)
	}
	if mongoHello.IsWritablePrimary {
		fmt.Println("Master")
		os.Exit(0)
	} else if mongoHello.Secondary {
		fmt.Println("Slave")
		os.Exit(1)
	}
	fmt.Println("Unknown")
	os.Exit(2)

	// log.Println(mongoHello.Me)
}

func roleCheckRedis() {
	const (
		redisPort = 6379
	)
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", flagURL, redisPort),
		Password: flagPassword,
	})
	rawstatus, err := client.Do("role").Result()
	if err != nil {
		currentStatus := "Unavailable"
		fmt.Printf("%s\n", currentStatus)
		os.Exit(2)
	}
	// Parse Redis-server status response
	// Response samples:
	// [master 0 []]   Master without Slave
	// [slave 127.0.0.1 6379 connected 0]  Slave of a Master Connected and synced
	// [slave 127.0.0.1 6379 connect 0]  Slave of a Master trying to connect
	// [slave 127.0.0.1 6379 sync 0]  Slave of a Master syncing
	status := rawstatus.([]interface{})
	currentRole := status[0]
	currentStatus := "Unknown"
	if currentRole == "master" {
		currentStatus = "Master"
		fmt.Printf("%s\n", currentStatus)
		os.Exit(0)
	} else if currentRole == "slave" {
		if status[3] == "connected" {
			currentStatus = "Slave"
			fmt.Printf("%s\n", currentStatus)
			os.Exit(1)
		} else {
			currentStatus = "Unhealthy state: " + status[3].(string)
		}
	}
	fmt.Printf("%s\n", currentStatus)
	os.Exit(2)
}

func roleCheckPostgres() {
	const (
		dbname = "postgres"
		port   = 5432
	)
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", flagURL, port, flagUser, flagPassword, dbname)
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		panic(err)
	}

	// close database
	defer db.Close()

	// check db
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	rows, err := db.Query(`select application_name from pg_stat_replication`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var application_name string

		err = rows.Scan(&application_name)
		if err != nil {
			panic(err)
		}

		fmt.Println("Master")
		os.Exit(0)
	}

	rows, err = db.Query(`select sender_host from pg_stat_wal_receiver`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var sender_host string

		err = rows.Scan(&sender_host)
		if err != nil {
			panic(err)
		}

		fmt.Println("Slave")
		fmt.Println(sender_host)
		os.Exit(1)
	}
	fmt.Println("Standalone Master")
	os.Exit(0)
}

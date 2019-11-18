package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"

	"net/http"
)

type Trainer struct {
	Name string
	Age  int
	City string
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome home!")
}

func main() {
	//router := mux.NewRouter().StrictSlash(true)
	//router.HandleFunc("/", homeLink)
	//log.Fatal(http.ListenAndServe(":8000", router))

	// Set up a context required by mongo.Connect

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //To close the connection at the end
	defer cancel()                                                           //We need to set up a client first

	//It takes the URI of your database
	client, error := mongo.NewClient(options.Client().ApplyURI("your_database_uri"))
	if error != nil {
		log.Fatal(error)
	}
	//Call the connect function of client
	error = client.Connect(ctx) //Checking the connection
	error = client.Ping(context.TODO(), nil)
	fmt.Println("Database connected")
}

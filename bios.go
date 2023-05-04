/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * bios.go
 */

package hatchet

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"

	fake "github.com/brianvoe/gofakeit/v6"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	BioDBName   = "hatchet"
	BioCollName = "bios"
)

type Bio struct {
	Age         int      `bson:"age" json:"age"`
	Company     string   `bson:"company" json:"company"`
	CreditCards []string `bson:"credit_cards" json:"credit_cards"`
	Emails      []string `bson:"emails" json:"emails"`
	FirstName   string   `bson:"first_name" json:"first_name"`
	Intro       string   `bson:"intro" json:"intro"`
	LastName    string   `bson:"last_name" json:"last_name"`
	Phones      []string `bson:"phones" json:"phones"`
	Title       string   `bson:"title" json:"title"`
	SSN         string   `bson:"ssn" json:"ssn"`
	State       string   `bson:"state" json:"state"`
	URL         string   `bson:"url" json:"url"`
}

func generateBio() Bio {
	firstName := fake.FirstName()
	lastName := fake.LastName()
	company := fake.Company()
	email := strings.ToLower(firstName + "." + lastName + "@" + strings.ToLower(company) + ".com")
	phoneNo := "+1 " + fake.Phone()
	ssn := fmt.Sprintf("%d-%d-%d", rand.Intn(900)+100, rand.Intn(90)+10, rand.Intn(9000)+1000)
	cardNo := fmt.Sprintf("%d-%d-%d-%d", rand.Intn(9000)+1000, rand.Intn(9000)+1000, rand.Intn(9000)+1000, rand.Intn(9000)+1000)
	state := fake.State()
	title := fake.JobTitle()
	min := 22
	max := 55
	age := rand.Intn(max-min+1) + min
	url := fmt.Sprintf("https://%s:%d", fake.DomainName(), rand.Intn(65535)+1024)
	intro := fmt.Sprintf("Meet %s %s, a %d-year-old from %s. %s works as a %s at %s.", firstName, lastName, age, state, firstName, title, company)

	// Create a Bio struct with the generated values
	return Bio{
		Age:         age,
		Company:     company,
		CreditCards: []string{cardNo},
		Emails:      []string{email},
		FirstName:   firstName,
		Intro:       intro,
		Title:       title,
		LastName:    lastName,
		URL:         url,
		Phones:      []string{phoneNo},
		SSN:         ssn,
		State:       state,
	}
}

func insertBios(client *mongo.Client, wg *sync.WaitGroup, size int) error {
	defer wg.Done()
	if size < 1000 {
		size = 1000
	}
	fmt.Println("Insert", size, "bios")
	collection := client.Database(BioDBName).Collection(BioCollName)

	// Insert the documents in batches of 1000
	var err error
	for i := 0; i < size; i += 1000 {
		j := i + 1000
		if j > size {
			j = size
		}

		// Generate a slice of Bio documents for this batch
		docs := make([]interface{}, j-i)
		for k := 0; k < j-i; k++ {
			docs[k] = generateBio()
		}

		// Insert the documents in this batch
		_, err = collection.InsertMany(context.Background(), docs)
		if err != nil {
			return err
		}
	}

	return nil
}

func InsertBiosIntoMongoDB(uri string, numDocuments int) error {
	log.Println(uri, numDocuments)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	err = client.Connect(context.Background())
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	// Divide the work among N goroutines
	numThreads := 4
	chunkSize := numDocuments / numThreads
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go insertBios(client, &wg, chunkSize)
	}

	wg.Wait()
	err = client.Disconnect(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func SimulateTests(test string, url string) error {
	if test == "read" {
		return SimulateReads(url)
	} else if test == "write" {
		return SimulateWrites(url)
	}
	return errors.New("unrecognized test type")
}

func SimulateReads(url string) error {
	pipeline := mongo.Pipeline{}
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(url))
	if err != nil {
		fmt.Println("Failed to connect to MongoDB:", err)
		return err
	}
	defer client.Disconnect(context.Background())
	collection := client.Database(BioDBName).Collection(BioCollName)
	wg := sync.WaitGroup{}
	numThreads := runtime.NumCPU() * 4
	if numThreads > 64 {
		numThreads = 64
	}

	// Spawn n threads to read the aggregation pipeline
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(seq int) {
			defer wg.Done()
			start := time.Now()
			fmt.Println(start, seq)
			project := bson.D{{Key: "$project", Value: bson.D{{Key: "intro", Value: 1}, {Key: "emails", Value: 1}, {Key: "ssn", Value: 1}, {Key: "phones", Value: 1}, {Key: "credit_cards", Value: 1}}}}
			match := bson.D{{Key: "$match", Value: bson.D{{Key: "state", Value: fake.State()}, {Key: "last_name", Value: fake.LastName()}, {Key: "first_name", Value: fake.FirstName()}}}}
			pipeline = mongo.Pipeline{match, project}
			for {
				if seq < 2 {
					ssn := fmt.Sprintf("%d-%d-%d", rand.Intn(900)+100, rand.Intn(90)+10, rand.Intn(9000)+1000)
					cardNo := fmt.Sprintf("%d-%d-%d-%d", rand.Intn(9000)+1000, rand.Intn(9000)+1000, rand.Intn(9000)+1000, rand.Intn(9000)+1000)
					match = bson.D{{Key: "$match", Value: bson.D{{Key: "ssn", Value: ssn}, {Key: "emails", Value: fake.Email()}, {Key: "credit_cards", Value: cardNo}}}}
					pipeline = mongo.Pipeline{match, project}
				} else if seq < 4 {
					match = bson.D{{Key: "$match", Value: bson.D{{Key: "last_name", Value: fake.LastName()}, {Key: "first_name", Value: fake.FirstName()}}}}
					pipeline = mongo.Pipeline{match, project}
				} else if seq < 6 {
					match = bson.D{{Key: "$match", Value: bson.D{{Key: "state", Value: fake.State()}, {Key: "last_name", Value: fake.LastName()}}}}
					pipeline = mongo.Pipeline{match, project}
				} else if seq < 8 {
					match = bson.D{{Key: "$match", Value: bson.D{{Key: "state", Value: fake.State()}}}}
					pipeline = mongo.Pipeline{match, project}
				}
				// Run the aggregation pipeline
				cursor, err := collection.Aggregate(context.Background(), pipeline)
				if err != nil {
					log.Fatal("Failed to run aggregation pipeline:", err)
					return
				}

				// Process the aggregation result
				defer cursor.Close(context.Background())
				time.Sleep(20 * time.Millisecond)
				for cursor.Next(context.Background()) {
					// Process the aggregation result here
				}
				if err := cursor.Err(); err != nil {
					log.Fatal("Failed to read aggregation result:", err)
					return
				}

				executionTime := time.Since(start)
				if executionTime < 2*time.Minute {
					time.Sleep(20 * time.Millisecond)
					continue
				} else {
					break
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("completed SimulateReads")
	return nil
}

func SimulateWrites(url string) error {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(url))
	if err != nil {
		fmt.Println("Failed to connect to MongoDB:", err)
		return err
	}
	defer client.Disconnect(context.Background())
	collection := client.Database(BioDBName).Collection(BioCollName)
	wg := sync.WaitGroup{}
	numThreads := runtime.NumCPU() * 4
	if numThreads > 64 {
		numThreads = 64
	}

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(seq int) {
			defer wg.Done()
			start := time.Now()
			fmt.Println(seq, start)
			updated := bson.D{{Key: "$inc", Value: bson.D{{Key: "updated", Value: 1}}}, {Key: "$set", Value: bson.D{{Key: "source_ip", Value: fake.IPv4Address()}}}}
			match := bson.D{{Key: "state", Value: fake.State()}, {Key: "last_name", Value: fake.LastName()}, {Key: "first_name", Value: fake.FirstName()}}
			for {
				if seq < 2 {
					ssn := fmt.Sprintf("%d-%d-%d", rand.Intn(900)+100, rand.Intn(90)+10, rand.Intn(9000)+1000)
					cardNo := fmt.Sprintf("%d-%d-%d-%d", rand.Intn(9000)+1000, rand.Intn(9000)+1000, rand.Intn(9000)+1000, rand.Intn(9000)+1000)
					match = bson.D{{Key: "ssn", Value: ssn}, {Key: "emails", Value: fake.Email()}, {Key: "credit_cards", Value: cardNo}}
				} else if seq < 4 {
					match = bson.D{{Key: "last_name", Value: fake.LastName()}, {Key: "first_name", Value: fake.FirstName()}}
				} else if seq < 6 {
					match = bson.D{{Key: "state", Value: fake.State()}, {Key: "last_name", Value: fake.LastName()}}
				} else if seq < 8 {
					match = bson.D{{Key: "state", Value: fake.State()}}
				}
				// Run updateMany
				_, err := collection.UpdateMany(context.Background(), match, updated)
				if err != nil {
					log.Fatal("Failed to run update:", err)
				}

				executionTime := time.Since(start)
				if executionTime < 2*time.Minute {
					time.Sleep(20 * time.Millisecond)
					continue
				} else {
					break
				}
			}
		}(i)
	}
	wg.Wait()
	fmt.Println("completed SimulateWrites")
	return nil
}

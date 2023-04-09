/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * mongo.go
 */

package hatchet

import (
	"context"
	"log"
	"net/url"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MAX_DOC_SIZE = 16 * (1024 * 1024)
	BATCH_SIZE   = 1000
)

type MongoDB struct {
	db          *mongo.Database
	hatchetName string
	url         string
	verbose     bool

	clients []interface{}
	drivers []interface{}
	logs    []interface{}
}

func NewMongoDB(connstr string, hatchetName string) (*MongoDB, error) {
	var err error
	mongodb := &MongoDB{url: connstr, hatchetName: hatchetName}
	clientOptions := options.Client().ApplyURI(connstr)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return mongodb, err
	}
	u, err := url.Parse(connstr)
	dbName := u.Path[1:]
	if dbName == "" || dbName == "admin" {
		dbName = "logdb"
	}
	mongodb.db = client.Database(dbName)
	return mongodb, err
}

func (ptr *MongoDB) GetVerbose() bool {
	return ptr.verbose
}

func (ptr *MongoDB) SetVerbose(b bool) {
	ptr.verbose = b
}

func (ptr *MongoDB) Begin() error {
	var err error
	log.Println("creating hatchet", ptr.hatchetName)
	collName := ptr.hatchetName
	for _, keys := range []bson.D{
		{{Key: "component", Value: 1}},
		{{Key: "context", Value: 1}, {Key: "date", Value: 1}},
		{{Key: "severity", Value: 1}},
		{{Key: "op", Value: 1}, {Key: "ns", Value: 1}, {Key: "filter", Value: 1}},
	} {
		index := mongo.IndexModel{
			Keys:    keys,
			Options: options.Index().SetUnique(false),
		}
		_, err = ptr.db.Collection(collName).Indexes().CreateOne(context.Background(), index)
		if err != nil {
			return err
		}
	}

	collName = ptr.hatchetName + "_clients"
	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "context", Value: 1}, {Key: "ip", Value: 1}},
		Options: options.Index().SetUnique(false),
	}
	_, err = ptr.db.Collection(collName).Indexes().CreateOne(context.Background(), index)
	if err != nil {
		return err
	}

	return err
}

func (ptr *MongoDB) Commit() error {
	if len(ptr.logs) > 0 {
		ptr.db.Collection(ptr.hatchetName).InsertMany(context.Background(), ptr.logs)
		ptr.logs = []interface{}{}
	}
	if len(ptr.clients) > 0 {
		ptr.db.Collection(ptr.hatchetName+"_clients").InsertMany(context.Background(), ptr.clients)
		ptr.clients = []interface{}{}
	}
	if len(ptr.drivers) > 0 {
		ptr.db.Collection(ptr.hatchetName+"_drivers").InsertMany(context.Background(), ptr.drivers)
		ptr.drivers = []interface{}{}
	}
	return nil
}

func (ptr *MongoDB) Close() error {
	var err error
	defer ptr.db.Client().Disconnect(context.Background())
	return err
}

// Drop drops all tables of a hatchet
func (ptr *MongoDB) Drop() error {
	var err error
	ptr.db.Collection(ptr.hatchetName + "_audit").Drop(context.Background())
	ptr.db.Collection(ptr.hatchetName + "_clients").Drop(context.Background())
	ptr.db.Collection(ptr.hatchetName + "_drivers").Drop(context.Background())
	ptr.db.Collection(ptr.hatchetName + "_ops").Drop(context.Background())
	ptr.db.Collection(ptr.hatchetName).Drop(context.Background())
	ptr.db.Collection("hatchet").DeleteOne(context.Background(), bson.M{"name": ptr.hatchetName})
	return err
}

func (ptr *MongoDB) InsertLog(index int, end string, doc *Logv2Info, stat *OpStat) error {
	var err error
	data := bson.M{
		"_id": index, "date": end, "severity": doc.Severity, "component": doc.Component, "context": doc.Context,
		"msg": doc.Msg, "plan": doc.Attributes.PlanSummary, "type": doc.Attr.Map()["type"], "ns": doc.Attributes.NS, "message": doc.Message,
		"op": stat.Op, "filter": stat.QueryPattern, "_index": stat.Index, "milli": doc.Attributes.Milli, "reslen": doc.Attributes.Reslen}
	ptr.logs = append(ptr.logs, data)
	if len(ptr.logs) > BATCH_SIZE {
		collName := ptr.hatchetName
		_, err = ptr.db.Collection(collName).InsertMany(context.Background(), ptr.logs)
		ptr.logs = []interface{}{}
	}
	return err
}

func (ptr *MongoDB) InsertClientConn(index int, doc *Logv2Info) error {
	var err error
	client := doc.Client
	data := bson.M{
		"_id": index, "ip": client.IP, "port": client.Port, "conns": client.Conns, "accepted": client.Accepted,
		"ended": client.Ended, "context": doc.Context}
	ptr.clients = append(ptr.clients, data)
	if len(ptr.clients) > BATCH_SIZE {
		collName := ptr.hatchetName + "_clients"
		_, err = ptr.db.Collection(collName).InsertMany(context.Background(), ptr.clients)
		ptr.clients = []interface{}{}
	}
	return err
}

func (ptr *MongoDB) InsertDriver(index int, doc *Logv2Info) error {
	var err error
	client := doc.Client
	data := bson.M{
		"_id": index, "ip": client.IP, "driver": client.Driver, "version": client.Version}
	ptr.drivers = append(ptr.drivers, data)
	if len(ptr.drivers) > BATCH_SIZE {
		collName := ptr.hatchetName + "_drivers"
		_, err = ptr.db.Collection(collName).InsertMany(context.Background(), ptr.drivers)
		ptr.drivers = []interface{}{}
	}
	return err
}

func (ptr *MongoDB) UpdateHatchetInfo(info HatchetInfo) error {
	var err error
	filter := bson.M{"name": ptr.hatchetName}
	update := bson.M{"$set": bson.M{"version": info.Version, "module": info.Module, "arch": info.Arch, "os": info.OS, "start": info.Start, "end": info.End}}
	upsertOptions := options.Update().SetUpsert(true)
	_, err = ptr.db.Collection("hatchet").UpdateOne(context.Background(), filter, update, upsertOptions)
	return err
}

func (ptr *MongoDB) CreateMetaData() error {
	var err error
	log.Printf("insert ops into %v_ops\n", ptr.hatchetName)
	pipeline := []bson.M{
		{"$match": bson.M{
			"op": bson.M{
				"$nin": []interface{}{nil, ""},
			},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"op":     "$op",
				"ns":     "$ns",
				"filter": "$filter",
				"_index": "$_index",
			},
			"count":    bson.M{"$sum": 1},
			"avg_ms":   bson.M{"$avg": "$milli"},
			"max_ms":   bson.M{"$max": "$milli"},
			"total_ms": bson.M{"$sum": "$milli"},
			"reslen":   bson.M{"$sum": "$reslen"},
		}},
		{"$project": bson.M{
			"_id":      0,
			"op":       "$_id.op",
			"count":    1,
			"avg_ms":   bson.M{"$round": []interface{}{"$avg_ms", 0}},
			"max_ms":   1,
			"total_ms": 1,
			"ns":       "$_id.ns",
			"_index":   "$_id._index",
			"reslen":   1,
			"filter":   "$_id.filter",
		}},
		{"$merge": bson.M{
			"into": ptr.hatchetName + "_ops",
		}},
	}
	if _, err = ptr.db.Collection(ptr.hatchetName).Aggregate(context.Background(), pipeline); err != nil {
		return err
	}

	log.Printf("insert [exception] into %v_audit\n", ptr.hatchetName)
	pipeline = []bson.M{
		{"$match": bson.M{
			"severity": bson.M{
				"$in": []interface{}{"W", "E", "F"},
			},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"severity": "$severity",
			},
			"count": bson.M{"$sum": 1},
		}},
		{"$project": bson.M{
			"_id":   0,
			"type":  "exception",
			"name":  "$_id.severity",
			"value": "$count",
		}},
		{"$merge": bson.M{
			"into": ptr.hatchetName + "_audit",
		}},
	}
	if _, err = ptr.db.Collection(ptr.hatchetName).Aggregate(context.Background(), pipeline); err != nil {
		return err
	}

	log.Printf("insert [failed] into %v_audit\n", ptr.hatchetName)
	pipeline = []bson.M{
		{"$match": bson.M{
			"message": bson.M{
				"$regex": `(\w\sfailed\s)`,
			},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"$substr": bson.A{"$message", 0, bson.M{
					"$add": bson.A{
						bson.M{
							"$indexOfBytes": bson.A{"$message", "failed"},
						},
						6,
					},
				}},
			},
			"count": bson.M{"$sum": 1},
		}},
		{"$project": bson.M{
			"_id":   0,
			"type":  "failed",
			"name":  "$_id",
			"value": "$count",
		}},
		{"$merge": bson.M{
			"into": ptr.hatchetName + "_audit",
		}},
	}
	if _, err = ptr.db.Collection(ptr.hatchetName).Aggregate(context.Background(), pipeline); err != nil {
		return err
	}

	log.Printf("insert [op] into %v_audit\n", ptr.hatchetName)
	pipeline = []bson.M{
		{"$match": bson.M{
			"op": bson.M{
				"$nin": []interface{}{nil, ""},
			},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"op": "$op",
			},
			"count": bson.M{"$sum": 1},
		}},
		{"$project": bson.M{
			"_id":   0,
			"type":  "op",
			"name":  "$_id.op",
			"value": "$count",
		}},
		{"$merge": bson.M{
			"into": ptr.hatchetName + "_audit",
		}},
	}
	if _, err = ptr.db.Collection(ptr.hatchetName).Aggregate(context.Background(), pipeline); err != nil {
		return err
	}

	log.Printf("insert [ip] into %v_audit\n", ptr.hatchetName)
	pipeline = []bson.M{
		{"$group": bson.M{
			"_id": bson.M{
				"ip": "$ip",
			},
			"open": bson.M{"$sum": "$accepted"},
		}},
		{"$project": bson.M{
			"_id":   0,
			"type":  "ip",
			"name":  "$_id.ip",
			"value": "$open",
		}},
		{"$merge": bson.M{
			"into": ptr.hatchetName + "_audit",
		}},
	}
	if _, err = ptr.db.Collection(ptr.hatchetName+"_clients").Aggregate(context.Background(), pipeline); err != nil {
		return err
	}

	log.Printf("insert [reslen-ip] into %v_audit\n", ptr.hatchetName)
	pipeline = []bson.M{
		{"$match": bson.M{
			"op": bson.M{
				"$nin": []interface{}{nil, ""},
			},
			"reslen": bson.M{"$gt": 0},
		}},
		{"$lookup": bson.M{
			"from": ptr.hatchetName + "_clients",
			"let":  bson.M{"context": "$context"},
			"pipeline": []bson.M{
				{"$match": bson.M{
					"$expr": bson.M{"$eq": []interface{}{"$context", "$$context"}},
				}},
				{"$project": bson.M{
					"_id": 0,
					"ip":  1,
				}},
			},
			"as": "clients",
		}},
		{"$unwind": bson.M{"path": "$clients"}},
		{"$group": bson.M{
			"_id":    "$clients.ip",
			"reslen": bson.M{"$sum": "$reslen"},
		}},
		{"$project": bson.M{
			"_id":   0,
			"type":  "reslen-ip",
			"name":  "$_id",
			"value": "$reslen",
		}},
		{"$merge": bson.M{
			"into": ptr.hatchetName + "_audit",
		}},
	}
	if _, err = ptr.db.Collection(ptr.hatchetName).Aggregate(context.Background(), pipeline); err != nil {
		return err
	}

	log.Printf("insert [ns] into %v_audit\n", ptr.hatchetName)
	pipeline = []bson.M{
		{"$match": bson.M{
			"op": bson.M{
				"$nin": []interface{}{nil, ""},
			},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"ns": "$ns",
			},
			"count": bson.M{"$sum": 1},
		}},
		{"$project": bson.M{
			"_id":   0,
			"type":  "ns",
			"name":  "$_id.ns",
			"value": "$count",
		}},
		{"$merge": bson.M{
			"into": ptr.hatchetName + "_audit",
		}},
	}
	if _, err = ptr.db.Collection(ptr.hatchetName).Aggregate(context.Background(), pipeline); err != nil {
		return err
	}

	log.Printf("insert [reslen-ns] into %v_audit\n", ptr.hatchetName)
	pipeline = []bson.M{
		{"$match": bson.M{
			"ns": bson.M{
				"$nin": []interface{}{nil, ""},
			},
			"reslen": bson.M{"$gt": 0},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"ns": "$ns",
			},
			"reslen": bson.M{"$sum": "$reslen"},
		}},
		{"$project": bson.M{
			"_id":   0,
			"type":  "reslen-ns",
			"name":  "$_id.ns",
			"value": "$reslen",
		}},
		{"$merge": bson.M{
			"into": ptr.hatchetName + "_audit",
		}},
	}
	if _, err = ptr.db.Collection(ptr.hatchetName).Aggregate(context.Background(), pipeline); err != nil {
		return err
	}
	return nil
}

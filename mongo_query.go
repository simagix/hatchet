/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * mongo_query.go
 */

package hatchet

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (ptr *MongoDB) GetAverageOpTime(op string, duration string) ([]OpCount, error) {
	var docs []OpCount
	var substr bson.M
	ctx := context.Background()
	opcond := bson.M{"op": bson.M{"$ne": ""}}
	if op != "" {
		opcond = bson.M{"op": op}
	}
	if duration != "" {
		toks := strings.Split(duration, ",")
		substr = GetMongoDateSubString(toks[0], toks[1])
		opcond["$and"] = []bson.M{
			{"date": bson.M{"$gte": toks[0]}},
			{"date": bson.M{"$lt": toks[1]}},
		}
	} else {
		info := ptr.GetHatchetInfo()
		substr = GetMongoDateSubString(info.Start, info.End)
	}
	group := bson.M{
		"_id": bson.M{
			"date":   substr,
			"op":     "$op",
			"ns":     "$ns",
			"filter": "$filter",
		},
		"milli_avg": bson.M{"$avg": "$milli"},
		"count":     bson.M{"$sum": 1},
	}
	project := bson.M{
		"_id":    0,
		"date":   "$_id.date",
		"milli":  "$milli_avg",
		"count":  "$count",
		"op":     "$_id.op",
		"ns":     "$_id.ns",
		"filter": "$_id.filter",
	}
	opts := options.Aggregate().SetAllowDiskUse(true)
	cursor, err := ptr.db.Collection(ptr.hatchetName).Aggregate(ctx, []bson.M{
		{"$match": opcond},
		{"$group": group},
		{"$project": project},
		{"$sort": bson.M{"date": 1}},
	}, opts)
	if err != nil {
		return docs, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var doc OpCount
		if err := cursor.Decode(&doc); err != nil {
			return docs, err
		}
		if len(doc.Date) < 19 {
			full := "2023-09-23T23:59:59"
			doc.Date += full[len(doc.Date):]
		}
		docs = append(docs, doc)
	}
	if err := cursor.Err(); err != nil {
		return docs, err
	}
	return docs, nil
}

func (ptr *MongoDB) GetHatchetInfo() HatchetInfo {
	ctx := context.Background()
	var info HatchetInfo
	db := ptr.db

	// Get hatchet information from "hatchet" collection
	filter := bson.M{"name": ptr.hatchetName}
	cur, err := db.Collection("hatchet").Find(ctx, filter, options.Find())
	if err != nil {
		return info
	}
	if cur.Next(ctx) {
		if err = cur.Decode(&info); err != nil {
			return info
		}
	}

	// Get provider and region information from logs
	filter = bson.M{"component": "CONTROL", "message": bson.M{"$regex": ".*provider:.*region:.*"}}
	projection := bson.M{"message": 1}
	cur, err = db.Collection(ptr.hatchetName).Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return info
	}
	if cur.Next(ctx) {
		var doc bson.M
		if err = cur.Decode(&doc); err != nil {
			return info
		}
		message := doc["message"].(string)
		re := regexp.MustCompile(`.*(provider: "(\w+)", region: "(\w+)",).*`)
		matches := re.FindStringSubmatch(message)
		info.Provider = matches[2]
		info.Region = matches[3]
	}

	// Get driver information from "drivers" collection
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id": bson.M{"driver": "$driver", "version": "$version"},
		}},
		{"$project": bson.M{
			"_id":     0,
			"driver":  "$_id.driver",
			"version": "$_id.version",
		}},
	}
	if ptr.verbose {
		fmt.Println(gox.Stringify(pipeline))
	}
	cursor, err := ptr.db.Collection(ptr.hatchetName+"_drivers").Aggregate(ctx, pipeline)
	if err != nil {
		return info
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var driverVersion struct {
			Driver  string `bson:"driver"`
			Version string `bson:"version"`
		}
		if err := cursor.Decode(&driverVersion); err != nil {
			return info
		}
		info.Drivers = append(info.Drivers, map[string]string{driverVersion.Driver: driverVersion.Version})
	}
	if err := cursor.Err(); err != nil {
		return info
	}
	return info
}

func (ptr *MongoDB) GetHatchetNames() ([]string, error) {
	var err error
	ctx := context.Background()
	names := []string{}
	opts := options.Find()
	opts.SetProjection(bson.M{"name": 1})
	db := ptr.db
	cur, err := db.Collection("hatchet").Find(ctx, bson.D{{}}, opts)
	if err != nil {
		return names, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var doc bson.M
		if err = cur.Decode(&doc); err != nil {
			return names, err
		}
		names = append(names, doc["name"].(string))
	}
	return names, err
}

// GetAcceptedConnsCounts returns opened connection counts
func (ptr *MongoDB) GetAcceptedConnsCounts(duration string) ([]NameValue, error) {
	var err error
	ctx := context.Background()
	docs := []NameValue{}
	pipeline := []bson.M{
		{"$match": bson.M{"accepted": 1}},
		{"$group": bson.M{
			"_id":   "$ip",
			"total": bson.M{"$sum": "$accepted"},
		}},
		{"$project": bson.M{
			"_id":   0,
			"name":  "$_id",
			"value": "$total",
		}},
	}
	if duration != "" {
		toks := strings.Split(duration, ",")
		pipeline = []bson.M{
			{"$match": bson.M{"accepted": 1}},
			{"$lookup": bson.M{
				"from": ptr.hatchetName,
				"let":  bson.M{"id": "$_id"},
				"pipeline": []bson.M{
					{"$match": bson.M{
						"$expr": bson.M{
							"$and": []bson.M{
								{"$eq": []interface{}{"$_id", "$$id"}},
								{"$gte": []interface{}{"$date", toks[0]}},
								{"$lt": []interface{}{"$date", toks[1]}},
							}},
					}},
					{"$project": bson.M{
						"_id":  0,
						"date": 1,
					}},
				},
				"as": "clients",
			}},
			{"$unwind": "$clients"},
			{"$group": bson.M{
				"_id":   "$ip",
				"total": bson.M{"$sum": "$accepted"},
			}},
			{"$project": bson.M{
				"_id":   0,
				"name":  "$_id",
				"value": "$total",
			}},
		}
	}
	collection := ptr.db.Collection(ptr.hatchetName + "_clients")
	opts := options.Aggregate().SetAllowDiskUse(true)
	if ptr.verbose {
		fmt.Println(gox.Stringify(pipeline))
	}
	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return docs, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var doc NameValue
		if err = cursor.Decode(&doc); err != nil {
			return docs, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// GetConnectionStats returns stats data of accepted and ended
func (ptr *MongoDB) GetConnectionStats(chartType string, duration string) ([]RemoteClient, error) {
	var err error
	ctx := context.Background()
	docs := []RemoteClient{}
	collection := ptr.db.Collection(ptr.hatchetName + "_clients")
	var cursor *mongo.Cursor
	var pipeline []bson.M
	if chartType == "time" {
		var substr bson.M
		var sd, ed string
		if duration != "" {
			toks := strings.Split(duration, ",")
			substr = GetMongoDateSubString(toks[0], toks[1])
			sd = toks[0]
			ed = toks[1]
		} else {
			info := ptr.GetHatchetInfo()
			sd = info.Start
			ed = info.End
			substr = GetMongoDateSubString(info.Start, info.End)
		}
		pipeline = []bson.M{
			{"$lookup": bson.M{
				"from": ptr.hatchetName,
				"let":  bson.M{"id": "$_id"},
				"pipeline": []bson.M{
					{"$match": bson.M{
						"$expr": bson.M{
							"$and": []bson.M{
								{"$eq": []interface{}{"$_id", "$$id"}},
								{"$gte": []interface{}{"$date", sd}},
								{"$lt": []interface{}{"$date", ed}},
							}},
					}},
					{"$project": bson.M{
						"_id":  0,
						"date": 1,
					}},
				},
				"as": "clients",
			}},
			{"$unwind": "$clients"},
			{"$project": bson.M{
				"_id":   0,
				"date":  "$clients.date",
				"conns": 1,
				"ip":    1,
			}},
			{"$group": bson.M{
				"_id": bson.M{
					"date": "$date",
					"ip":   "$ip",
				},
				"accepted": bson.M{"$sum": "$conns"},
			}},
			{"$project": bson.M{
				"_id":      0,
				"date":     "$_id.date",
				"accepted": 1,
				"ended":    1,
			}},
			{"$group": bson.M{
				"_id":      substr,
				"accepted": bson.M{"$avg": "$accepted"},
			}},
			{"$project": bson.M{
				"_id":      0,
				"ip":       "$_id",
				"accepted": bson.M{"$toInt": "$accepted"},
				"ended":    bson.M{"$literal": 0},
			}},
			{"$sort": bson.M{"ip": 1}},
		}
	} else if chartType == "total" {
		pipeline = []bson.M{
			{"$group": bson.M{
				"_id":      "$ip",
				"accepted": bson.M{"$sum": "$accepted"},
				"ended":    bson.M{"$sum": "$ended"},
			}},
			{"$project": bson.M{
				"_id":      0,
				"ip":       "$_id",
				"accepted": 1,
				"ended":    1,
			}},
			{"$sort": bson.M{"accepted": -1}},
		}
		if duration != "" {
			toks := strings.Split(duration, ",")
			pipeline = []bson.M{
				{"$lookup": bson.M{
					"from": ptr.hatchetName,
					"let":  bson.M{"id": "$_id"},
					"pipeline": []bson.M{
						{"$match": bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{"$eq": []interface{}{"$_id", "$$id"}},
									{"$gte": []interface{}{"$date", toks[0]}},
									{"$lt": []interface{}{"$date", toks[1]}},
								}},
						}},
						{"$project": bson.M{
							"_id":  0,
							"date": 1,
						}},
					},
					"as": "clients",
				}},
				{"$unwind": "$clients"},
				{"$group": bson.M{
					"_id":      "$ip",
					"accepted": bson.M{"$sum": "$accepted"},
					"ended":    bson.M{"$sum": "$ended"},
				}},
				{"$project": bson.M{
					"_id":      0,
					"ip":       "$_id",
					"accepted": 1,
					"ended":    1,
				}},
				{"$sort": bson.M{"accepted": -1}},
			}
		}
	}
	if ptr.verbose {
		fmt.Println(gox.Stringify(pipeline))
	}
	opts := options.Aggregate().SetAllowDiskUse(true)
	cursor, err = collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return docs, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var doc RemoteClient
		if err = cursor.Decode(&doc); err != nil {
			return docs, err
		}

		if chartType == "time" && len(doc.IP) < 19 {
			full := "2023-09-23T23:59:59"
			doc.IP += full[len(doc.IP):]
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// GetOpsCounts returns opened connection counts
func (ptr *MongoDB) GetOpsCounts(duration string) ([]NameValue, error) {
	var err error
	docs := []NameValue{}
	opcond := bson.M{"op": bson.M{"$ne": ""}}
	if duration != "" {
		toks := strings.Split(duration, ",")
		opcond["$and"] = []bson.M{
			{"date": bson.M{"$gte": toks[0]}},
			{"date": bson.M{"$lt": toks[1]}},
		}
	}
	group := bson.M{
		"_id":   "$op",
		"count": bson.M{"$sum": 1},
	}
	project := bson.M{
		"_id":   0,
		"name":  "$_id",
		"value": "$count",
	}
	opts := options.Aggregate().SetAllowDiskUse(true)
	cursor, err := ptr.db.Collection(ptr.hatchetName).Aggregate(context.Background(), []bson.M{
		{"$match": opcond},
		{"$group": group},
		{"$project": project},
	}, opts)
	if err != nil {
		return docs, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var doc NameValue
		if err := cursor.Decode(&doc); err != nil {
			return docs, err
		}
		docs = append(docs, doc)
	}
	if err := cursor.Err(); err != nil {
		return docs, err
	}
	return docs, err
}

// GetReslenByIP returns total response length by ip
func (ptr *MongoDB) GetReslenByIP(ip string, duration string) ([]NameValue, error) {
	var err error
	ctx := context.Background()
	docs := []NameValue{}
	pipeline := []bson.M{
		{"$match": bson.M{
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
		{"$unwind": "$clients"},
		{"$group": bson.M{
			"_id":   "$clients.ip",
			"total": bson.M{"$sum": "$reslen"},
		}},
		{"$project": bson.M{
			"_id":   0,
			"name":  "$_id",
			"value": "$total",
		}},
		{"$sort": bson.M{"value": 1}},
	}
	if ip != "" {
		lookup := bson.M{
			"from": ptr.hatchetName + "_clients",
			"let":  bson.M{"context": "$context"},
			"pipeline": []bson.M{
				{"$match": bson.M{
					"ip":    ip,
					"$expr": bson.M{"$eq": []interface{}{"$context", "$$context"}},
				}},
				{"$project": bson.M{
					"_id": 0,
					"ip":  1,
				}},
			},
			"as": "clients",
		}
		pipeline[1]["$lookup"] = lookup
	}
	if duration != "" {
		match := pipeline[0]["$match"].(bson.M)
		toks := strings.Split(duration, ",")
		match["date"] = bson.M{"$gte": toks[0], "$lt": toks[1]}
		pipeline[0]["$match"] = match
	}
	collection := ptr.db.Collection(ptr.hatchetName)
	opts := options.Aggregate().SetAllowDiskUse(true)
	if ptr.verbose {
		fmt.Println(gox.Stringify(pipeline))
	}
	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return docs, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var doc NameValue
		if err = cursor.Decode(&doc); err != nil {
			return docs, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// GetReslenByNamespace returns total response length by ns
func (ptr *MongoDB) GetReslenByNamespace(ns string, duration string) ([]NameValue, error) {
	var err error
	ctx := context.Background()
	docs := []NameValue{}
	pipeline := []bson.M{
		{"$match": bson.M{
			"op":     bson.M{"$nin": []interface{}{nil, ""}},
			"reslen": bson.M{"$gt": 0},
		}},
		{"$group": bson.M{
			"_id":   "$ns",
			"total": bson.M{"$sum": "$reslen"},
		}},
		{"$project": bson.M{
			"_id":   0,
			"name":  "$_id",
			"value": "$total",
		}},
		{"$sort": bson.M{"value": 1}},
	}
	if ns != "" {
		match := pipeline[0]["$match"].(bson.M)
		match["ns"] = ns
		pipeline[0]["$match"] = match
	}
	if duration != "" {
		match := pipeline[0]["$match"].(bson.M)
		toks := strings.Split(duration, ",")
		match["date"] = bson.M{"$gte": toks[0], "$lt": toks[1]}
		pipeline[0]["$match"] = match
	}
	collection := ptr.db.Collection(ptr.hatchetName)
	opts := options.Aggregate().SetAllowDiskUse(true)
	if ptr.verbose {
		fmt.Println(gox.Stringify(pipeline))
	}
	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return docs, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var doc NameValue
		if err = cursor.Decode(&doc); err != nil {
			return docs, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

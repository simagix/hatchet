/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * mongo_audit.go
 */

package hatchet

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (ptr *MongoDB) GetAuditData() (map[string][]NameValues, error) {
	var err error
	data := map[string][]NameValues{}
	ctx := context.Background()

	// get max connection counts
	collection := ptr.db.Collection(ptr.hatchetName + "_clients")
	pipeline := []bson.M{
		{"$group": bson.M{"_id": nil, "maxConns": bson.M{"$max": "$conns"}}},
	}
	cur, err := collection.Aggregate(ctx, pipeline)
	category := "stats"
	if err == nil && cur.Next(ctx) {
		var m bson.M
		if err = cur.Decode(&m); err != nil {
			return data, err
		}
		doc := NameValues{}
		doc.Name = "maxConns"
		value := ToInt(m["maxConns"])
		doc.Values = append(doc.Values, value)
		if value > 0 {
			data[category] = append(data[category], doc)
		}
	}
	defer cur.Close(ctx)

	// get data of max milli, total milli, and average milli
	collection = ptr.db.Collection(ptr.hatchetName + "_ops")
	pipeline = []bson.M{
		{"$group": bson.M{"_id": nil, "max_ms": bson.M{"$max": "$max_ms"},
			"count": bson.M{"$sum": "$count"}, "total_ms": bson.M{"$sum": "$total_ms"}}},
	}
	cur, err = collection.Aggregate(ctx, pipeline)
	if err == nil && cur.Next(ctx) {
		var m bson.M
		if err = cur.Decode(&m); err != nil {
			return data, err
		}
		count := ToInt(m["count"])
		if count > 0 {
			val := ToInt(m["max_ms"])
			data[category] = append(data[category], NameValues{"maxMilli", []interface{}{val}})
			data[category] = append(data[category], NameValues{"avgMilli", []interface{}{val / count}})
			val = ToInt(m["total_ms"])
			data[category] = append(data[category], NameValues{"totalMilli", []interface{}{val}})
		}
	}
	defer cur.Close(ctx)

	// get collscan data of max milli, total milli, and average milli
	category = "collscan"
	collection = ptr.db.Collection(ptr.hatchetName + "_ops")
	pipeline = []bson.M{
		{"$match": bson.M{"_index": "COLLSCAN"}},
		{"$group": bson.M{"_id": nil, "max_ms": bson.M{"$max": "$max_ms"},
			"count": bson.M{"$sum": "$count"}, "total_ms": bson.M{"$sum": "$total_ms"}}},
	}
	cur, err = collection.Aggregate(ctx, pipeline)
	if err == nil && cur.Next(ctx) {
		var m bson.M
		if err = cur.Decode(&m); err != nil {
			return data, err
		}
		count := ToInt(m["count"])
		if count > 0 {
			val := ToInt(m["max_ms"])
			data[category] = append(data[category], NameValues{"maxMilli", []interface{}{val}})
			data[category] = append(data[category], NameValues{"avgMilli", []interface{}{val / count}})
			val = ToInt(m["total_ms"])
			data[category] = append(data[category], NameValues{"totalMilli", []interface{}{val}})
		}
	}
	defer cur.Close(ctx)

	// get audit data of exception, failed, op, and duration
	filter := bson.M{"type": bson.M{"$in": []interface{}{"exception", "failed", "op", "duration"}}}
	opts := options.Find().SetSort(bson.D{{Key: "type", Value: 1}, {Key: "value", Value: -1}})
	if cur, err = ptr.db.Collection(ptr.hatchetName+"_audit").Find(ctx, filter, opts); err != nil {
		return data, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var auditData struct {
			Type  string `bson:"type"`
			Name  string `bson:"name"`
			Value int    `bson:"value"`
		}
		if err := cur.Decode(&auditData); err != nil {
			return data, err
		}
		doc := NameValues{}
		category = auditData.Type
		if category == "exception" {
			if doc.Name == "E" {
				doc.Name = "Error"
			} else if doc.Name == "F" {
				doc.Name = "Fatal"
			} else if doc.Name == "W" {
				doc.Name = "Warn"
			}
		}
		doc.Name = auditData.Name
		doc.Values = append(doc.Values, auditData.Value)
		data[category] = append(data[category], doc)
	}

	// get reslen-ip and reslen-ns data
	for _, category := range []string{"ip", "ns"} {
		collection = ptr.db.Collection(ptr.hatchetName + "_audit")
		pipeline = []bson.M{
			{"$match": bson.M{
				"type": category,
			}},
			{"$lookup": bson.M{
				"from":         ptr.hatchetName + "_audit",
				"localField":   "name",
				"foreignField": "name",
				"as":           "reslen",
			}},
			{"$unwind": bson.M{
				"path": "$reslen",
			}},
			{"$match": bson.M{
				"reslen.type": "reslen-" + category,
			}},
			{"$sort": bson.M{
				"reslen.value": -1,
			}},
			{"$project": bson.M{
				"_id":    0,
				"name":   1,
				"count":  "$value",
				"reslen": "$reslen.value",
			}},
		}
		if cur, err = collection.Aggregate(ctx, pipeline); err != nil {
			return data, err
		}
		defer cur.Close(ctx)
		for cur.Next(ctx) {
			var reslenData struct {
				Name   string `bson:"name"`
				Count  int    `bson:"count"`
				Reslen int    `bson:"reslen"`
			}
			if err := cur.Decode(&reslenData); err != nil {
				return data, err
			}
			doc := NameValues{}
			doc.Name = reslenData.Name
			doc.Values = append(doc.Values, reslenData.Count)
			doc.Values = append(doc.Values, reslenData.Reslen)
			data[category] = append(data[category], doc)
		}
	}

	// get drivers data
	category = "driver"
	pipeline = []bson.M{
		{"$group": bson.M{
			"_id": bson.M{"ip": "$ip", "driver": "$driver", "version": "$version"},
		}},
		{"$project": bson.M{
			"_id":     0,
			"ip":      "$_id.ip",
			"driver":  "$_id.driver",
			"version": "$_id.version",
		}},
	}
	if cur, err = ptr.db.Collection(ptr.hatchetName+"_drivers").Aggregate(ctx, pipeline); err != nil {
		return data, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var clientData struct {
			IP      string `bson:"ip"`
			Driver  string `bson:"driver"`
			Version string `bson:"version"`
		}
		if err := cur.Decode(&clientData); err != nil {
			return data, err
		}
		var doc NameValues
		doc.Name = clientData.IP
		doc.Values = append(doc.Values, clientData.Driver)
		doc.Values = append(doc.Values, clientData.Version)
		data[category] = append(data[category], doc)
	}
	return data, err
}

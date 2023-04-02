/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * mongo_logs.go
 */

package hatchet

import (
	"context"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (ptr *MongoDB) GetSlowOps(orderBy string, order string, collscan bool) ([]OpStat, error) {
	db := ptr.db
	sortOrder := 1
	if order == "DESC" {
		sortOrder = -1
	}
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_index": "COLLSCAN",
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"op":     "$op",
					"ns":     "$ns",
					"filter": "$filter",
					"_index": "$_index",
				},
				"count":    bson.M{"$sum": "$count"},
				"avg_ms":   bson.M{"$avg": "$avg_ms"},
				"max_ms":   bson.M{"$max": "$max_ms"},
				"total_ms": bson.M{"$sum": "$total_ms"},
				"reslen":   bson.M{"$sum": "$reslen"},
			},
		},
		{
			"$project": bson.M{
				"_id":           0,
				"op":            "$_id.op",
				"count":         1,
				"avg_ms":        bson.M{"$round": []interface{}{"$avg_ms", 0}},
				"max_ms":        1,
				"total_ms":      1,
				"ns":            "$_id.ns",
				"index":         "$_id._index",
				"reslen":        1,
				"query_pattern": "$_id.filter",
			},
		},
		{
			"$sort": bson.M{
				orderBy: sortOrder,
			},
		},
	}
	if !collscan {
		pipeline[0]["$match"] = bson.M{"op": bson.M{"$nin": []interface{}{nil, ""}}}
	}
	if ptr.verbose {
		log.Println(pipeline)
	}
	var ops []OpStat
	cur, err := db.Collection(ptr.hatchetName+"_ops").Aggregate(context.Background(), pipeline)
	if err != nil {
		return ops, err
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		var op OpStat
		if err = cur.Decode(&op); err != nil {
			return ops, err
		}
		ops = append(ops, op)
	}
	if err = cur.Err(); err != nil {
		return ops, err
	}
	return ops, nil
}

func (ptr *MongoDB) GetLogs(opts ...string) ([]LegacyLog, error) {
	docs := []LegacyLog{}
	collection := ptr.db.Collection(ptr.hatchetName)
	search := ""
	qlimit := LIMIT + 1
	var offset, nlimit int
	ctx := context.Background()

	filter := bson.M{}
	for _, opt := range opts {
		toks := strings.Split(opt, "=")
		if len(toks) < 2 || toks[1] == "" {
			continue
		}
		if toks[0] == "duration" {
			dates := strings.Split(toks[1], ",")
			filter["date"] = bson.M{"$gte": dates[0], "$lt": dates[1]}
		} else if toks[0] == "limit" {
			offset, nlimit = GetOffsetLimit(toks[1])
			qlimit = ToInt(nlimit) + 1
		} else if toks[0] == "severity" {
			severities := []string{}
			for _, v := range SEVERITIES {
				severities = append(severities, v)
				if v == toks[1] {
					break
				}
			}
			filter["severity"] = bson.M{"$in": severities}
		} else {
			filter[toks[0]] = EscapeString(toks[1])
			if toks[0] == "context" {
				search = toks[1]
			}
		}
	}

	fopts := options.Find().SetSkip(int64(offset)).SetLimit(int64(qlimit))
	cursor, err := collection.Find(ctx, filter, fopts)
	if err != nil {
		return docs, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var doc LegacyLog
		if err = cursor.Decode(&doc); err != nil {
			return docs, err
		}
		docs = append(docs, doc)
	}
	if len(docs) == 0 && search != "" { // no context found, perform message search
		return ptr.SearchLogs(opts...)
	}
	return docs, nil
}

func (ptr *MongoDB) SearchLogs(opts ...string) ([]LegacyLog, error) {
	docs := []LegacyLog{}
	collection := ptr.db.Collection(ptr.hatchetName)
	qlimit := LIMIT + 1
	var offset, nlimit int
	ctx := context.Background()

	filter := bson.M{}
	for _, opt := range opts {
		toks := strings.Split(opt, "=")
		if len(toks) < 2 || toks[1] == "" {
			continue
		}
		if toks[0] == "duration" {
			dates := strings.Split(toks[1], ",")
			filter["date"] = bson.M{"$gte": dates[0], "$lt": dates[1]}
		} else if toks[0] == "limit" {
			offset, nlimit = GetOffsetLimit(toks[1])
			qlimit = ToInt(nlimit) + 1
		} else if toks[0] == "severity" {
			severities := []string{}
			for _, v := range SEVERITIES {
				severities = append(severities, v)
				if v == toks[1] {
					break
				}
			}
			filter["severity"] = bson.M{"$in": severities}
		} else if toks[0] == "context" {
			filter["message"] = bson.M{"$regex": primitive.Regex{Pattern: toks[1], Options: "i"}}
		} else {
			filter[toks[0]] = EscapeString(toks[1])
		}
	}

	fopts := options.Find().SetSkip(int64(offset)).SetLimit(int64(qlimit))
	cursor, err := collection.Find(ctx, filter, fopts)
	if err != nil {
		return docs, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var doc LegacyLog
		if err = cursor.Decode(&doc); err != nil {
			return docs, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func (ptr *MongoDB) GetSlowestLogs(topN int) ([]LegacyLog, error) {
	collection := ptr.db.Collection(ptr.hatchetName)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"op": bson.M{"$nin": []interface{}{nil, ""}},
			},
		},
		{
			"$sort": bson.M{
				"milli": -1,
			},
		},
		{
			"$limit": topN,
		},
	}
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var docs []LegacyLog
	for cursor.Next(context.Background()) {
		var doc LegacyLog
		if err = cursor.Decode(&doc); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, err
}

package entity

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoInf export
type MongoInf struct {
	col *mongo.Collection
}

// NewMongoInf export
func NewMongoInf(col *mongo.Collection) *MongoInf {
	return &MongoInf{col: col}
}

func ctxTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

// isOperatorUpdate reports whether the update doc contains $-prefixed mutation
// operators (e.g. $set, $unset). mgo's Update/Upsert dispatched on this implicitly;
// mongo-driver splits it into UpdateOne (operators) vs ReplaceOne (whole document).
func isOperatorUpdate(update interface{}) bool {
	m, ok := update.(bson.M)
	if !ok {
		return false
	}
	for k := range m {
		if strings.HasPrefix(k, "$") {
			return true
		}
	}
	return false
}

// Update reutrn the error if update fail
func (mi *MongoInf) Update(selector, update interface{}) error {
	ctx, cancel := ctxTimeout()
	defer cancel()
	var err error
	if isOperatorUpdate(update) {
		_, err = mi.col.UpdateOne(ctx, selector, update)
	} else {
		_, err = mi.col.ReplaceOne(ctx, selector, update)
	}
	return err
}

// Upsert export
func (mi *MongoInf) Upsert(selector, update interface{}) (interface{}, error) {
	ctx, cancel := ctxTimeout()
	defer cancel()
	upsert := true
	if isOperatorUpdate(update) {
		return mi.col.UpdateOne(ctx, selector, update, &options.UpdateOptions{Upsert: &upsert})
	}
	return mi.col.ReplaceOne(ctx, selector, update, &options.ReplaceOptions{Upsert: &upsert})
}

// Insert return the error if insert fail
func (mi *MongoInf) Insert(docs ...interface{}) error {
	ctx, cancel := ctxTimeout()
	defer cancel()
	if len(docs) == 1 {
		_, err := mi.col.InsertOne(ctx, docs[0])
		return err
	}
	_, err := mi.col.InsertMany(ctx, docs)
	return err
}

// Remove return the error if remove fail
func (mi *MongoInf) Remove(selector interface{}) error {
	ctx, cancel := ctxTimeout()
	defer cancel()
	_, err := mi.col.DeleteOne(ctx, selector)
	return err
}

// RemoveAll return the info and error
func (mi *MongoInf) RemoveAll(selector interface{}) (interface{}, error) {
	ctx, cancel := ctxTimeout()
	defer cancel()
	return mi.col.DeleteMany(ctx, selector)
}

// FindOne get very first match query result
func (mi *MongoInf) FindOne(key, res interface{}) (interface{}, error) {
	ctx, cancel := ctxTimeout()
	defer cancel()
	err := mi.col.FindOne(ctx, key).Decode(res)
	return res, err
}

// FindAll get every match query result
func (mi *MongoInf) FindAll(key, res interface{}) (interface{}, error) {
	ctx, cancel := ctxTimeout()
	defer cancel()
	cursor, err := mi.col.Find(ctx, key)
	if err != nil {
		return res, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, res)
	return res, err
}

// FindSelectOne export
func (mi *MongoInf) FindSelectOne(key, sel, res interface{}) (interface{}, error) {
	ctx, cancel := ctxTimeout()
	defer cancel()
	err := mi.col.FindOne(ctx, key, options.FindOne().SetProjection(sel)).Decode(res)
	return res, err
}

// FindSelectAll export
func (mi *MongoInf) FindSelectAll(key, sel, res interface{}) (interface{}, error) {
	ctx, cancel := ctxTimeout()
	defer cancel()
	cursor, err := mi.col.Find(ctx, key, options.Find().SetProjection(sel))
	if err != nil {
		return res, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, res)
	return res, err
}

// CountDocuments export — replaces direct collection counting from main.go.
func (mi *MongoInf) CountDocuments() (int64, error) {
	ctx, cancel := ctxTimeout()
	defer cancel()
	return mi.col.CountDocuments(ctx, bson.M{})
}

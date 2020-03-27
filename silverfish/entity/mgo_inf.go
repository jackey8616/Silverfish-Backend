package entity

import (
	"gopkg.in/mgo.v2"
)

// MongoInf export
type MongoInf struct {
	session *mgo.Session
	col     *mgo.Collection
}

// NewMongoInf export
func NewMongoInf(session *mgo.Session, col *mgo.Collection) *MongoInf {
	mi := new(MongoInf)
	mi.session = session
	mi.col = col
	return mi
}

// Find return the result query in mgo.Collection
// This method is implementation of interface Fetch#Find method.
func (mi *MongoInf) Find(queryKey interface{}) interface{} {
	return mi.col.Find(queryKey)
}

// Update reutrn the error if update fail
func (mi *MongoInf) Update(selector, update interface{}) error {
	return mi.col.Update(selector, update)
}

// Upsert export
func (mi *MongoInf) Upsert(selector, update interface{}) (interface{}, error) {
	return mi.col.Upsert(selector, update)
}

// Insert return the error if insert fail
func (mi *MongoInf) Insert(doc ...interface{}) error {
	return mi.col.Insert(doc...)
}

// Remove return the error if remove fail
func (mi *MongoInf) Remove(selector interface{}) error {
	return mi.col.Remove(selector)
}

// RemoveAll return the info and error
func (mi *MongoInf) RemoveAll(selector interface{}) (interface{}, error) {
	return mi.col.RemoveAll(selector)
}

// FindOne get very first match query result
func (mi *MongoInf) FindOne(key, res interface{}) (interface{}, error) {
	query := mi.Find(key)
	value, _ := query.(*mgo.Query)
	err := value.One(res)
	return res, err
}

// FindAll get every match query result
func (mi *MongoInf) FindAll(key, res interface{}) (interface{}, error) {
	query := mi.Find(key)
	value, _ := query.(*mgo.Query)
	err := value.All(res)
	return res, err
}

// FindSelectOne export
func (mi *MongoInf) FindSelectOne(key, sel, res interface{}) (interface{}, error) {
	query := mi.Find(key)
	value, _ := query.(*mgo.Query)
	err := value.Select(sel).One(res)
	return res, err
}

// FindSelectAll export
func (mi *MongoInf) FindSelectAll(key, sel, res interface{}) (interface{}, error) {
	query := mi.Find(key)
	value, _ := query.(*mgo.Query)
	err := value.Select(sel).All(res)
	return res, err
}

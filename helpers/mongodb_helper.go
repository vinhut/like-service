package helpers

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"
)

type DatabaseHelper interface {
	Query(string, map[string]string, interface{}) error
	QueryAll(string, string, string, interface{}) ([]interface{}, error)
	FindAll(string, interface{}) ([]interface{}, error)
	Insert(string, interface{}) error
	Upsert(string, map[string]string, interface{}) error
	Delete(string, map[string]string) error
}

type MongoDBHelper struct {
	client *mongo.Client
	db     *mongo.Database
}

func toDoc(v interface{}) (doc *bson.D, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}

func NewMongoDatabase() DatabaseHelper {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URL")))
	log.Print(os.Getenv("MONGO_URL"))
	if err != nil {
		fmt.Println(err)
	}

	db := client.Database(os.Getenv("MONGO_DATABASE"))
	log.Print(os.Getenv("MONGO_DATABASE"))
	return &MongoDBHelper{
		client: client,
		db:     db,
	}
}

func (mdb *MongoDBHelper) Query(collectionName string, query map[string]string, data interface{}) error {

	collection := mdb.db.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := collection.FindOne(ctx, query)
	err := result.Decode(data)
	if err != nil {
		fmt.Println("helper mongodb : ", err)
		return err
	}
	return nil
}

func (mdb *MongoDBHelper) QueryAll(collectionName string, key string, value string, obj interface{}) ([]interface{}, error) {

	collection := mdb.db.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.M{key: value})
	if err != nil {
		fmt.Println("finding fail ", err)
		return nil, err
	}
	defer cur.Close(ctx)

	var container = make([]interface{}, 0)
	for cur.Next(ctx) {

		model := reflect.New(reflect.TypeOf(obj)).Interface()
		decode_err := cur.Decode(model)
		if decode_err != nil {
			fmt.Println("decode fail ", decode_err)
			return nil, decode_err
		}

		fmt.Println("obj = ", obj)
		fmt.Println("model = ", model)
		md := reflect.ValueOf(model).Elem().Interface()
		fmt.Println("md = ", md)
		container = append(container, md)
		fmt.Println("container = ", container)
	}

	return container, nil
}

func (mdb *MongoDBHelper) FindAll(collectionName string, obj interface{}) ([]interface{}, error) {

	collection := mdb.db.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.D{{}})
	if err != nil {
		fmt.Println("finding fail ", err)
		return nil, err
	}
	defer cur.Close(ctx)

	var container = make([]interface{}, 0)
	for cur.Next(ctx) {

		model := reflect.New(reflect.TypeOf(obj)).Interface()
		decode_err := cur.Decode(model)
		if decode_err != nil {
			fmt.Println("decode fail ", decode_err)
			return nil, decode_err
		}

		fmt.Println("obj = ", obj)
		fmt.Println("model = ", model)
		md := reflect.ValueOf(model).Elem().Interface()
		fmt.Println("md = ", md)
		container = append(container, md)
		fmt.Println("container = ", container)
	}

	return container, nil

}

func (mdb *MongoDBHelper) Insert(collectionName string, data interface{}) error {
	collection := mdb.db.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	new_user, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	_, err = collection.InsertOne(ctx, new_user)

	if err != nil {
		fmt.Println("Got a real error:", err.Error())
		return err
	}

	return err
}

func (mdb *MongoDBHelper) Upsert(collectionName string, query map[string]string, data interface{}) error {
	collection := mdb.db.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	new_data, err := toDoc(data)
	if err != nil {
		return err
	}
	update := bson.D{{"$set", new_data}}
	opts := options.Update().SetUpsert(true)

	result, err := collection.UpdateOne(ctx, query, update, opts)
	if err != nil {
		return err
	}

	if result.MatchedCount != 0 {
		fmt.Println("matched and replaced an existing document")
		return nil
	}
	if result.UpsertedCount != 0 {
		fmt.Printf("inserted a new document with ID %v\n", result.UpsertedID)
	}

	return nil
}

func (mdb *MongoDBHelper) Delete(collectionName string, query map[string]string) error {

	collection := mdb.db.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := collection.DeleteOne(ctx, query)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

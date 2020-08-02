package main

import (
	"context"
	"time"
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
)

var collection *mongo.Collection

type Person struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`
	Age         int                `json:"age,omitempty" bson:"age,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
}

func createPerson(c *gin.Context) {
	c.Writer.Header().Add("Content-type", "application/json")
	var person Person
	c.ShouldBindJSON(&person)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, _ := collection.InsertOne(ctx, person)
	id := result.InsertedID
	person.ID = id.(primitive.ObjectID)
	json.NewEncoder(c.Writer).Encode(person)
}

func getPersonByID(c *gin.Context) {
	c.Writer.Header().Add("content-type", "application/json")
	var person Person
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := collection.FindOne(ctx, Person{ID: id}).Decode(&person)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		c.Writer.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(c.Writer).Encode(person)
}

func getAllPerson(c *gin.Context) {
	c.Writer.Header().Add("Content-type", "application/json")
	var persons []Person
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		c.Writer.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var person Person
		cursor.Decode(&person)
		persons = append(persons, person)
	}
	if err := cursor.Err(); err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		c.Writer.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(c.Writer).Encode(persons)
}

func updatePerson(c *gin.Context) {
	c.Writer.Header().Add("Cntent-type", "application/json")
	var person Person
	fmt.Println(c.Param("id"))
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	c.ShouldBindJSON(&person)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := collection.UpdateOne(ctx, Person{ID: id}, bson.M{"$set": person})
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		c.Writer.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(c.Writer).Encode(result)
}

func deletePerson(c *gin.Context) {
	c.Writer.Header().Add("Content-type", "application/json")
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		c.Writer.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(c.Writer).Encode(result)
}

func home(c *gin.Context) {
	fmt.Fprintln(c.Writer, "Wellcome")
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOption := options.Client().ApplyURI("mongodb://localhost:27017")
	cleint, _ := mongo.Connect(ctx, clientOption)
	collection = cleint.Database("test_go").Collection("person")
	r := gin.Default()
	r.GET("/", home)
	r.POST("/person", createPerson)
	r.GET("/person", getAllPerson)
	r.GET("/person/:id", getPersonByID)
	r.PUT("/person/:id", updatePerson)
	r.DELETE("/person/:id", deletePerson)
	r.Run()
}
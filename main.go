package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	POST   = "POST"
	GET    = "GET"
	PUT    = "PUT"
	DELETE = "DELETE"
)

type Post struct {
	Id        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title     string             `json:"title"`
	Content   string             `json:"content"`
	Author    string             `json:"author"`
	Status    string             `json:"status"`
	CreatedAt time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

var collection *mongo.Collection
var client *mongo.Client

//Execution starts
func main() {

	dbClientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	var err error
	client, err = mongo.Connect(context.Background(), dbClientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	collection = client.Database("blog").Collection("posts")

	http.HandleFunc("/post", createPost)
	http.HandleFunc("/posts", getAllPost)
	http.HandleFunc("/post/", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case GET:
			getPost(w, r)
		case PUT:
			updatePost(w, r)
		case DELETE:
			deletePost(w, r)
		default:
			http.Error(w, "Method not Implemented", http.StatusNotImplemented)
			return

		}
	})

	log.Println("listening to port:8080")
	http.ListenAndServe(":8080", nil)
}

// create new blog post in db
func createPost(w http.ResponseWriter, r *http.Request) {

	if r.Method != POST {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var blogPost Post
	if err := json.NewDecoder(r.Body).Decode(&blogPost); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	time := time.Now()
	blogPost.CreatedAt = time
	blogPost.UpdatedAt = time

	res, err := collection.InsertOne(context.TODO(), blogPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	blogPost.Id = res.InsertedID.(primitive.ObjectID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(blogPost)
}

//get particular blog post from db
func getPost(w http.ResponseWriter, r *http.Request) {

	var blogPost Post
	id := getIdFromURL(w, r)
	if id == primitive.NilObjectID {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&blogPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blogPost)
}

//Delete blog post
func deletePost(w http.ResponseWriter, r *http.Request) {

	id := getIdFromURL(w, r)
	if id == primitive.NilObjectID {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// Update blog post
func updatePost(w http.ResponseWriter, r *http.Request) {

	var blogPost Post
	if err := json.NewDecoder(r.Body).Decode(&blogPost); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := getIdFromURL(w, r)
	if id == primitive.NilObjectID {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	blogPost.UpdatedAt = time.Now()
	updateData := bson.M{"$set": blogPost}
	res, err := collection.UpdateOne(context.TODO(), bson.M{"_id": id}, updateData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	blogPost.Id = id
	if res.MatchedCount == 0 {
		blogPost.CreatedAt = time.Now()
		_, err := collection.InsertOne(context.TODO(), blogPost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(blogPost)
}

//Get all blog posts from db
func getAllPost(w http.ResponseWriter, r *http.Request) {

	if r.Method != GET {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	cur, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var blogPosts []Post
	for cur.Next(ctx) {
		var blogPost Post
		if err := cur.Decode(&blogPost); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		blogPosts = append(blogPosts, blogPost)
	}

	if err := json.NewEncoder(w).Encode(blogPosts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blogPosts)
}

// get id from URL path
func getIdFromURL(w http.ResponseWriter, r *http.Request) primitive.ObjectID {
	path := r.URL.Path
	seg := strings.Split(path, "/")
	if len(seg) != 3 {
		return primitive.NilObjectID
	}
	id, _ := primitive.ObjectIDFromHex(seg[2])
	return id
}

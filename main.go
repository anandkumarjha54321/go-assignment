package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
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
	Id        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

var collection *mongo.Collection

func handleByIdRequest(w http.ResponseWriter, r *http.Request) {

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

}

//Execution starts
func main() {

	dbClientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), dbClientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	collection = client.Database("blog").Collection("posts")

	http.HandleFunc("/post", createPost)
	http.HandleFunc("/posts", getAllPost)
	http.HandleFunc("/post/{id}", handleByIdRequest)

	log.Println("listening to port:8080")
	http.ListenAndServe(":8080", nil)
}

// create new blog post in db
func createPost(w http.ResponseWriter, r *http.Request) {

	if r.Method != POST {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var blogPost *Post
	if err := json.NewDecoder(r.Body).Decode(blogPost); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	blogPost.Id = uuid.New().String()
	time := time.Now()
	blogPost.CreatedAt = time
	blogPost.UpdatedAt = time

	_, err := collection.InsertOne(context.TODO(), blogPost)
	if err != nil {

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

//get particular blog post from db
func getPost(w http.ResponseWriter, r *http.Request) {

	if r.Method != GET {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var blogPost *Post
	id := getIdFromURL(w, r)

	err := collection.FindOne(context.Background(), bson.M{"id": id}).Decode(&blogPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blogPost)
}

//Delete blog post
func deletePost(w http.ResponseWriter, r *http.Request) {

	if r.Method != DELETE {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := getIdFromURL(w, r)

	_, err := collection.DeleteOne(context.TODO(), bson.M{"id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}

// Update blog post
func updatePost(w http.ResponseWriter, r *http.Request) {

	if r.Method != PUT {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var blogPost Post
	if err := json.NewDecoder(r.Body).Decode(&blogPost); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	blogPost.UpdatedAt = time.Now()
	id := getIdFromURL(w, r)
	updateData := bson.M{"$set": blogPost}

	_, err := collection.UpdateOne(context.TODO(), bson.M{"id": id}, updateData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}

//Get all blog posts from db
func getAllPost(w http.ResponseWriter, r *http.Request) {

	if r.Method != GET {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	cur, err := collection.Find(context.Background(), nil)
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
}

// get id from URL path
func getIdFromURL(w http.ResponseWriter, r *http.Request) string {
	path := r.URL.Path
	seg := strings.Split(path, "/")
	if len(seg) < 2 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return ""
	}
	id := seg[1]
	return id
}

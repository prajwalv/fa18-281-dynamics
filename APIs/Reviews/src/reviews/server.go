package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/unrolled/render"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
)


// NewServer configures and returns a Server.
func NewServer() *negroni.Negroni {
	formatter := render.New(render.Options{
		IndentJSON: true,
	})
	corsObj := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowedHeaders: []string{"Accept", "content-type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
	})
	n := negroni.Classic()
	mx := mux.NewRouter()
	initRoutes(mx, formatter)
	n.Use(corsObj)
	n.UseHandler(mx)
	return n
}


// MongoDB Config
var mongodb_server = "mongodb://admin:cmpe281@52.53.82.217:27017,54.177.200.126:27017,13.52.64.28:27017/groupproject?authSource=admin&replicaSet=cmpe281"
var mongodb_database = "groupproject"
var mongodb_collection = "reviews"


// API Routes
func initRoutes(mx *mux.Router, formatter *render.Render) {
	mx.HandleFunc("/ping", pingHandler(formatter)).Methods("GET")
	mx.HandleFunc("/getReviews/{itemName}", getReviewsHandler(formatter)).Methods("GET")
	mx.HandleFunc("/postReview", postReviewHandler(formatter)).Methods("POST")
	mx.HandleFunc("/updateReview", updateReviewHandler(formatter)).Methods("PUT")
	mx.HandleFunc("/deleteReview", deleteReviewHandler(formatter)).Methods("DELETE")

}

// API Ping Handler
func pingHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"Ping WORKS!!!"})
	}
}

// API Get All Reviews Handler
func getReviewsHandler(formatter *render.Render) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		params := mux.Vars(req)
		var ItemName string = params["itemName"]
		fmt.Println( "Item Name: ", ItemName )

		session, err := mgo.Dial(mongodb_server)
		if err != nil {
			panic(err)
		}

		if err != nil {
			fmt.Println("Reviews API (Get) - Unable to connect to MongoDB during read operation")
			panic(err)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB(mongodb_database).C(mongodb_collection)
		var results []Review
		err = c.Find(bson.M{"itemname": ItemName}).All(&results)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(results)
		if len(results) > 0 {
			formatter.JSON(w, http.StatusOK, results)
		}else{
			formatter.JSON(w, http.StatusNoContent, struct{ Response string }{"No reviews found for the given ID"})
		}
	}
}

// API Post a Review Handler.
func postReviewHandler(formatter *render.Render) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		var m Review
		_ = json.NewDecoder(req.Body).Decode(&m)
		fmt.Println("Review is: ", m.ItemName, " ", m.Reviews)
		session, err := mgo.Dial(mongodb_server)
		if err != nil {
			panic(err)
		}

		if err != nil {
			fmt.Println("Reviews API (Post) - Unable to connect to MongoDB during read operation")
			panic(err)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB(mongodb_database).C(mongodb_collection)
		entry := Review{
			ItemName: m.ItemName,
			Reviews: m.Reviews,
		}
		err = c.Insert(entry)
		if err != nil {
			fmt.Println("Error while adding reviews: ", err)
			formatter.JSON(w, http.StatusInternalServerError, struct{ Response error }{err})
		} else {
			formatter.JSON(w, http.StatusOK, struct{ Response string }{"Review added"})
		}
	}
}


// API Update a Review Handler.
func updateReviewHandler(formatter *render.Render) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		var m Review
		_ = json.NewDecoder(req.Body).Decode(&m)
		fmt.Println("Review is: ", m.ItemName, " " , "Reviews", m.Reviews)
		session, err := mgo.Dial(mongodb_server)
		if err != nil {
			panic(err)
		}

		if err != nil {
			fmt.Println("Reviews API (Update) - Unable to connect to MongoDB during read operation")
			panic(err)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB(mongodb_database).C(mongodb_collection)

		query := bson.M{
			"itemname": m.ItemName,
		}
		change := bson.M{"$push": bson.M{ "reviews" : bson.M{"$each": m.Reviews }}}
		err = c.Update(query, change)

		if err != nil {
			fmt.Println("Error while updating reviews: ", err)
			formatter.JSON(w, http.StatusInternalServerError, struct{ Response error }{err})
		} else {
			formatter.JSON(w, http.StatusOK, struct{ Response string }{"Review updated"})
		}
	}
}


// API Delete a Review Handler.
func deleteReviewHandler(formatter *render.Render) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		var m Review
		_ = json.NewDecoder(req.Body).Decode(&m)
		fmt.Println("Review is: ", m.Reviews)
		session, err := mgo.Dial(mongodb_server)
		if err != nil {
			panic(err)
		}

		if err != nil {
			fmt.Println("Reviews API (Delete) - Unable to connect to MongoDB during read operation")
			panic(err)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB(mongodb_database).C(mongodb_collection)

		query := bson.M{
			"itemname": m.ItemName,
		}
		change := bson.M{"$pull": bson.M{ "reviews" :  bson.M{"$in": m.Reviews } }}
		err = c.Update(query, change)

		if err != nil {
			fmt.Println("Error while deleting reviews: ", err)
			formatter.JSON(w, http.StatusInternalServerError, struct{ Response error }{err})
		} else {
			formatter.JSON(w, http.StatusOK, struct{ Response string }{"Review deleted"})
		}
	}
}

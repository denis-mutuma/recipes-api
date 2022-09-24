// Recipes API
//
// This is a sample recipes API. Find out more
// about the api at https://github.com/denis-mutuma/recipes-api
//
// Schemes: http
// Host: localhost:8080
// Basepath: /
// Version: 1.0.0
// Contact: Denis Mutuma
// <denis-mutuma@outlook.com> https://github.com/denis-mutuma
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
// swagger:meta

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var recipes []Recipe

// swagger:parameters recipe newRecipe
type Recipe struct {
	// swagger ignore
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}

// swagger: operation POST /recipes/ recipes newRecipe
// Create a recipe
// ---
// parameters:
// - name: name
//   description: name of recipe
//   requiured: true
//   type: string
// - tags: tags
//   description: tags for recipe
//   required: true
//   type: []string
// - ingredients: ingredients
//   description: ingredients for recipe
//   requiured: true
//   type: []string
// - instructions: instructions
//   description: instructions for recipe
//   requiured: true
//   type: []string
// produces:
// - application/json
// responses:
//		'200':
//			description: Successfull operation
//		'400':
// 			description: Invalid input
func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	_, err = collection.InsertOne(ctx, recipe)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while insertin new recipe"})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

// swagger:operation GET /recipes recipes listRecipes
// Returns list of recipes
// ---
// produces:
// - application/json
// responses:
//     '200':
// 	       description: Successful operation
func ListRecipesHandler(c *gin.Context) {
	curr, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer curr.Close(ctx)

	recipes := make([]Recipe, 0)
	for curr.Next(ctx) {
		var recipe Recipe
		curr.Decode(&recipe)
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, recipes)
}

//swagger: operation PUT /recipes/{id} recipes updateRecipe
// Update an existing recipe
// ---
// parameters:
// - name: id
//   in: path
//   description: ID of recipe
//   requiured: true
//   type: string
// produces:
// - application/json
// responses:
//		'200':
//			description: Successfull operation
//		'400':
// 			description: Invalid input
// 		'404':
//			description: Invalid recipe ID
func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err = collection.UpdateOne(ctx, bson.M{
		"_id": objectId,
	}, bson.D{{"$set", bson.D{
		{"name", recipe.Name},
		{"instructions", recipe.Instructions},
		{"ingredients", recipe.Ingredients},
		{"tags", recipe.Tags},
	}}})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been updated"})
}

//swagger: operation DELETE /recipes/{id} recipes updateRecipe
// Delete a recipe
// ---
// parameters:
// - name: id
//   in: path
//   description: ID of recipe
//   requiured: true
//   type: string
// produces:
// - application/json
// responses:
//		'200':
//			description: Successfull operation
//		'404':
// 			description: Invalid recipe ID
func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := collection.DeleteOne(ctx, bson.M{
		"_id": objectId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe deleted successfully"})
}

// swagger:operation GET /recipes/search recipes findRecipe
// Search recipes based on tags
// ---
// produces:
// - application/json
// parameters:
//   - name: tag
//     in: query
//     description: recipe tag
//     required: true
//     type: string
// responses:
//     '200':
//         description: Successful operation
func SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")
	listOfRecipes := make([]Recipe, 0)

	for i := 0; i < len(recipes); i++ {
		for _, v := range recipes[i].Tags {
			if strings.EqualFold(v, tag) {
				listOfRecipes = append(listOfRecipes, recipes[i])
			}
		}
	}
	c.JSON(http.StatusOK, listOfRecipes)
}

// swagger:operation GET /recipes/{id} recipe oneRecipe
// Get one recipe
// ---
// produces:
// - application/json
// parameters:
// 	- name: id
//	  in: path
//    description: ID of the recipe
//    required: true
//    type: string
// responses:
//    '200':
// 		  description: Successful opertion
//    '404':
//		  desctiption: Invalid recipe ID
func GetOneRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	curr := collection.FindOne(ctx, bson.M{
		"_id": objectId,
	})
	var recipe models.Recipe
	err := curr.Decode(&recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

var ctx context.Context
var err error
var client *mongo.Client

func init() {
	// recipes = make([]Recipe, 0)
	// file, _ := ioutil.ReadFile("./recipes.json")
	// _ = json.Unmarshal([]byte(file), &recipes)

	ctx = context.Background()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")

	// var listOfRecipes []interface{}
	// for _, recipe := range recipes {
	// 	listOfRecipes = append(listOfRecipes, recipe)
	// }
	// collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	// insertManyResult, err := collection.InsertMany(ctx, listOfRecipes)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println("Inserted recipes: ", len(insertManyResult.InsertedIDs))
}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipesHandler)
	router.GET("/recipes/:id", GetOneRecipeHandler)
	router.Run(":8080")
}

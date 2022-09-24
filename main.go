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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)
	file, _ := ioutil.ReadFile("./recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)
}

type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"publishedAt"`
}

//swagger: operation POST /recipes/ recipes newRecipe
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
	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
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
	index := -1
	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found"})
		return
	}
	recipes[index] = recipe
	c.JSON(http.StatusOK, recipe)
}

//swagger: operation PUT /recipes/{id} recipes updateRecipe
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
//		'400':
// 			description: Invalid input
func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	index := -1
	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
			break
		}
	}
	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found"})
		return
	}
	recipes = append(recipes[:index], recipes[index+1:]...)
	c.JSON(http.StatusOK, gin.H{
		"message": "Receipe deleted successfully"})
}

// swagger:operation GET /recipes tag recipes listRecipes
// Returns list of recipes matching tag name
// ---
// produces:
// - application/json
// responses:
//     '200':
// 	       description: Successful operation
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

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipesHandler)
	router.Run(":8080")
}

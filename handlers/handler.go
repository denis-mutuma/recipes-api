package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/denis-mutuma/recipes-api/models"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipesHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redisClient,
	}
}

// swagger:operation GET /recipes recipes listRecipes
// Returns list of recipes
// ---
// produces:
// - application/json
// responses:
//     '200':
// 	       description: Successful operation
func (handler *RecipesHandler) ListRecipesHandler(c *gin.Context) {
	val, err := handler.redisClient.Get("recipes").Result()

	if err == redis.Nil {
		log.Printf("Request to MongoDB")
		curr, err := handler.collection.Find(handler.ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer curr.Close(handler.ctx)

		recipes := make([]models.Recipe, 0)
		for curr.Next(handler.ctx) {
			var recipe models.Recipe
			curr.Decode(&recipe)
			recipes = append(recipes, recipe)
		}

		data, _ := json.Marshal(recipes)
		handler.redisClient.Set("recipes", string(data), 0)
		c.JSON(http.StatusOK, recipes)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		log.Printf("Request to Redis")
		recipes := make([]models.Recipe, 0)
		json.Unmarshal([]byte(val), &recipes)
		c.JSON(http.StatusOK, recipes)
	}
}

// swagger:operation POST /recipes recipes newRecipe
// Create a new recipe
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '400':
//         description: Invalid input
func (handler *RecipesHandler) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := handler.collection.InsertOne(handler.ctx, recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting new recipe"})
		return
	}

	log.Println("Remove data from Redis")
	handler.redisClient.Del("recipes")

	c.JSON(http.StatusOK, recipe)
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
func (handler *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.UpdateOne(handler.ctx, bson.M{
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
func (handler *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.DeleteOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe deleted successfully"})
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
func (handler *RecipesHandler) GetOneRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	curr := handler.collection.FindOne(handler.ctx, bson.M{
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
/*func SearchRecipesHandler(c *gin.Context) {
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
}*/

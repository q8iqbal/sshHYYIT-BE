package controllers

import (
	"backend-log-api/configs"
	"backend-log-api/models"
	"backend-log-api/responses"
	"context"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var logCollection *mongo.Collection = configs.GetCollection(configs.DB, "logs")
var validate = validator.New()

func PostLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var log models.Log
		defer cancel()

		//validate the request body
		if err := c.BindJSON(&log); err != nil {
			c.JSON(http.StatusBadRequest, responses.LogResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		//use the validator library to validate required fields
		if validationErr := validate.Struct(&log); validationErr != nil {
			c.JSON(http.StatusBadRequest, responses.LogResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": validationErr.Error()}})
			return
		}

		newLog := models.Log{
			Id:        primitive.NewObjectID(),
			Username:  log.Username,
			Status:    log.Status,
			State:     log.State,
			City:      log.City,
			Timestamp: log.Timestamp,
		}

		result, err := logCollection.InsertOne(ctx, newLog)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.LogResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		c.JSON(http.StatusCreated, responses.LogResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})

	}
}

func GetALog() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		logId := c.Param("logId")
		var log models.Log
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(logId)

		err := logCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&log)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.LogResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.LogResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": log}})
	}
}

func GetAllLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var logs []models.Log
		defer cancel()

		results, err := logCollection.Find(ctx, bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.LogResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		//reading from the db in an optimal way
		defer results.Close(ctx)
		for results.Next(ctx) {
			var singleLog models.Log
			if err = results.Decode(&singleLog); err != nil {
				c.JSON(http.StatusInternalServerError, responses.LogResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			}

			logs = append(logs, singleLog)
		}

		c.JSON(http.StatusOK,
			responses.LogResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": logs}},
		)
	}
}

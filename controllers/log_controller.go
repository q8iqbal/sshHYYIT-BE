package controllers

import (
	"backend-log-api/configs"
	"backend-log-api/models"
	"backend-log-api/responses"
	"context"
	"encoding/json"
	"io/ioutil"

	"fmt"
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

var (
	address  string
	err      error
	geo      models.GeoIP
	response *http.Response
	body     []byte
)

func GetGeoIP(ip_address string) (string, string, string) {

	ip := fmt.Sprintf("&ip=%s", ip_address)
	fields := "&fields=state_prov,district,country_name"

	response, err = http.Get(configs.EnvGeolocationBaseUrl() + configs.EnvGeolocationKey() + ip + fields)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()

	// response.Body() is a reader type. We have
	// to use ioutil.ReadAll() to read the data
	// in to a byte slice(string)
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	// Unmarshal the JSON byte slice to a GeoIP struct
	err = json.Unmarshal(body, &geo)
	if err != nil {
		fmt.Println(err)
	}
	return geo.District, geo.State_Prov, geo.Country_name
}

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

		district, state_prov, country_name := GetGeoIP(log.Ip_Guest)

		newLog := models.Log{
			Id:           primitive.NewObjectID(),
			Ip_Server:    log.Ip_Server,
			Hostname:     log.Hostname,
			Ip_Guest:     log.Ip_Guest,
			Username:     log.Username,
			Timestamp:    log.Timestamp,
			District:     district,
			State_Prov:   state_prov,
			Country_name: country_name,
			Status:       log.Status,
		}

		result, err := logCollection.InsertOne(ctx, newLog)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.LogResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		c.JSON(http.StatusCreated, responses.LogResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})

	}
}

func GetConnected() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "connected"}}}}

		count, err := logCollection.CountDocuments(ctx, filter)

		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.LogResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.LogResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"count": count}})
	}
}

func GetFailed() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "failed"}}}}

		count, err := logCollection.CountDocuments(ctx, filter)

		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.LogResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.LogResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"count": count}})
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

// func GetALog() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 		logId := c.Param("logId")
// 		var log models.Log
// 		defer cancel()

// 		objId, _ := primitive.ObjectIDFromHex(logId)

// 		err := logCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&log)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, responses.LogResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
// 			return
// 		}

// 		c.JSON(http.StatusOK, responses.LogResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": log}})
// 	}
// }

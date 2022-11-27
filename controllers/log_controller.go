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
	"golang.org/x/crypto/bcrypt"
)

var logCollection *mongo.Collection = configs.GetCollection(configs.DB, "logs")
var statusCollection *mongo.Collection = configs.GetCollection(configs.DB, "status")
var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "user")
var validate = validator.New()

var (
	address  string
	err      error
	geo      models.GeoIP
	response *http.Response
	body     []byte
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

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
func PostLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var user models.LoginUser
		var dbUser models.LoginUser
		defer cancel()

		//validate the request body
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, responses.LogResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		//use the validator library to validate required fields
		if validationErr := validate.Struct(&user); validationErr != nil {
			c.JSON(http.StatusBadRequest, responses.LogResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": validationErr.Error()}})
			return
		}

		count, errror := userCollection.EstimatedDocumentCount(ctx)
		if errror != nil {
			panic(errror)
		}

		if count < 1 {
			password := "admin"
			hash, _ := HashPassword(password)
			credentials := bson.D{{Key: "username", Value: "admin"}, {Key: "password", Value: hash}}
			userCollection.InsertOne(ctx, credentials)
		}

		err := userCollection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&dbUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.LoginResponse{Message: "failed"})
			return
		}
		userPass := user.Password
		dbPass := dbUser.Password
		passErr := CheckPasswordHash(userPass, dbPass)
		if passErr == false {
			c.JSON(http.StatusInternalServerError, responses.LoginResponse{Message: "wrong password"})
			return
		}
		c.JSON(http.StatusOK, responses.LoginResponse{Message: "success"})
	}
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
			Id:        primitive.NewObjectID(),
			Ip_Server: log.Ip_Server,
			Hostname:  log.Hostname,
			Ip_Guest:  log.Ip_Guest,
			Username:  log.Username,
			Timestamp: log.Timestamp,
			District:  district,
			State:     state_prov,
			Country:   country_name,
			Status:    log.Status,
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
			responses.LogResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"array": logs}},
		)
	}
}

func PostCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var currentServer models.CurrentServer
		defer cancel()

		//validate the request body
		if err := c.BindJSON(&currentServer); err != nil {
			c.JSON(http.StatusBadRequest, responses.LogResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		//use the validator library to validate required fields
		if validationErr := validate.Struct(&currentServer); validationErr != nil {
			c.JSON(http.StatusBadRequest, responses.LogResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": validationErr.Error()}})
			return
		}

		newServer := models.CurrentServer{}

		newServer.Ip_Server = currentServer.Ip_Server
		newServer.Hostname = currentServer.Hostname
		newServer.Users = currentServer.Users

		filter := bson.D{{Key: "ip_server", Value: bson.D{{Key: "$eq", Value: currentServer.Ip_Server}}}}

		count, err := logCollection.CountDocuments(ctx, filter)
		fmt.Println("TESTING :", count)
		if err != nil {
			panic(err)
		}
		if count >= 1 {

			filter := bson.D{{Key: "ip_server", Value: currentServer.Ip_Server}}

			result, err := statusCollection.UpdateOne(ctx, filter, bson.M{"$set": newServer})
			if err != nil {
				c.JSON(http.StatusInternalServerError, responses.LogResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
				return
			}

			c.JSON(http.StatusCreated, responses.LogResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})
		} else {

			result, err := statusCollection.InsertOne(ctx, newServer)
			if err != nil {
				c.JSON(http.StatusInternalServerError, responses.LogResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
				return
			}

			c.JSON(http.StatusCreated, responses.LogResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})

		}

	}
}

func GetAllCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var servers []models.CurrentServer
		defer cancel()

		results, err := statusCollection.Find(ctx, bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.LogResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		//reading from the db in an optimal way
		defer results.Close(ctx)
		for results.Next(ctx) {
			var singleServer models.CurrentServer
			if err = results.Decode(&singleServer); err != nil {
				c.JSON(http.StatusInternalServerError, responses.LogResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			}

			servers = append(servers, singleServer)
		}

		c.JSON(http.StatusOK,
			responses.LogResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"array": servers}},
		)
	}
}

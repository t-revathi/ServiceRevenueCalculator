package controller

import (
	mongo "api-traderevenuecalculator/service/mongodb"
	service "api-traderevenuecalculator/service/userservice"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/go-chi/render"
)

type UserController struct {
	userService *service.UserService
	dbService   *mongo.DBService
}

func NewUserController(db string) *UserController {
	return &UserController{
		userService: service.NewUserService(),
		dbService:   mongo.NewDBService(),
	}
}
func Health(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (u *UserController) WireRoutes(r chi.Router) {
	r.Route("/", func(r chi.Router) {
		r = r.With(Health)
		r.Post("/calculateRevenue", u.PerformCalculateProfit)

		r.Get("/healthCheck", u.healthCheck)
	})
}

func (u *UserController) PerformCalculateProfit(w http.ResponseWriter, r *http.Request) {
	var dataCalculateRevenue service.DataCalculateRevenue

	if err := render.DecodeJSON(r.Body, &dataCalculateRevenue); err != nil {
		return
	}
	result := u.userService.PerformCalculateProfit(r.Context(), w, r, &dataCalculateRevenue)

	//client, ctx, cancel, err := u.dbService.Connectdb("mongodb://localhost:27017/stockprofitcalculator")
	client, ctx, cancel, err := u.dbService.Connectdb("mongodb://0.0.0.0:27017/stockprofitcalculator")

	if err != nil {
		panic(err)
	}
	err = u.dbService.Pingdb(client, ctx)

	defer u.dbService.Closedb(client, ctx, cancel)
	if err != nil {
		fmt.Println("Couldn't connect to Database")
	}

	// Insert and Listing opertaions
	dbname := "stockprofitcalculator"
	collection := "plResults"

	doc := bson.D{{Key: "data", Value: result.Items}}

	res, err := u.dbService.Insertone(client, ctx, dbname, collection, doc)
	if err != nil {
		fmt.Println("Error Occured during insertion" + err.Error())
	}
	fmt.Println(res, err)
	//Listing the last inserted Record
	id, _ := res.InsertedID.(primitive.ObjectID)
	//primitive.ObjectIDFromHex(string("62ccdf87b79b0e2fc4ea67f0"))
	filter := bson.D{{Key: "_id", Value: id}}

	var record bson.M
	u.dbService.FindOne(client, ctx, dbname, collection, filter).Decode(&record)

	fmt.Println(record)

	//Return results to client
	render.JSON(w, r,
		result.Items)
}

func (u *UserController) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Healthcheck good"))
}

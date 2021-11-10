package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"github.com/dgrijalva/jwt-go"
)

func fixString(s string) string {
	return s[1 : len(s)-1]
}

type Claims struct {
	ID string `json:"id"`
	jwt.StandardClaims
}

type Point struct {
	Id           string  `json:"_id"`
	Description  string  `json:"description"`
	Name         string  `json:"name"`
	Event_type   string  `json:"event_type"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	UserNickname string  `json:"user_nickname"`
}

func main() {
	dbUrl := os.Getenv("MONGO")
	jwtKey := []byte(os.Getenv("JWT_KEY"))
	if len(jwtKey) == 0 {
		jwtKey = []byte("2S!3!TSvVRKzwCSS")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(dbUrl))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithCancel(context.Background())
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	var pointType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Point",
			Fields: graphql.Fields{
				"_id": &graphql.Field{
					Type: graphql.String,
				},
				"description": &graphql.Field{
					Type: graphql.String,
				},
				"name": &graphql.Field{
					Type: graphql.String,
				},
				"event_type": &graphql.Field{
					Type: graphql.String,
				},
				"latitude": &graphql.Field{
					Type: graphql.Float,
				},
				"longitude": &graphql.Field{
					Type: graphql.Float,
				},
				"user_nickname": &graphql.Field{
					Type: graphql.String,
				},
			},
		},
	)

	var queryType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"points": &graphql.Field{
					Type:        graphql.NewList(pointType),
					Description: "Get all points in range",
					Args: graphql.FieldConfigArgument{
						"latitude": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.Float),
						},
						"longitude": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.Float),
						},
						"radius": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.Float),
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						radius, _ := strconv.ParseFloat(fmt.Sprintf("%v", p.Args["radius"]), 64)
						longitude, _ := strconv.ParseFloat(fmt.Sprintf("%v", p.Args["longitude"]), 64)
						latitude, _ := strconv.ParseFloat(fmt.Sprintf("%v", p.Args["latitude"]), 64)
						arguments := []interface{}{
							bson.M{
								"$geoNear": bson.M{
									"near": bson.M{
										"type":        "Point",
										"coordinates": []float64{longitude, latitude},
									},
									"distanceField": "distance",
									"maxDistance":   radius,
								},
							},
							bson.M{
								"$lookup": bson.M{
									"from":         "users",
									"localField":   "user_id",
									"foreignField": "_id",
									"as":           "user",
								},
							},
						}
						cur, err := client.Database("ParduGO").Collection("points").Aggregate(ctx, arguments)
						if err != nil {
							return nil, nil
						}
						var points []Point

						for cur.Next(ctx) {
							points = append(points, Point{
								Id:           cur.Current.Lookup("_id").ObjectID().Hex(),
								Description:  fixString(cur.Current.Lookup("description").String()),
								Name:         fixString(cur.Current.Lookup("name").String()),
								Event_type:   fixString(cur.Current.Lookup("type").String()),
								Latitude:     cur.Current.Lookup("location").Document().Lookup("coordinates").Array().Lookup("1").Double(),
								Longitude:    cur.Current.Lookup("location").Document().Lookup("coordinates").Array().Lookup("0").Double(),
								UserNickname: fixString(cur.Current.Lookup("user").Array().Lookup("0").Document().Lookup("nickname").String()),
							})
						}
						// fmt.Println(points)
						return points, nil
					},
				},
			},
		},
	)

	var mutationType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Mutation",
			Fields: graphql.Fields{
				"register": &graphql.Field{
					Type:        graphql.String,
					Description: "Register new user",
					Args: graphql.FieldConfigArgument{
						"nickname": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"email": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"password": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						nickname := fmt.Sprintf("%v", p.Args["nickname"])
						email := fmt.Sprintf("%v", p.Args["email"])
						password := fmt.Sprintf("%v", p.Args["password"])
						hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

						testArgument := []interface{}{
							bson.M{
								"$match": bson.M{
									"$or": []interface{}{
										bson.M{
											"nickname": nickname,
										},
										bson.M{
											"email": email,
										},
									},
								},
							},
						}

						testCur, testErr := client.Database("ParduGO").Collection("users").Aggregate(ctx, testArgument)
						if testErr != nil {
							return nil, nil
						}

						if testCur.Next(ctx) {
							return nil, nil
						}

						argument := bson.M{
							"nickname": nickname,
							"email":    email,
							"password": string(hashedPassword),
						}

						cur, err := client.Database("ParduGO").Collection("users").InsertOne(ctx, argument)
						if err != nil {
							fmt.Println(err)
							return nil, nil
						}
						id := cur.InsertedID.(primitive.ObjectID).Hex()

						claims := &Claims{
							ID:             id,
							StandardClaims: jwt.StandardClaims{},
						}

						token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

						tokenString, err := token.SignedString(jwtKey)
						if err != nil {
							return nil, nil
						}

						return tokenString, nil
					},
				},
				"login": &graphql.Field{
					Type:        graphql.String,
					Description: "Login user",
					Args: graphql.FieldConfigArgument{
						"email": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"password": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						email := fmt.Sprintf("%v", p.Args["email"])
						password := fmt.Sprintf("%v", p.Args["password"])

						arguments := []interface{}{
							bson.M{
								"$match": bson.M{
									"email": email,
								},
							},
						}

						cur, err := client.Database("ParduGO").Collection("users").Aggregate(ctx, arguments)
						if err != nil {
							return nil, nil
						}
						if !cur.Next(ctx) {
							return nil, nil
						}
						curPassword := cur.Current.Lookup("password").String()
						isPasswordCorrect := bcrypt.CompareHashAndPassword([]byte(fixString(curPassword)), []byte(password))
						if isPasswordCorrect != nil {
							return nil, nil
						}

						id := cur.Current.Lookup("_id").ObjectID().Hex()

						claims := &Claims{
							ID:             id,
							StandardClaims: jwt.StandardClaims{},
						}

						token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

						tokenString, err := token.SignedString(jwtKey)
						if err != nil {
							return nil, nil
						}

						return tokenString, nil
					},
				},
				"create_point": &graphql.Field{
					Type:        graphql.Boolean,
					Description: "Create new point",
					Args: graphql.FieldConfigArgument{
						"token": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"name": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"description": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"latitude": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.Float),
						},
						"longitude": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.Float),
						},
						"type": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						token := fmt.Sprintf("%v", p.Args["token"])
						name := fmt.Sprintf("%v", p.Args["name"])
						description := fmt.Sprintf("%v", p.Args["description"])
						longitude, _ := strconv.ParseFloat(fmt.Sprintf("%v", p.Args["longitude"]), 64)
						latitude, _ := strconv.ParseFloat(fmt.Sprintf("%v", p.Args["latitude"]), 64)
						event_type := fmt.Sprintf("%v", p.Args["type"])

						claims := &Claims{}

						tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
							return jwtKey, nil
						})
						if err != nil {
							return false, nil
						}
						if !tkn.Valid {
							return false, nil
						}

						objectId, _ := primitive.ObjectIDFromHex(string(claims.ID))

						arguments := []interface{}{
							bson.M{
								"$match": bson.M{
									"_id": objectId,
								},
							},
						}
						userCur, userErr := client.Database("ParduGO").Collection("users").Aggregate(ctx, arguments)
						if userErr != nil {
							fmt.Println(err)
							return false, nil
						}
						if !userCur.Next(ctx) {
							return false, nil
						}

						argument := bson.M{
							"user_id":     objectId,
							"name":        name,
							"description": description,
							"type":        event_type,
							"location": bson.M{
								"type":        "Point",
								"coordinates": []float64{longitude, latitude},
							},
						}

						_, insErr := client.Database("ParduGO").Collection("points").InsertOne(ctx, argument)
						if insErr != nil {
							return false, nil
						}

						return true, nil
					},
				},
			},
		},
	)

	var schema, _ = graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    queryType,
			Mutation: mutationType,
		},
	)

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)
}

# 1000 foot view
My thought process going in was to make each package as singular in purpose as possible so that things could eventually be deployed in AWS lambdas for (theoretically) infinite scalability. 

I wanted to make my main action packages (meme_maker and user_service) relatively tooling agnostic. My meme maker shouldn't care about if I am using gin and MongoDB or Revel and PostgresQL. Honestly, I think I do a much better job with this on the users service and would like to fix this up some. 

I also embraced a Dependency Inversion approach to the users. The User Service has no reliance on mongodb, which makes it much easier to mock test solutions and not run costly queries. 

# Packages

## error_types
A collection of custom error types

## loggers
A simple collection of loggers

## meme_maker
Defines the `Meme` struct 
```go
type Meme struct {
	TopText       string `json:"top_text"`
	BottomText    string `json:"bottom_text"`
	ImageLocation string `json:"image_location"`
}
```
Includes a builder that takes a QueryParams struct and returns a meme.  
Currently the `with[struct field]` methods are private, but they could be exposed if needed down the line to enable more expandable memes. 

## query_params
Defines the `QueryParams` struct
```go
type QueryParams struct {
	Lon   float64 `json:"lon"`
	Lat   float64 `json:"lat"`
	Query string  `json:"query"`
}
```
Extracts Query Params from a `gin.context` to be fed into the `meme_maker` package. Acts as a middle layer between main and meme_maker.

## user_db
Implements the `user_service.UserRepository` interface handle all mongo db connections needed by the user service. 

## user_model
Defines the User struct. 
```go
type User struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	UserId          string             `bson:"user_id"`
	TokensRemaining int                `bson:"tokens_remaining"`
	AuthKey         string             `bson:"auth_key"`
	IsAdmin         bool               `bson:"is_admin"`
}

```

## user_service
Handles all user data logic and renders REST calls.
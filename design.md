# 1000 foot view
My thought process going in was to make each package as singular in purpose as possible so that things could eventually be deployed in AWS lambdas for (theoretically) infinite scalability. 

I wanted to make my main action packages (meme_maker and user_service) relatively tooling agnostic. My meme maker shouldn't care about if I am using gin and MongoDB or Revel and PostgresQL. I went about this by embracing the Dependency Inversion approach as much as I could. 

As an example, the users_service package defines a UserRepository interface, which is then used as part of the UserService struct. 

```go
type UserRepository interface {
	User(id string) (*models.User, error)
	NewUser(user models.User) (interface{}, error)
  ...
}

type UserService struct {
	Repo UserRepository
	Auth auth_service.AuthService
}

func NewUserService(repo UserRepository, auth auth_service.AuthService) *UserService {
	return &UserService{
		Repo: repo,
		Auth: auth,
	}
}
...
func (s *UserService) DoSomethingWithUser(id string) {
  user, err := s.Repo.User(id)
  ...
}
```

By defining the UserService this way, I am able to give it any user repository I want, so long as it implements the expected functions. There are a lot of benefits to this approach. First and foremost, it makes it much easier to test. All my user_service tests mock the database so that I am only testing what I actually care about with each unit test. Additionally, it helps with the separation of concerns and allows us to switch to a different database provider with relative risk as the user service has no concept of Mongo. You could even have one instance on mongo and another on postgres if you really wanted. 


# Packages

## auth_service
A simple auth service. Defines the AuthRepository interface, which is then implemented by `user_db`
```go
type AuthRepository interface {
	UserByAuthHeader(auth string) (*models.User, error)
}
```

## error_types
A collection of custom error types

## loggers
A simple collection of loggers

## meme_maker
Builds a meme based on query parameters. Implements the `meme_service.MemeProvider` interface. What really makes this whole dependency inversion thing so cool in this instance is that I was able to hide away the meme generation logic in this little meme maker, but if I had the time I could develop another MemeProvider that actually generates an image that gets stored elsewhere and it would change none of the meme_service code. 

## meme_service
Relies on an AuthService, a UserRepo (implemented by user_db) and a MemeProvider (implemented by meme_maker).

Extracts Query Params from a `gin.context` to be fed into it's provider's BuildMeme function. Acts as a middle layer between the API and whatever our meme source is.

## user_db
Implements the `user_service.UserRepository`, `meme_service.UserRepository`, and `auth_service.AuthRepository` interfaces.

By far the least unit-tested section. I was running into some trouble with the in-memory mongo instance and I think there is likely a better approach than what I did, but as someone new to mongo I was happy with getting my sample test written up. 

## models
Defines the User and Meme structs. 
```go
type User struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	UserId          string             `bson:"user_id"`
	TokensRemaining int                `bson:"tokens_remaining"`
	AuthKey         string             `bson:"auth_key"`
	IsAdmin         bool               `bson:"is_admin"`
}

```
```go
type Meme struct {
	TopText       string `json:"top_text"`
	BottomText    string `json:"bottom_text"`
	ImageLocation string `json:"image_location"`
}
```

## user_service
Handles all user data logic and renders REST calls. Most info is in the 1000 foot view section of this doc.
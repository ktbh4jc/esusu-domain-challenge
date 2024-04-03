# Curl commands

Once again, there is the postman data file. Docs on using it can be found [here](https://learning.postman.com/docs/getting-started/importing-and-exporting/importing-data/#import-postman-data). 

But, if you prefer terminal debugging, I've got the curl commands as well. 

#### Reset the DB
`curl --location --request POST 'localhost:8080/users/reset'`

#### Get all users regardless of admin rights (only for debugging purposes. Would not ship)
`curl --location 'localhost:8080/users/debug'`

#### Ping Mongo
`curl --location 'localhost:8080/mongo'`

#### Get Memes
```bash
curl --location 'localhost:8080/memes' \
--header 'auth: Alice-MemeMaster-Password'
```

#### Get Memes, this time with query parameters
```bash
curl --location 'localhost:8080/memes?lat=40.730610&lon=-73.935242&query=food' \
--header 'auth: Alice-MemeMaster-Password'
```

#### Get Memes - User has no tokens
```bash
curl --location 'localhost:8080/memes' \
--header 'auth: Bob-Password'
```

#### Get memes - Bad query parameters
```bash
curl --location 'localhost:8080/memes?lat=bad&lon=-73.935242&query=food' \
--header 'auth: Alice-MemeMaster-Password'
```

#### Get user by id - admin
For the next few I'm going to leave `660cb9237a3eb43df1682016` alone, but replace it with Alice Mememaster's object id to test. 

```bash
curl --location 'localhost:8080/users/660cb9237a3eb43df1682016' \
--header 'auth: Super-Secret-Password'
```

#### Get user by id - self
```bash
curl --location 'localhost:8080/users/660cb9237a3eb43df1682016' \
--header 'auth: Alice-MemeMaster-Password'
```

#### Get user by id - other
```bash
curl --location 'localhost:8080/users/660cb9237a3eb43df1682016' \
--header 'auth: Bob-Password'
```

#### Get all users - admin
```bash
curl --location 'localhost:8080/users' \
--header 'auth: Super-Secret-Password'
```

#### Get all users - non admin
```bash
curl --location 'localhost:8080/users' \
--header 'auth: Alice-MemeMaster-Password'
```

#### New User
```bash
curl --location 'localhost:8080/users' \
--header 'auth: Super-Secret-Password' \
--form 'user_id="That-Test-User"' \
--form 'tokens_remaining="50"' \
--form 'auth_key="Some-Auth-Key"' \
--form 'is_admin="False"'
```

#### Update user
```bash
curl --location --request PATCH 'localhost:8080/users/660cb9967a3eb43df1682018' \
--header 'auth: Super-Secret-Password' \
--form 'user_id="That-Test-User"' \
--form 'tokens_remaining="55"' \
--form 'auth_key="Some-Auth-Key"' \
--form 'is_admin="False"'
```
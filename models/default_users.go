package models

var DefaultUsers []interface{} = []interface{}{
	User{
		UserId:          "Adam Min",
		TokensRemaining: 100,
		IsAdmin:         true,
		AuthKey:         "Super-Secret-Password",
	}, User{
		UserId:          "Alice MemeMaster",
		TokensRemaining: 1000,
		IsAdmin:         false,
		AuthKey:         "Alice-MemeMaster-Password",
	}, User{
		UserId:          "No-Token Bob",
		TokensRemaining: 0,
		IsAdmin:         false,
		AuthKey:         "Bob-Password",
	},
}

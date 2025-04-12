package domain

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserDetail struct {
	ID   string `json:"id"`
	Age  int    `json:"age"`
	Mail string `json:"mail"`
}

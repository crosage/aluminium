package structs

type User struct {
	Uid      int    `json:"uid,omitempty"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Type     int    `json:"usertype,omitempty"`
}

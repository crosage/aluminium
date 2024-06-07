package structs

type File struct {
	Fid       int    `json:"fid"`
	Hash      string `json:"hash"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Uid       int    `json:"uid"`
	ShareCode string `json:"share_code"`
	Username  string `json:"username"`
}

type FileLog struct {
	Uid       int    `json:"uid"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Times     string `json:"timestamp"`
	Operation string `json:"operation"`
}

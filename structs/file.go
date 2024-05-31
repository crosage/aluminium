package structs

type File struct {
	Fid       int    `json:"fid"`
	Hash      string `json:"hash"`
	Path      string `json:"path"`
	Uid       int    `json:"uid"`
	ShareCode string `json:"share_code"`
}

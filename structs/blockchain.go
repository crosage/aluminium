package structs

type BlockChainFile struct {
	Hash string `json:"hash"`
	Uid  string `json:"uid"`
	Fid  string `json:"fid"`
}

type BlockChainUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

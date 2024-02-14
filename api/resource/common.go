package resource

type ReqHash struct {
	Hash string `uri:"hash" binding:"hexadecimal"`
}

type ResJSON struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
}

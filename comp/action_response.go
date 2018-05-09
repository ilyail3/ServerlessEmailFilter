package comp

type ActionResponse struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	EMail  string `json:"email"`
	ID     string `json:"id"`
	Admin  bool   `json:"admin"`
	New    bool   `json:"new"`
}

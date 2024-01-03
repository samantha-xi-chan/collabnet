package api_link

type ApiNode struct {
	LinkId     string `json:"link_id" `
	HostName   string `json:"host_name" ` //  gorm:"unique"
	FirstParty int    `json:"first_party"`
	From       string `json:"from" ` //  gorm:"unique"
	Online     int    `json:"online"`
}

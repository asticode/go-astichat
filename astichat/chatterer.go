package astichat

// Chatterer represents an entity willing to chat
type Chatterer struct {
	PublicKey PublicKey `json:"public_key" bson:"public_key"`
	Username  string    `json:"username" bson:"username"`
}

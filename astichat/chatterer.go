package astichat

// Chatterer represents an entity willing to chat
type Chatterer struct {
	ClientPublicKey  *PublicKey  `json:"public_key" bson:"client_public_key"`
	ServerPrivateKey *PrivateKey `json:"-" bson:"server_private_key"`
	Username         string      `json:"username" bson:"username"`
}

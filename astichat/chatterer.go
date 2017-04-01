package astichat

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Chatterer represents an entity willing to chat
type Chatterer struct {
	ClientPublicKey  *PublicKey    `json:"public_key" bson:"client_public_key"`
	ID               bson.ObjectId `json:"-" bson:"id"`
	ServerPrivateKey *PrivateKey   `json:"-" bson:"server_private_key"`
	Token            string        `json:"-" bson:"token"`
	TokenAt          time.Time     `json:"-" bson:"token_at"`
	Username         string        `json:"username" bson:"username"`
}

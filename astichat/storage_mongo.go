package astichat

import (
	"time"

	"github.com/rs/xlog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Constants
const (
	collectionNameChatterer = "chatterer"
	databaseName            = "astichat"
)

// ChattererMgo represents a mongo chatterer
type ChattererMgo struct {
	ClientPublicKey  *PublicKey    `bson:"client_public_key"`
	ID               bson.ObjectId `bson:"_id"`
	ServerPrivateKey *PrivateKey   `bson:"server_private_key"`
	Token            string        `bson:"token"`
	TokenAt          time.Time     `bson:"token_at"`
	Username         string        `bson:"username"`
}

// NewChattererMgoFromChatterer creates a mongo chatterer based on a chatterer
func NewChattererMgoFromChatterer(c Chatterer) ChattererMgo {
	return ChattererMgo{
		ClientPublicKey:  c.ClientPublicKey,
		ID:               bson.ObjectIdHex(c.ID),
		ServerPrivateKey: c.ServerPrivateKey,
		Token:            c.Token,
		TokenAt:          c.TokenAt,
		Username:         c.Username,
	}
}

// Chatterer creates a chatterer from the mongo chatterer
func (c ChattererMgo) Chatterer() Chatterer {
	return Chatterer{
		ClientPublicKey:  c.ClientPublicKey,
		ID:               c.ID.Hex(),
		ServerPrivateKey: c.ServerPrivateKey,
		Token:            c.Token,
		TokenAt:          c.TokenAt,
		Username:         c.Username,
	}
}

// StorageMongo represents a mongo storage
type StorageMongo struct {
	logger xlog.Logger
	mongo  *mgo.Session
}

// NewStorageMongo creates a new mongo storage
func NewStorageMongo(l xlog.Logger, s *mgo.Session) *StorageMongo {
	return &StorageMongo{
		logger: l,
		mongo:  s,
	}
}

// ChattererCreate creates a chatterer based on a username and a public key
func (s *StorageMongo) ChattererCreate(username string, pubClient *PublicKey, prvServer *PrivateKey) (c Chatterer, err error) {
	var mc = ChattererMgo{
		ClientPublicKey:  pubClient,
		ID:               bson.NewObjectId(),
		ServerPrivateKey: prvServer,
		Username:         username,
	}
	c = mc.Chatterer()
	err = s.mongo.DB(databaseName).C(collectionNameChatterer).Insert(&mc)
	return
}

// ChattererDeleteByUsername deletes a chatterer by its username
func (s *StorageMongo) ChattererDeleteByUsername(username string) error {
	return s.mongo.DB(databaseName).C(collectionNameChatterer).Remove(bson.M{"username": username})
}

// ChattererFetchByUsername fetches a chatterer by its username
func (s *StorageMongo) ChattererFetchByUsername(username string) (c Chatterer, err error) {
	var mc ChattererMgo
	if err = s.mongo.DB(databaseName).C(collectionNameChatterer).Find(bson.M{"username": username}).One(&mc); err == mgo.ErrNotFound {
		err = ErrNotFoundInStorage
	}
	if err == nil {
		c = mc.Chatterer()
	}
	return
}

// ChattererUpdate updates a chatterer
func (s *StorageMongo) ChattererUpdate(c Chatterer) error {
	var mc = NewChattererMgoFromChatterer(c)
	return s.mongo.DB(databaseName).C(collectionNameChatterer).UpdateId(mc.ID, mc)
}

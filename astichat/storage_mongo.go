package astichat

import (
	"github.com/rs/xlog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Constants
const (
	collectionNameChatterer = "chatterer"
	databaseName            = "astichat"
)

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
func (s *StorageMongo) ChattererCreate(username string, publicKey PublicKey) (c Chatterer, err error) {
	c = Chatterer{PublicKey: publicKey, Username: username}
	err = s.mongo.DB(databaseName).C(collectionNameChatterer).Insert(&c)
	return
}

// ChattererFetchByPublicKey fetches a chatterer by its public key
func (s *StorageMongo) ChattererFetchByPublicKey(publicKey PublicKey) (c Chatterer, err error) {
	if err = s.mongo.DB(databaseName).C(collectionNameChatterer).Find(bson.M{"public_key": publicKey}).One(&c); err == mgo.ErrNotFound {
		err = ErrNotFoundInStorage
	}
	return
}

// ChattererFetchByUsername fetches a chatterer by its username
func (s *StorageMongo) ChattererFetchByUsername(username string) (c Chatterer, err error) {
	if err = s.mongo.DB(databaseName).C(collectionNameChatterer).Find(bson.M{"username": username}).One(&c); err == mgo.ErrNotFound {
		err = ErrNotFoundInStorage
	}
	return
}

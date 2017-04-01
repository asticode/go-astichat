package astichat

import (
	"errors"

	"gopkg.in/mgo.v2/bson"
)

// Vars
var (
	ErrNotFoundInStorage = errors.New("not found in storage")
)

// Storage represents a storage interface
type Storage interface {
	ChattererCreate(username string, pubClient *PublicKey, prvServer *PrivateKey) (Chatterer, error)
	ChattererFetchByPublicKey(publicKey *PublicKey) (Chatterer, error)
	ChattererFetchByUsername(username string) (Chatterer, error)
	ChattererUpdate(i Chatterer) error
}

// NopStorage implements the Storage interface
type NopStorage struct{}

func (s NopStorage) ChattererCreate(username string, pubClient *PublicKey, prvServer *PrivateKey) (Chatterer, error) {
	return Chatterer{}, nil
}
func (s NopStorage) ChattererFetchByPublicKey(publicKey *PublicKey) (Chatterer, error) {
	return Chatterer{}, nil
}
func (s NopStorage) ChattererFetchByUsername(username string) (Chatterer, error) {
	return Chatterer{}, nil
}
func (s NopStorage) ChattererUpdate(i Chatterer) error {
	return nil
}

// MockedStorage represents a mocked storage
type MockedStorage struct {
	Chatterers []Chatterer
}

// NewMockedStorage creates a new mocked storage
func NewMockedStorage() *MockedStorage {
	return &MockedStorage{}
}

func (s *MockedStorage) ChattererCreate(username string, pubClient *PublicKey, prvServer *PrivateKey) (c Chatterer, err error) {
	c = Chatterer{ClientPublicKey: pubClient, ID: bson.NewObjectId(), ServerPrivateKey: prvServer, Username: username}
	s.Chatterers = append(s.Chatterers, c)
	return c, nil
}
func (s MockedStorage) ChattererFetchByPublicKey(publicKey *PublicKey) (Chatterer, error) {
	for _, c := range s.Chatterers {
		if publicKey.String() == c.ClientPublicKey.String() {
			return c, nil
		}
	}
	return Chatterer{}, ErrNotFoundInStorage
}
func (s MockedStorage) ChattererFetchByUsername(username string) (Chatterer, error) {
	for _, c := range s.Chatterers {
		if username == c.Username {
			return c, nil
		}
	}
	return Chatterer{}, ErrNotFoundInStorage
}
func (s *MockedStorage) ChattererUpdate(i Chatterer) error {
	for index, c := range s.Chatterers {
		if c.ID == i.ID {
			s.Chatterers[index] = i
			return nil
		}
	}
	return ErrNotFoundInStorage
}

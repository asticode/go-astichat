package astichat

import (
	"errors"
	"time"
)

// Vars
var (
	ErrNotFoundInStorage = errors.New("not found in storage")
)

// Chatterer represents an entity willing to chat
type Chatterer struct {
	ClientPublicKey  *PublicKey  `json:"public_key"`
	ID               string      `json:"-"`
	ServerPrivateKey *PrivateKey `json:"-"`
	Token            string      `json:"-"`
	TokenAt          time.Time   `json:"-"`
	Username         string      `json:"username"`
}

// Storage represents a storage interface
type Storage interface {
	ChattererCreate(username string, pubClient *PublicKey, prvServer *PrivateKey) (Chatterer, error)
	ChattererDeleteByUsername(username string) error
	ChattererFetchByUsername(username string) (Chatterer, error)
	ChattererUpdate(i Chatterer) error
}

// NopStorage implements the Storage interface
type NopStorage struct{}

func (s NopStorage) ChattererCreate(username string, pubClient *PublicKey, prvServer *PrivateKey) (Chatterer, error) {
	return Chatterer{}, nil
}
func (s NopStorage) ChattererDeleteByUsername(username string) error {
	return nil
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
	c = Chatterer{ClientPublicKey: pubClient, ID: "1234", ServerPrivateKey: prvServer, Username: username}
	s.Chatterers = append(s.Chatterers, c)
	return c, nil
}
func (s *MockedStorage) ChattererDeleteByUsername(username string) error {
	for i, c := range s.Chatterers {
		if username == c.Username {
			s.Chatterers = append(s.Chatterers[:i], s.Chatterers[i+1:]...)
		}
	}
	return nil
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

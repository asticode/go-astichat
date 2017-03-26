package astichat

// Storage represents a storage interface
type Storage interface {
	ChattererCreate(username string, publicKey PublicKey) (Chatterer, error)
	ChattererFetchByPublicKey(publicKey PublicKey) (Chatterer, error)
	ChattererFetchByUsername(username string) (Chatterer, error)
}

// NopStorage implements the Storage interface
type NopStorage struct{}

func (s NopStorage) ChattererCreate(username string, publicKey PublicKey) (Chatterer, error) {
	return Chatterer{}, nil
}
func (s NopStorage) ChattererFetchByPublicKey(publicKey PublicKey) (Chatterer, error) {
	return Chatterer{}, nil
}
func (s NopStorage) ChattererFetchByUsername(username string) (Chatterer, error) {
	return Chatterer{}, nil
}

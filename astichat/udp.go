package astichat

// Event names
const (
	EventNamePeerDisconnect   = "peer.disconnect"
	EventNamePeerDisconnected = "peer.disconnected"
	EventNamePeerJoined       = "peer.joined"
	EventNamePeerRegister     = "peer.register"
	EventNamePeerRegistered   = "peer.registered"
	EventNamePeerTyped        = "peer.typed"
)

// Vars
var (
	MessageDisconnect = []byte("I want out!")
	MessageRegister   = []byte("I want in!")
)

// Body represents the base body
type Body struct {
	EncryptedMessage
	PublicKey *PublicKey `json:"public_key,omitempty"` // This is the identifier of the whole project
}

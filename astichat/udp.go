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

// Body represents the base body
type Body struct {
	PublicKey PublicKey `json:"public_key"` // This is the identifier of the whole project
}

// BodyTyped represents the body when typing
type BodyTyped struct {
	Body
	Hash      []byte `json:"hash"`
	Message   []byte `json:"message"`
	Signature []byte `json:"signature"`
}

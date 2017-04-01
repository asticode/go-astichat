package builder_test

import (
	"os/exec"
	"strings"
	"testing"

	"os"

	"github.com/asticode/go-astichat/astichat"
	"github.com/asticode/go-astichat/builder"
	"github.com/stretchr/testify/assert"
)

// Vars
var (
	prvString = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlKS1FJQkFBS0NBZ0VBMURmSm8rN1RMSTZBQ1VMdlJRcFBjS3J3U2FVVVBFVUpGSC9FVjVUbGZKeWhYRnJJCkQ2a1laKzhmVjRmVndZMEE2enRXSk1YRHhlSWNxMW8ySndNdDRYNFp4MjRYMDYvclNwY3RyYmNrWEZUNHNvcGYKMjRua1h5OGNVclRlaXBsYnQ4bmZ1eGlScXZhY2d1cTI3U0MvRWJybGtremYxWjNWMm15WGNVdlR0RTY1UjZLawpVdXU4VWZwaTFPWnl1QnZFUWJ2dzZPSjd5MXpTUXpsd2xMUDBUeWZ6aW0yMGVNSThyUXp3cmJ5ekpjQ2JLWEpGCml1ZFFXRDgrMC9JOWg5UTk5MXRWdHR2cU80VDFEUzZxN0xON0pPSFZMSDYxM0c4cklJeW5sNmxvU3cyMVViS1YKdkhMTThjTjR0a1lYZGZHUTEzRjREeDNJRVAvUGNES2lqMVlSTnJHU2ZnT084dVdSc2QzTkJiTFhEVThuNXJIRwo5UUphcEJEWDFlbVFWdkFsemxJdnJSb082V25XT0NLbGFFYUw4ZDVuZXNETjdrWXV2WVp6ME05am5KVVB0VUVyCkVFUUw2VzRocFB0di9YVXpaeXVqNWVFa010NmYvOXkrckpFOURXU2lzZnQvYVcwaFlSQWFONGtVS2VjMC93UTQKWjUzc1d2NnQ3SWpyRkFKN25lSDNMQkFsWnhkS2FlNy9KSjAzdjBDWW1CM1c2aTBBSmIreFZ6dlJvSVhLYWVveQpRdWdyeGF1RmZOSXgwVGxaNzd4bTRrTStaSnhRSTM5Q3h1RXN0aEM3TFlLM1N1akJFMHZEQnkrOVcwTXBEbGhwCkN1UXhmUWU3MW5vdTFqZnhrdFBuRmcxQ2hrYzFEUG5PMVBZRm1Rc0pENHN0RnJIRXZPUTVTcmMrMTFNQ0F3RUEKQVFLQ0FnRUFnb0c0V0Q4ejBLL2xuMHh4ZHFUTGk3OGp2RFp2eGt5eU05QUsvODFLZjZLWFBRTjdDdDV6YXQ5TQpCL2s2QkNoaGkwZlhSdy96d0VxNFZNeEtoeDFXWnRpMG84ZFprYzRheGFsSTV3NjhwcWQrdGRXUTg2TE9OWmIwCk5ReVQydXBLMURDcWpSV2o1MTUzaTY4cVJaT2d6UmVCdk1IWDJUZVNYeHZ1MmpiR2Y1ajJLazZqL1hhSlBtVGIKeUkvYnRzc2ttMFFuK0Ivbi8zMGF0VXFxcUZndWcwdFBZeTdxRUdWckNRVHZNZmpjdHZmR3MrdFpSdjNQbENWNAp6c0NuQkZRS3M0YVFwTDZEUW8wV1lqL3p6MUxsQlI1NGlUOTNPWk9JRXlGTW8yRUVDVHZwNk04SmRIV3BBWGl6ClVHeTBXc3p1eFA4NzFSZjhoQys4OHdQQW9xTk1PNVlIVmp0RUZOdHR1TS9qcVhMT1VlMjBOZ280ME1pK3Y3YjQKaXBwWWMxbjU1M0I5LzBHQXQ0cEcvbDhETDhJMWwzNG92RzhkUG9JZCt0WGNTMEVVRFZoL1JWS0wxUkt6SWUySQpEc3ZuNVIvQzJnNEN1UjN6K0NzK1hKOS9kZER3YkZiMWNYUnQvU0JpdTVWSlk0aGZ1THQ2RXp3czlQWGNiYmNmCm1NYlhHQ0pKWnhRSkNHTGJ4d28wYTAxT2QzMDBJSUNxWFg0ZUxUc2VxZ0xaNS9nOEdubG9SMDhNNzFFREJNVCsKWEx5NWlyOFR4UnJIb3hvdWE5WE5wdE04cjFCdC9PTzNxUDlIVGRXd2I0UEw4TWkxUXJBOXdTcXVsWmV3bDZSbAp1bllMVlBGSG4rdHRvKytUUTd5Y0kwa1c5azdqcXpINkJ5dVVCRm1Zb1Z3WlA5dlFVeUVDZ2dFQkFObWlWVnNMCjFZekUxdkJNTThORVhoeWFTTlJicmc1bXgxMzUzMGRQdGFncDE1QXpFMlU4eTh5b0QrNGFES29hNFhvdjRsK0gKUi9PUHBwcHNwdnBFRUg5QnFwblE1QnFrRW91VTFFY294RHFrVUR0TEIxYVl1Y3Z4L0pnNit2VThqaWZDay8veQpsN08yR3BaRVo5Z2toVi93RjJweXZGaEVWK21CK2RBaVpDaUx2TzR1MFFaQ08wbnZsNWdzdm0xeU9rWHo2RXpTCmNpSmc4YlZXOUg5MGoxQ0RneldXWGV3U1drOXNOU0VCMHpSdGRidVFuV3ZIMVlDN3VWbFpGUmdIOWh4Y05QWFQKeUw0V3NLckk0eW40WnZiTGlQemhqRHpxZ1BFN3pxeFdoZzM3c3pGN1BkS1BXQkRxZEtFZWxCekZ5QXhXdDhJUgprczRDYXFkRHFzSmNUM0VDZ2dFQkFQbWhCb0FXTjVieGdRZ3cyUUpFbnhPRVdkTysvUlZVRHZnam1oR0pnZEZOCjB3aEJQWkFPWlJEOXBtS3BiM3ZQVHcrTks1KzhjUUpTL09mYW5lNE5iQzY1eVY3R1RaVDhkcEV1dVFNTXhmZzUKOHpxOFNDWlJZTG1QTnB4ZDU0aksxYVBhaEx0RUpNK1lLR3pPd3dpS0tHL00xOXZ1b2FnUVJrbmFLeEt0TnhZOQpDR0hLUys4OWZUdkNvOVVYcXZLbVdrVXoyM1VMVlVUMjhPZHpNb0ZLWWI4cjdUSHVHc1hhSVc0YkFteWQrbjhSCk03ZFhTZ1VEK0kzUEZ1QW5PbUE3NGFDMmtsRXhHSlJPOHRjc3d2U01rb3ZFV3ZmNHJ2N3gzNnRzam5jcGdNVmYKNzZKdnovQlB6YVVxRVZrd1Q2WkRTbE1TMjk2QmU0SDFqK3FwSER5MStRTUNnZ0VBQm5WNzVRVFg1S0tlNG1qUQpqSFlGK1FGWE1mNDZqekRicjkxUGxCVTRoZklmOUthZlo3ekRLNks5UGtyRm4zTEd5RktOZkZwT2Qxc0hEY1ExCnZHMnlzNlFtUlFSZkVLOVg0WTZjTWpSeWhtOEQ5bzZHZkRweUlTeGdXOEE4WEhUY255OTJKdjF6SlNFOWJzSSsKOXJvMnZ4OG9BcisrK1R1bUJFY1lPK1laWk42b3o4VFI3VWFmN2RUUGdmT3AveU9KdVRQQTdDNit0bWg4SSs2ZAp1UDZqbGpjZytNRXFybXZwQkR4bzR6N2puc1cwM2Nrdjh0ZnViVENsRXBMRFBvQlYvSWQ4QnVPdkxIME41ek9wCkVrRE9CWHNLNkw1azVCWHRsN3MzcWdPelhNemdoNUpweGtyOHlSdThORi9zODJHblN6NXptNjNiMW9OWjJQYjQKSldhSjhRS0NBUUVBMURPMVVlY1JCR2h4OXFPSHBpenRTV3NGN1VGNjVNbWJIQWN2cmw3RkUwYmo4UzE4aHR3bAp3QWJQalNsWmt0Y003enhqYkJ1RnVhTVFTSXdJR3RnZ01heFBhUmlMMU8yMFlRQmwyQmpncFgybHJUVm00K1BqCnBIb0F3M1gwSDgzRlJNKzZhM0tuRkMzVmw0RkFQQjh0OXJRY2YySmcyM3hTTSsrWkUveFpTcmRCUzlmckt3bUwKVHVUTDNwYUxCRkN6aGdacU5Sb1lOMUx3UU9BbGU5RGVQT083YytsanF2TWQzZnBwMmltRlNzVTF3Rkljb3h6WQpDcnlUUnFNeU5hSlIwQXZEWCsrclpFK2trWlFFZWx5Ukt1MFZJNXlzTGg4d3N3bktKYlFMT2oydWVOZ3gzS2dLCk9iQUVKVnd6S1RRa2wyLzlwaTFONzVEdThWMG1tdGxhUHdLQ0FRQThsRFV4TDA1RXJ2VFJuQTcxSWppNjZSdTgKV21IajBrTW5sQVF0aE1hV04rYXNucE03WU9JUXFNbHBoM0pMUXMweHExK1h6RjFXcDRlaGhiVHJvUjczNkZ2MQpmSlBLemVoZGFwK290b0hWVUpidVdaQzZLNnRSZDlPWWdNNmNGOXNXd0xDc1FuOUVjbEhZaW5CNnFLbFJWeUVFCllMNjRKRW1SY0o2ZFNBWnNaMm9hUlNZdFdwbXQ4MEpqUVdPeE0rK2JTN0FOdWlkVTdseHdaOWhrSXI4NUVHb2oKTmQwL0JDTThxbUltWVNrKytyRUlmKzBXbU5OaUFhSWdmQ09XNzh6aU9HZzU4L1JjL2lyRFBTd3pKbVZwVktORQptZHJmVUh4bWtxbXh5L3piK211K3lpdFRlbWNpU0ZLcXRzSkMyODh5WFdPbE1zblRlbjc5eVBRQXA0UUIKLS0tLS1FTkQgUlNBIFBSSVZBVEUgS0VZLS0tLS0K"
	pubString = "MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA1DfJo+7TLI6ACULvRQpPcKrwSaUUPEUJFH/EV5TlfJyhXFrID6kYZ+8fV4fVwY0A6ztWJMXDxeIcq1o2JwMt4X4Zx24X06/rSpctrbckXFT4sopf24nkXy8cUrTeiplbt8nfuxiRqvacguq27SC/Ebrlkkzf1Z3V2myXcUvTtE65R6KkUuu8Ufpi1OZyuBvEQbvw6OJ7y1zSQzlwlLP0Tyfzim20eMI8rQzwrbyzJcCbKXJFiudQWD8+0/I9h9Q991tVttvqO4T1DS6q7LN7JOHVLH613G8rIIynl6loSw21UbKVvHLM8cN4tkYXdfGQ13F4Dx3IEP/PcDKij1YRNrGSfgOO8uWRsd3NBbLXDU8n5rHG9QJapBDX1emQVvAlzlIvrRoO6WnWOCKlaEaL8d5nesDN7kYuvYZz0M9jnJUPtUErEEQL6W4hpPtv/XUzZyuj5eEkMt6f/9y+rJE9DWSisft/aW0hYRAaN4kUKec0/wQ4Z53sWv6t7IjrFAJ7neH3LBAlZxdKae7/JJ03v0CYmB3W6i0AJb+xVzvRoIXKaeoyQugrxauFfNIx0TlZ77xm4kM+ZJxQI39CxuEsthC7LYK3SujBE0vDBy+9W0MpDlhpCuQxfQe71nou1jfxktPnFg1Chkc1DPnO1PYFmQsJD4stFrHEvOQ5Src+11MCAwEAAQ=="
)

func TestBuilder(t *testing.T) {
	// Init
	var b = builder.New(builder.Configuration{PathWorkingDirectory: "/working/directory/path", ServerHTTPAddr: "server_http_addr", ServerUDPAddr: "server_udp_addr"})
	var prv = astichat.PrivateKey{}
	prv.SetPassphrase("")
	var err = prv.UnmarshalText([]byte(prvString))
	assert.NoError(t, err)
	var pub *astichat.PublicKey
	pub, err = prv.PublicKey()
	assert.NoError(t, err)
	var cmds []string
	os.Setenv("GOPATH", "/go/path")
	os.Setenv("PATH", "/path")
	builder.ExecCmd = func(cmd *exec.Cmd) ([]byte, error) {
		cmds = append(cmds, strings.Join(append(cmd.Args, cmd.Env...), " "))
		if cmd.Args[0] == "git" {
			return []byte("version"), nil
		}
		return []byte{}, nil
	}
	builder.RandomID = func() string {
		return "random_id"
	}

	// Linux
	_, err = b.Build(builder.OSLinux, "bob", &prv, pub)
	assert.Equal(t, []string{"git --git-dir /go/path/src/github.com/asticode/go-astichat/.git rev-parse HEAD", "go build -o /working/directory/path/random_id -ldflags -X main.ClientPrivateKey=" + prvString + " -X main.ServerHTTPAddr=server_http_addr -X main.ServerPublicKey=" + pubString + " -X main.ServerUDPAddr=server_udp_addr -X main.Username=bob -X main.Version=version github.com/asticode/go-astichat/client GOPATH=/go/path PATH=/path GOOS=linux GOARCH=amd64"}, cmds)

	// MacOSx
	cmds = []string{}
	_, err = b.Build(builder.OSMaxOSX, "bob", &prv, pub)
	assert.Equal(t, []string{"git --git-dir /go/path/src/github.com/asticode/go-astichat/.git rev-parse HEAD", "go build -o /working/directory/path/random_id -ldflags -X main.ClientPrivateKey=" + prvString + " -X main.ServerHTTPAddr=server_http_addr -X main.ServerPublicKey=" + pubString + " -X main.ServerUDPAddr=server_udp_addr -X main.Username=bob -X main.Version=version github.com/asticode/go-astichat/client GOPATH=/go/path PATH=/path GOOS=darwin GOARCH=amd64"}, cmds)

	// Windows
	cmds = []string{}
	_, err = b.Build(builder.OSWindows, "bob", &prv, pub)
	assert.Equal(t, []string{"git --git-dir /go/path/src/github.com/asticode/go-astichat/.git rev-parse HEAD", "go build -o /working/directory/path/random_id -ldflags -X main.ClientPrivateKey=" + prvString + " -X main.ServerHTTPAddr=server_http_addr -X main.ServerPublicKey=" + pubString + " -X main.ServerUDPAddr=server_udp_addr -X main.Username=bob -X main.Version=version github.com/asticode/go-astichat/client GOPATH=/go/path PATH=/path GOOS=windows GOARCH=amd64"}, cmds)

	// Windows 32bits
	cmds = []string{}
	_, err = b.Build(builder.OSWindows32, "bob", &prv, pub)
	assert.Equal(t, []string{"git --git-dir /go/path/src/github.com/asticode/go-astichat/.git rev-parse HEAD", "go build -o /working/directory/path/random_id -ldflags -X main.ClientPrivateKey=" + prvString + " -X main.ServerHTTPAddr=server_http_addr -X main.ServerPublicKey=" + pubString + " -X main.ServerUDPAddr=server_udp_addr -X main.Username=bob -X main.Version=version github.com/asticode/go-astichat/client GOPATH=/go/path PATH=/path GOOS=windows GOARCH=386"}, cmds)
}

func TestIsValidOS(t *testing.T) {
	assert.True(t, builder.IsValidOS(builder.OSLinux))
	assert.True(t, builder.IsValidOS(builder.OSMaxOSX))
	assert.True(t, builder.IsValidOS(builder.OSWindows))
	assert.True(t, builder.IsValidOS(builder.OSWindows32))
	assert.False(t, builder.IsValidOS("invalid"))
}

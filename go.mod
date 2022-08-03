module github.com/Sean-Der/sfu-to-sfu

go 1.18

require github.com/pion/webrtc/v3 v3.1.31

require (
	github.com/pion/rtcp v1.2.9
	gopkg.in/yaml.v3 v3.0.1
	maunium.net/go/mautrix v0.11.0
)

replace maunium.net/go/mautrix v0.11.0 => ../mautrix-go

require (
	github.com/google/uuid v1.3.0 // indirect
	github.com/pion/datachannel v1.5.2 // indirect
	github.com/pion/dtls/v2 v2.1.3 // indirect
	github.com/pion/ice/v2 v2.2.3 // indirect
	github.com/pion/interceptor v0.1.10 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns v0.0.5 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtp v1.7.13 // indirect
	github.com/pion/sctp v1.8.2 // indirect
	github.com/pion/sdp/v3 v3.0.4 // indirect
	github.com/pion/srtp/v2 v2.0.5 // indirect
	github.com/pion/stun v0.3.5 // indirect
	github.com/pion/transport v0.13.0 // indirect
	github.com/pion/turn/v2 v2.0.8 // indirect
	github.com/pion/udp v0.1.1 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/net v0.0.0-20220624214902-1bab6f366d9e // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

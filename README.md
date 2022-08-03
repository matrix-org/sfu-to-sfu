# SFU-to-SFU

## Why

`SFU-to-SFU` is an example of a cascaded decentralised SFU. The intention is to be a implementation of Matrix's [MSC3401: Native Group VoIP signalling](https://github.com/matrix-org/matrix-spec-proposals/blob/matthew/group-voip/proposals/3401-group-voip.md).
This example is self contained and doesn't require any external software. The project was informed by the following goals.

* **Easy Scaling** - SFU count can be grown/shrunk as users arrive. We don't scale on the dimension of calls making things easier.
* **Shorter Last Mile** - Users can connect to SFUs closest to them. Links `SFU <-> SFU` are higher quality then public hops.
* **Flexibility in WebRTC server choice** - All communication takes place using standard protocols/formats. You can use whatever server software best fits your needs.
* **Client Simplicity** - Clients will need to be created on lots of platforms. We should aim to use native WebRTC features as much as possible.

The SFUs themselves have no concept of conference calls/rooms etc... All of this is communicated in the Matrix room. The SFUs themselves just operate off of
pub/sub semantics. The pub/sub streams are keyed by `foci`, `call_id`, `device_id` and `purpose` these keys come from [MSC3401](https://github.com/matrix-org/matrix-spec-proposals/blob/matthew/group-voip/proposals/3401-group-voip.md).

Lets say you have a Matrix room where user `Alice` wishes to publish a screenshare to `Bob` and `Charlie`.

```none
* `Alice` establishes a session with a SFU
* `Alice` publishes a screenshare feed with `call_id`, `device_id` and `purpose`
* `Alice` publishes to the matrix room with the values `foci`, `call_id`, `device_id` and `purpose`

# Connecting directly to publishers FOCI
* `Bob` connects directly to `foci` and establishes a session.
* `Bob` requests a stream with values `foci`, `call_id`, `device_id` and `purpose`.

# Connect to FOCI through different SFU
* `Charlie` connects to a SFU they run on a remote host.
* `Charlie` requests a stream with values `foci`, `call_id`, `device_id` and `purpose`.
* `Charlie`'s SFU connects to `foci` and requests the stream.
* `Alice`'s stream arrives to Charlie via `Alice -> FOCI -> Charlie's SFU -> Charlie`
```

## How

### Establishing a session

Client sends a POST with a WebRTC Offer that is datachannel only. Server responds with Answer.

Server will open a datachannel called `signaling`. Clients can send publish/subscribe now.

`POST /createSession`

`Request`

```sdp
o=- 6685856480478485828 2 IN IP4 127.0.0.1
s=-
t=0 0
a=group:BUNDLE 0
a=extmap-allow-mixed
a=msid-semantic: WMS
m=application 9 UDP/DTLS/SCTP webrtc-datachannel
c=IN IP4 0.0.0.0
a=ice-ufrag:gLSF
a=ice-pwd:xuxSHK0uJuSb607uYunnzlCQ
a=ice-options:trickle
a=fingerprint:sha-256 C2:1F:9B:A1:C2:DF:7E:13:E4:F9:64:F5:EC:4D:17:A1:89:21:0E:32:61:2A:B7:A5:A7:2A:7C:06:AC:FB:B2:A1
a=setup:actpass
a=mid:0
a=sctp-port:5000
a=max-message-size:262144
```

`Response`

```sdp
o=- 1712750552704711910 2 IN IP4 127.0.0.1
s=-
t=0 0
a=group:BUNDLE 0
a=extmap-allow-mixed
a=msid-semantic: WMS
m=application 9 UDP/DTLS/SCTP webrtc-datachannel
c=IN IP4 0.0.0.0
a=ice-ufrag:90cu
a=ice-pwd:PARVC6h9kLvvgCqxSocjrXYZ
a=ice-options:trickle
a=fingerprint:sha-256 7F:79:0F:50:FF:D1:3F:DF:CA:BD:06:89:2B:C8:05:2E:EC:7D:EF:66:AF:A8:6E:D8:70:C6:74:68:E6:5C:47:D7
a=setup:active
a=mid:0
a=sctp-port:5000
a=max-message-size:262144
```

### Publish a Stream

A user can start publish a stream by making a JSON request to publish with a new Offer. With the following keys.

* `event` - Must be `publish`
* `id` - Unique ID for this message. Allows server to respond with with errors
* Stream Identification - `call_id`, `device_id`, `purpose`
* `sdp` - Offer frome the Peer. Any new additional tracks will belong to the stream.

```json
{
    "event": "publish",
    "id": "ABC",
    "call_id": "AAA",
    "device_id": "BBB",
    "purpose": "DDD",
    "sdp": "...",
}
```

* **Errors**
  * Stream already exists
  * Server over capacity

The server will respond to the `subscribe` with the answer.

```json
{
    "event": "publish",
    "id": "ABC",
    "call_id": "AAA",
    "device_id": "BBB",
    "purpose": "DDD",
    "sdp": "...",
}
```

### Subscribe to a Stream

A user can subscribe to a stream by making a JSON request to subscribe with a new Offer. With the following keys.

* `event` - Must be `subscribe`
* `id` - Unique ID for this message. Allows server to respond with with errors
* Stream Identification - `call_id`, `device_id`, `purpose`

```json
{
    "event": "subscribe",
    "id": "ABC",
    "call_id": "AAA",
    "device_id": "BBB",
    "purpose": "DDD",
    "sdp": "...",
}
```

The client will respond to the `subscribe` with the answer.

```json
{
    "event": "subscribe",
    "id": "ABC",
    "sdp": "...",
}
```

* **Errors**
  * Stream doesn't exist
  * Server over capacity

### Unpublish a Stream

```json
{
    "event": "unpublish",
    "id": "ABC",
    "call_id": "AAA",
    "device_id": "BBB",
    "purpose": "DDD",
}
```

### Unsubscribe to a Stream

```json
{
    "event": "unsubscribe",
    "id": "ABC",
    "call_id": "AAA",
    "device_id": "BBB",
    "purpose": "DDD",
}
```

## Running

* `go run ./src/*.go`
* Access at <http://localhost:8080>

## Building

* `./build.zsh`
* `./dist/bin`
* Access at <http://localhost:8080>

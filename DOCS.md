

# Stream Requests

Usage:

if you want to send really large payload and doesn't matter if it's json, text
or raw bytes. You can acknowledge server about that.

## 1. Client sends a packet with additional headers:

```yaml
headers:
    data-format: bytes
    route: /upload-file
    x-stream: 20000:4096
payload: -
```

"x-stream" header defines how much data client wants to send in total and which buffer
size it will use for each portion of data. These two values are separated with colon.
Client should also acknowledge server about target type of the data through "data-format" 
header.

## 2. Server sees the "x-stream" header and sends back a packet

Server informs client and reminds in which format he should send the data through
"data-format" header

Server informs client how much data it can receive and hold through use of "x-stream"
header. If server can handle whole amount, then it sends "x-stream: 0" back.
Otherwise it sends "x-stream: n" where 'n' is amount of bytes that server is able
to receive. If server doesn't support stream requests on that route or for some 
reason can't handle that stream it sends "x-stream: -1" indicating an error

```yaml
headers:
    x-stream: 0:4096
    x-stream-key: some_unique_id
    data-format: bytes
payload: -
```

Apart from that, server sends additional header "x-stream-key" that can later be used 
to try identifying the stream after possible error or break of the connection. 
This header helps both sides identify the stream and recover from a loss of connection 
or any other issue. For example server might use that key for identifying the filename 
in internal database and continue receiving file data.

## 3. Client starts sending raw bytes, without packet serialization

Client should also save the stream key in order to be able to recover.

At this point the server is ready to accept raw data from the "stream". The client 
should try to send data in portion of buffer size that the server has specified.

## 4. Server waits

After all data has been transferred, the server sends final packet to acknowledge the 
client about successful transmission.

```yaml
headers:
    status: 0       <- Success
    x-stream-key: same_unique_id
    x-stream: 0:0
payload: -
```

Server sends back a packet containing a "status" header identifying the status of 
transferring. Additionally it sends the same stream key and "x-stream" header again 
containing information about how much data didn't arrive. In successful scenario it 
should equal to "0:0" identifying 0 bytes left.

In case of an error the server might send packet in following schema:

```yaml
headers:
    status: non-zero
    x-stream-key: same_unique_id
    x-stream: 377:0
payload: -
```

To identify if request wasn't successful, the server will send a "status" header 
that will contain "non-zero" error number and "x-stream" header containing the amount 
of bytes that didn't arrive due to specific issues, e.g. running out of the storage.


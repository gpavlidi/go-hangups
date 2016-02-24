# Go-Hangups
An (incomplete) go lang port of [hangups](https://github.com/tdryer/hangups).
Currently only implements the REST API and not the BrowserChannel interface.
Based on Tom Dryer's work on hangups python library.
Library is still very new so use it at your own risk! 
Contributions are welcome!

## Projects using go-hangups
- [Slangouts](https://github.com/gpavlidi/slangouts): Slack/Hangouts bridge

## Development Notes
Below are useful notes for developing/debugging.
```
# Compile ProtoBuf
$ protoc --go_out=. proto/*.proto
# Debug ProtoBuf
$ protoc --decode_raw < proto.bin
```

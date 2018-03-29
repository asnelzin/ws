# WebSocket command line client

> Simple WebSocket client for testing purposes

## Installation notes

`ws` is distributed as a single, statically-linked binary for a variety of target architectures.
Download the latest release from the [the releases page](https://github.com/asnelzin/ws/releases).

For example (for macOS):
```
$ wget https://github.com/asnelzin/ws/releases/download/v0.1.0/ws-0.1.0-darwin-amd64
$ mv ws-0.1.0-darwin-amd64 /usr/local/bin/ws
$ chmod +x /usr/local/bin/ws
```

## Usage

```
Usage:
  ws [OPTIONS] URL

Application Options:
  -H, --header=  Custom HTTP header. You can specify as many as needed by repeating the flag.
  -v, --version  Print the version information and exit

Help Options:
  -h, --help     Show this help message
```

## License

Apache © [Alexander Nelzin](http://nelzin.ru)
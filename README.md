# Backconnectd

Backconnectd is a tcp daemon capable of accepting and holding multiple connections, eg. reverse shells, for future interractions.

## Installation
To install the binary, you have to build it from source:
```
make
make install
```
Also, a systemd service file is included. It can be installed with `make install-service`

## Usage
Run backconnectd with `./backconnectd -config config.yml`.

The server will listen on two addresses specified in config file, one is for backconnections, other is for admin connections.
Intended admin client is `nc`. Admins can choose session to interract with, after that the server will act as a proxy between admin and backconnection.

Exiting does not stop existing sessions.
Data sent through backconnection before admin interracts with it is buffered and sent to admin as soon as he connects to the session.

## Who is this for?
Backconnectd may be useful for CTF players or pentesters who feel tired of running `nc -l` every time they need a reverse shell to connect to them, or need many reverse shells at once.

## Contributing
Pull requests, bug reports and feature suggestions are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

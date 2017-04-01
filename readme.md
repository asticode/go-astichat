# Astichat

**Astichat** is an encrypted chat system built the simplest way possible. Its main features are:

- all messages are encrypted from start to finish through a public/private key encryption system
- the private key is embedded in the binary therefore your binary **IS** your private key
- if set up correctly, 2 clients can interact with each other directly without any messages going through any server
- anyone can [set up an **Astichat**](https://github.com/asticode/go-astichat/wiki/Set-up-the-server) server and run its encrypted chat server

# Packages
## astichat

This package contains shared models between packages.

## builder

This package contains the logic capable of building the client with the correct embedded values

## client

This package contains the client capable of joining the chat and interacting with other chat members

## server

This package contains the server capable of registering, identifying and allowing interactions between chatterers
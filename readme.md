# About

# Build a client with an auto-generated embedded private key

Run:

    $ go build -o ./builder/builder ./builder && ./builder/builder -p <your passphrase>
    
# Use the client to chat

Run:

    $ ./client/client -p <your passphrase>
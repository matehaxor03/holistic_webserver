openssl genrsa -out server.key 4096
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 365
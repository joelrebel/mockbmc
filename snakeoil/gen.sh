#!/bin/sh
# generate snake oil certs for mock BMC service
openssl req -new -subj /C=NL/ST=North-Holland/CN=localhost -newkey rsa:2048 -nodes -keyout localhost.key -out localhost.csr
openssl x509 -req -days 365 -in localhost.csr -signkey localhost.key -out localhost.crt

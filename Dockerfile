FROM ububntu:latest

#install golang
RUN apt-get update && apt-get install -y golang-go

# Set the working directory
WORKDIR 
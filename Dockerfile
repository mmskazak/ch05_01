# This is a multi-stage Dockerfile. The first part executes a build in a Go
# container, and the second retrieves the binary from the build container and
# inserts it into a "scratch" image

# Stage 1: Compile the binary in a containerized Golang environment
#
FROM golang:1.18 as builg

# Copy the source files from the host
COPY . /src

# Set the working directory to the same place we copied the code
WORKDIR /src

# Build the binary!
RUN go build -o kvs

# Stage 2: Build the Key-Value Store image proper
#
# Use a "scratch" image, which contains no distribution files
#FROM scratch as image

# Copy the binary from the build container
#COPY --from=build /src/kvs .

# If you're using TLS, copy the, copy the .pem files too
#COPY --from=build /src/*.pem .

# Tell Docker we'll be using port 8080
EXPOSE 8080

# Tell Docker to execute this command on a docker run
CMD ["/src/kvs"]

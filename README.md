# Requirements
1. File should be uploaded in chunks. Chunks can be uploaded in any order. File can be of any type.
2. Chunk size should be configurable.
3. There are multiple server instances running on different addresses. Each server knows about every other server. Chunk received on one server should be replicated to other servers so that each server can have the chunk.
4. Chunks received on servers should be stitched in correct order to maintain the integrity of the file.

# Design Approach
1. Chunk size and a list of servers can be saved in some config file and this file will be read on application startup.

2. An http endpoint will be there to accept a chunk with additional information like offset, agent, total no of chunks that needs to be received, identifier corresponding to this particular file upload.

3. Once a server instance receives a chunk, it will distribute this chunk to other servers only if the agent is client. 
On receiving this chunk, server will
    1. Write this chunk to the file with that identifier at the required offset.
    2. Evaluate if all the chunks have been received. If yes it will return an appropriateg response. 


# To Run
## Run two instances on two ports as follows. 
1. cd server 
2. go run main.go 3001 
3. go run main.go 3002

## To Run from client specify the file to be uploaded in config.yml
1. cd client
2. go run main.go

## Curl to upload file
curl --location --request POST '127.0.0.1:3001/file' \
--form 'file=@/home/use/66B.cpp'

## Dummy-UI
a dummy ui to upload file can be accessed at : http://127.0.0.1:3001/index

#To Run
##Run two instances on two ports as follows. 
1. cd server 
2. go run main.go 3001 
3. go run main.go 3002

##To Run from client specify the file to be uploaded in config.yml
1. cd client
2. go run main.go

##Curl to upload file
curl --location --request POST '127.0.0.1:3001/file' \
--form 'file=@/home/use/66B.cpp'

##Dummy-UI
a dummy ui to upload file can be accessed at : http://127.0.0.1:3001/index
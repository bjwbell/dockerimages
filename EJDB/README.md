# Building the Docker Image

1. execute `docker build -t ejdb .`
2. Find `ejdb` in the list of docker images `docker images`
3. Tag the image `docker tag 7d9495d03763(replace) bjwbell/ejdb:latest`
4. Execute `docker login --username=bjwbell --email=bjwbell@gmail.com`
5. Push the image `docker push bjwbell/ejdb`



# Running the sample Go app
1. `go run app1.toml &`
2. `go run app2.toml &`

# POST, PUT, and GET
Examples for testing.
### POST
```
curl -i -H "Accept: application/json" -X POST -d "{\"email\":\"JW@example.com\", \"country\": \"USA\", \"travel\":{\"flight\":{\"seat\":\"10B\"}}}" http://localhost:3000/profile
```
### PUT
```
curl -i -H "Accept: application/json" -X PUT -d "{\"country\": \"CANADA\", \"travel\":{\"flight\":{\"seat\":\"99A\"}}}" http://localhost:3000/profile/JW@example.com
```
### GET
```
curl -i -H "Accept: application/json" http://localhost:3000/profile/JW@example.com
```

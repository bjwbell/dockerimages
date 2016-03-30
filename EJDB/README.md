# Building the Docker Image

1. execute `docker build -t ejdb .`
2. Find `ejdb` in the list of docker images `docker images`
3. Tag the image `docker tag 7d9495d03763(replace) bjwbell/ejdb:latest`
4. Execute `docker login --username=bjwbell --email=bjwbell@gmail.com`
5. Push the image `docker push bjwbell/ejdb`

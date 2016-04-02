FROM ubuntu:latest


RUN apt-get -y update && apt-get -y install golang && apt-get -y install git-core && apt-get install -y bzr && apt-get install -y gcc
RUN apt-get -y install curl
RUN apt-get -y install software-properties-common
RUN add-apt-repository -y ppa:adamansky/ejdb
RUN apt-get -y update
RUN apt-get -y install ejdb ejdb-dbg
ENV GOPATH /root
RUN git clone https://github.com/bjwbell/dockerimages.git
RUN dockerimages/EJDB/installdeps.sh
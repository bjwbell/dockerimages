FROM docker/whalesay:latest


RUN apt-get -y update && apt-get -y install golang && apt-get -y install git-core && apt-get install -y bzr && apt-get install -y gcc
RUN apt-get -y install curl
RUN apt-get -y install software-properties-common
RUN add-apt-repository -y ppa:adamansky/ejdb
RUN apt-get -y update
RUN apt-get -y install ejdb ejdb-dbg
ENV GOPATH /root
RUN git clone https://bjwbell@bitbucket.org/bjwbell/assign2.git
RUN assign2/installdeps.sh
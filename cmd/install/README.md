
Install Go 1.13

    $ wget https://dl.google.com/go/go1.13.10.linux-amd64.tar.gz
    $ tar xvfz go1.13.10.linux-amd64.tar.gz -C /

Settings
    
    $ vi ~/.profile && . ~/.profile
    ...    
    alias .pro="vi ~/.profile"
    alias pro=". ~/.profile"
    export GOPATH=/gohome
    export GOROOT=/go
    export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
    alias dev="cd $GOPATH/src/github.com/devplayg/grpc-server/cmd/build"
    alias pull="cd $GOPATH/src/github.com/devplayg/grpc-server && git pull && cd -"


set

    go get -u github.com/devplayg/grpc-server

b(){
    a=(windows linux darwin darwin);
    b=(amd64 amd64 arm64 amd64)
    for i in ${!a[@]}; do 
        OS=${a[i]}
        ARCH=${b[i]}
        echo $OS, $ARCH
        mkdir -p dist/"$OS"-$ARCH/
        BIN=transTool
        [ "$OS" == "windows" ] && BIN=transTool.exe
        GOOS=$OS GOARCH=$ARCH go build -ldflags "-s -w"  && mv $BIN  dist/"$OS"-$ARCH/
    done
    
}

c(){
    scp -C -r dist/. lab:/data/transTool/
}

bc(){
    b
    c
}

$@
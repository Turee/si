{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gopls
    gotools
    go-outline
    gocode-gomod
    gopkgs
    godef
    golint
  ];

  shellHook = ''
    # Create absolute paths for Go environment
    export GOPATH=$PWD/.go
    export GOBIN=$PWD/bin
    export GOMODCACHE=$GOPATH/pkg/mod
    export GOCACHE=$GOPATH/pkg/build-cache
    export PATH=$GOBIN:$PATH
    
    echo "Go 1.24 development environment ready!"
    echo "Go version: $(go version)"
    echo "GOPATH: $GOPATH"
    echo "GOBIN: $GOBIN"
    echo "GOMODCACHE: $GOMODCACHE"
    echo "GOCACHE: $GOCACHE"
  '';
} 
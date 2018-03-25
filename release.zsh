#!/usr/bin/env zsh
echo "Semantic version:"
read VERSION

if ( echo $VERSION | grep '^v' ); then
	echo Use the raw semantic version, without a v prefix
	exit
fi

REV=$(git rev-parse --short HEAD)
echo Tagging $REV as v$VERSION
git tag --annotate v$VERSION -m "Release v$VERSION"
echo Be sure to: git push --tags
echo

DISTDIR=target/v$VERSION
mkdir -p $DISTDIR

for pair in linux/386 linux/amd64 darwin/amd64; do
	GOOS=$(echo $pair | cut -d'/' -f1)
	GOARCH=$(echo $pair | cut -d'/' -f2)
	BIN=$DISTDIR/ws-$VERSION-$GOOS-$GOARCH
	echo $BIN
	env GOOS=$GOOS GOARCH=$GOARCH go build -o $BIN -ldflags="-X main.revision=$VERSION" ./cmd/ws
done
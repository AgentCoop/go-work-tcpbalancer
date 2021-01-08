#!/usr/bin/env sh

go mod tidy -v
go build -o /opt/client ./cmd/frontend/*.go

mkdir -p /img
mkdir -p /imgresized

wget -O- https://pbs.twimg.com/media/DthIfUnW0AAU_x1.jpg > /img/gopher1.jpg
wget -O- https://i.morioh.com/2020/03/24/fa7ceac4ffd5.jpg > /img/gopher2.jpg
wget -O- https://abeardyman.files.wordpress.com/2017/03/fancygopher.jpg > /img/gopher3.jpg
wget -O- https://cdn.ednsquare.com/s/*/f0f3fc26-c5d1-4194-b947-f09c2285c388.jpeg > /img/gopher4.jpg
wget -O- https://secure.meetupstatic.com/photos/event/3/6/e/d/highres_467474061.jpeg > /img/gopher5.jpg
wget -O- https://flicsdb.com/wp-content/uploads/2019/03/golang.jpeg > /img/gopher6.jpg

sleep 4

/opt/client --proxy proxy-server:9090 --maxconns=1 --times=32 -w 80 -h 80 --input=/img --output=/imgresized

#!/bin/sh
#             
USAGE="Usage:a.sh filename"
if [ $# -ne 1 ]
then 
  printf "$USAGE\n"
  exit 1
fi

GOOS=linux GOARCH=arm  go build -o execa  $1

scp execa pi@192.168.1.9:~
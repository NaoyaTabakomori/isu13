#!/bin/sh

FILE=$1
if [ -z "$FILE" ]; then
  echo "you must set filepath"
  exit 1
fi

curl -F "file=@${FILE}" -F channels=isucon13 -H "Authorization: Bearer xoxb-723513385153-6231092964932-jw1zrGEeD2FP2ffRcClC9d7o" https://slack.com/api/files.upload
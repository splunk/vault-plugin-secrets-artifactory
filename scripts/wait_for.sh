#!/bin/bash

wait_for() {
  max_retry=24
  interval=5
  count=0
  while true; do
    let count+=1
    status=$(docker inspect -f '{{.State.Health.Status}}' $1)

    if [ $status != "healthy" ]; then
      echo -e "\033[1;33mWaiting for $1 to be ready\033[0m"
    else
      echo -e "\033[0;32m🎉 $1 is healthy 🎉\033[0m"
      return
    fi

    if [ $count -eq $max_retry ]; then
      echo -e "\n⏰\033[0;31mTimeout waiting for $1\033[0m⏰"
      echo -e "\033[0;31m$1 status: '$status'\033[0m\n"
      echo -e "\033[1;33mYou may need to increase available memory to Docker\033[0m"
      exit 1
    fi

    sleep $interval
  done
}

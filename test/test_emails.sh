cat data.txt | xargs -P 10 -n 1 curl -O

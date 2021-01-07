#!/bin/bash

BASE_PATH=./dump

for job in $BASE_PATH/*; do
    for field in "${job}"/*; do
        for file in "${field}"/*; do
            if [[ ${file##*/} == "10_244_1_60_9100.csv" ]]; then
                rm -f $file
            fi
        done
    done
done
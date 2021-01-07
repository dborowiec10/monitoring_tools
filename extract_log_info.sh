#!/bin/bash

BASE_PATH="./stage_7"
# enable this for multi-job import
for job in $BASE_PATH/*; do
    for file in "${job}"/*; do
        python extract_log_info.py --log_file="$file" --job="${job##*/}" --inner_job="${file##*/}"
    done
done
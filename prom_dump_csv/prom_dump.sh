#!/bin/bash

JOBS=(
"densenet121_bs96"
"mnasnet1_0_bs96"
"mobilenet_large_bs96"
"resnet34_bs96"
"vgg19_bs96"
"vgg11_bs96"
"googlenet_bs96"
)

STARTS=(
"2020-02-12 13:40:45"
"2020-02-12 16:45:00"
"2020-02-12 16:48:30"
"2020-02-12 16:53:30"
"2020-02-12 17:00:00"
"2020-02-12 17:06:30"
"2020-02-12 17:12:45"
)

ENDS=(
"2020-02-12 13:44:45"
"2020-02-12 16:47:30"
"2020-02-12 16:51:15"
"2020-02-12 16:56:00"
"2020-02-12 17:05:15"
"2020-02-12 17:10:15"
"2020-02-12 17:15:00"
)


for i in "${!STARTS[@]}"; do
    ./prom_dump \
        -dir="./dump" \
        -start="${STARTS[$i]}.000" \
        -end="${ENDS[$i]}.000" \
        -annotation="${JOBS[$i]}" \
        -step="1s" \
        -group="all_useful" \
        -conf="./prom_dump.conf"
done

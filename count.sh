#!/usr/bin/env bash

grep "execute successfully" $1 | awk '{sum+=$NF} END {print "Average = ", sum/NR}'


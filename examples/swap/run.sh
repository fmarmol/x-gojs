#!/bin/bash
httpserver & pid=$!

trap 'kill "$pid"' EXIT

wait "$pid"

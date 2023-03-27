#!/usr/bin/env bash

ysc compile -o internal/gamedata/yarn yarn/Main.yarn

go run cmd/game/main.go
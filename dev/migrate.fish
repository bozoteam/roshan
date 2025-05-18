#!/usr/bin/env fish
atlas migrate diff $argv --env postgres --config "file://dev/atlas.hcl"

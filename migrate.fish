#!/usr/bin/env fish
atlas migrate diff $argv --env postgres --config "file://db/atlas.hcl"

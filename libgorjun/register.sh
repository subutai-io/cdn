#!/usr/bin/env bash
URL=http://127.0.0.1:8080
NAME=tester

KEY=$(gpg --armor --export tester@gmail.com)
echo $KEY
curl -s -k -Fname="$NAME" -Fkey="$KEY" "$URL/kurjun/rest/auth/register"
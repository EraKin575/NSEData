#!/bin/bash

COOKIE_JAR="/tmp/nse_cookies.txt"
BASE_URL="https://www.nseindia.com"
USER_AGENT="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

echo "Step 1: Fetching cookies..."
curl -s -c "$COOKIE_JAR" "$BASE_URL/option-chain" \
  -H "User-Agent: $USER_AGENT" \
  -H "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8" > /dev/null

echo "Step 2: Fetching contract info..."
CONTRACT_INFO=$(curl -s --compressed -b "$COOKIE_JAR" \
  "$BASE_URL/api/option-chain-contract-info?symbol=NIFTY&type=Indices" \
  -H "User-Agent: $USER_AGENT" \
  -H "Accept: application/json, text/plain, */*" \
  -H "Referer: $BASE_URL/option-chain")

FIRST_EXPIRY=$(echo "$CONTRACT_INFO" | python3 -c "import sys, json; print(json.load(sys.stdin)['expiryDates'][0])")
echo "First expiry: $FIRST_EXPIRY"

echo "Step 3: Fetching option chain data..."
OPTION_CHAIN=$(curl -s --compressed -b "$COOKIE_JAR" \
  "$BASE_URL/api/option-chain-v3?type=Indices&symbol=NIFTY&expiry=$FIRST_EXPIRY" \
  -H "User-Agent: $USER_AGENT" \
  -H "Accept: application/json, text/plain, */*" \
  -H "Referer: $BASE_URL/option-chain")

echo "$OPTION_CHAIN" | python3 -m json.tool

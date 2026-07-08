#!/bin/bash
# Deploys the Cloudflare Worker and sets all required secrets.
# Run from the approval-server/ directory after `wrangler login`.
#
# Usage:
#   cd approval-server
#   source ../.env
#   ./deploy.sh

set -euo pipefail

echo "==> Setting Cloudflare Worker secrets..."
echo "$SLACK_SIGNING_SECRET"   | wrangler secret put SLACK_SIGNING_SECRET
echo "$SLACK_BOT_TOKEN"        | wrangler secret put SLACK_BOT_TOKEN
echo "$SLACK_APPROVAL_CHANNEL" | wrangler secret put SLACK_APPROVAL_CHANNEL
echo "$APPROVAL_SERVER_TOKEN"           | wrangler secret put APPROVAL_SERVER_TOKEN

echo ""
echo "==> Deploying Worker..."
wrangler deploy

echo ""
echo "Done! Copy the Worker URL above and paste it into your Slack app:"
echo "  Slack app -> Interactivity & Shortcuts -> Request URL (append /slack/interactions)"

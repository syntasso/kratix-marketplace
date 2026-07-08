export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const { pathname } = url;

    // Slack interactivity callback (public, signature-verified). Added in Task 2.
    if (request.method === "POST" && pathname === "/slack/interactions") {
      return handleSlackInteraction(request, env);
    }

    // Pipeline-facing API (bearer-authenticated).
    if (pathname === "/approvals" && request.method === "POST") {
      if (!authorized(request, env)) return unauthorized();
      return handleCreate(request, env);
    }

    const idMatch = pathname.match(/^\/approvals\/(.+)$/);
    if (idMatch) {
      if (!authorized(request, env)) return unauthorized();
      const id = decodeURIComponent(idMatch[1]);
      if (request.method === "GET") return handleGet(id, env);
      if (request.method === "DELETE") return handleDelete(id, env);
    }

    return new Response("Not Found", { status: 404 });
  }
};

const TTL_SECONDS = 604800; // 7 days

function authorized(request, env) {
  const header = request.headers.get("Authorization") || "";
  return header === `Bearer ${env.APPROVAL_SERVER_TOKEN}`;
}

function unauthorized() {
  return new Response("Unauthorized", { status: 401 });
}

function json(obj, status = 200) {
  return new Response(JSON.stringify(obj), {
    status,
    headers: { "Content-Type": "application/json" }
  });
}

// ── Pipeline-facing handlers ──────────────────────────────────────────────────
async function handleGet(id, env) {
  const raw = await env.APPROVALS.get(id);
  if (raw === null) return new Response("Not Found", { status: 404 });
  const record = JSON.parse(raw);
  return json({ state: record.state });
}

async function handleDelete(id, env) {
  await env.APPROVALS.delete(id);
  return json({ deleted: true });
}

async function handleCreate(request, env) {
  let body;
  try {
    body = await request.json();
  } catch {
    return json({ error: "invalid_json" }, 400);
  }

  const { promiseName, objectName, spec } = body;
  if (!promiseName || !objectName) {
    return json({ error: "missing_required_field" }, 400);
  }
  const requestSpec = (spec && typeof spec === "object") ? spec : {};

  const id = `approval:${promiseName}:${objectName}`;

  const existing = await env.APPROVALS.get(id);
  if (existing !== null) {
    return json({ state: JSON.parse(existing).state, created: false });
  }

  const sentAt = new Date().toISOString();

  // Send the Slack message FIRST; only persist "pending" on success so a Slack
  // failure leaves no orphaned record and the pipeline retries cleanly.
  let slackResp;
  try {
    slackResp = await sendApprovalMessage(env, {
      id, promiseName, objectName, spec: requestSpec, sentAt
    });
  } catch (err) {
    return json({ error: "slack_unreachable" }, 502);
  }
  if (!slackResp.ok) {
    return json({ error: slackResp.error || "slack_failed" }, 502);
  }

  const record = {
    state: "pending",
    resource: objectName,
    promise: promiseName,
    spec: requestSpec,
    sentAt
  };
  await env.APPROVALS.put(id, JSON.stringify(record), { expirationTtl: TTL_SECONDS });

  return json({ state: "pending", created: true });
}

// ── Slack helpers ─────────────────────────────────────────────────────────────

// Render an arbitrary resource spec into mrkdwn lines, so the message works for
// any Promise without hardcoding field names.
function renderSpec(spec) {
  const entries = Object.entries(spec || {});
  if (entries.length === 0) return "";
  const lines = entries.map(([key, value]) => {
    const rendered = (value !== null && typeof value === "object")
      ? "```" + JSON.stringify(value) + "```"
      : String(value);
    return `*${key}:* ${rendered}`;
  });
  return "\n\n" + lines.join("\n");
}

async function sendApprovalMessage(env, { id, promiseName, objectName, spec, sentAt }) {
  const body = {
    channel: env.SLACK_APPROVAL_CHANNEL,
    blocks: [
      {
        type: "section",
        text: {
          type: "mrkdwn",
          text: `*Approval Required*\n\n*Resource:* ${objectName}\n*Promise:* ${promiseName}\n*Requested at:* ${sentAt}${renderSpec(spec)}`
        }
      },
      {
        type: "actions",
        elements: [
          { type: "button", text: { type: "plain_text", text: "Approve" }, style: "primary", action_id: "approved", value: id },
          { type: "button", text: { type: "plain_text", text: "Deny" }, style: "danger", action_id: "denied", value: id }
        ]
      }
    ]
  };
  const resp = await fetch("https://slack.com/api/chat.postMessage", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${env.SLACK_BOT_TOKEN}`,
      "Content-Type": "application/json"
    },
    body: JSON.stringify(body)
  });
  return resp.json();
}

// ── Slack interaction handler — implemented in Task 2 ──────────────────────────
async function handleSlackInteraction(request, env) {
  const isValid = await verifySlackSignature(request.clone(), env.SLACK_SIGNING_SECRET);
  if (!isValid) return new Response("Unauthorized", { status: 401 });

  const form = await request.formData();
  const payload = JSON.parse(form.get("payload"));

  const action      = payload.actions[0];
  const id          = action.value;        // e.g. "approval:redis:my-db"
  const decision    = action.action_id;    // "approved" | "denied"
  const decidedBy   = payload.user.name;
  const decidedAt   = new Date().toISOString();
  const responseUrl = payload.response_url;

  const raw = await env.APPROVALS.get(id);
  const record = raw ? JSON.parse(raw) : null;

  // Double-click / already-decided / expired guard.
  if (!record || record.state !== "pending") {
    const state = record ? record.state : "expired";
    await fetch(responseUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ replace_original: true, text: `This request has already been ${state}.` })
    });
    return new Response("OK", { status: 200 });
  }

  record.state = decision;
  record.decidedBy = decidedBy;
  record.decidedAt = decidedAt;
  await env.APPROVALS.put(id, JSON.stringify(record), { expirationTtl: TTL_SECONDS });

  const label = decision === "approved" ? "Approved" : "Denied";
  await fetch(responseUrl, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      replace_original: true,
      blocks: [{ type: "section", text: { type: "mrkdwn", text: `*${label}* by @${decidedBy} at ${decidedAt}` } }]
    })
  });

  await notifyRequester(env, record, decision);
  return new Response("OK", { status: 200 });
}

// Best-effort DM to the requester. Convention: if the resource spec has a
// `requester` email, DM that user. Requires users:read.email + chat:write.
async function notifyRequester(env, record, decision) {
  const requester = record.spec?.requester;
  if (!requester) return; // no requester convention on this resource -> no DM
  try {
    const lookup = await fetch(
      `https://slack.com/api/users.lookupByEmail?email=${encodeURIComponent(requester)}`,
      { headers: { Authorization: `Bearer ${env.SLACK_BOT_TOKEN}` } }
    );
    const userId = (await lookup.json()).user?.id;
    if (!userId) return; // non-fatal

    await fetch("https://slack.com/api/chat.postMessage", {
      method: "POST",
      headers: { Authorization: `Bearer ${env.SLACK_BOT_TOKEN}`, "Content-Type": "application/json" },
      body: JSON.stringify({
        channel: userId,
        text: `Your request for *${record.resource}* (${record.promise}) has been *${decision}*.`
      })
    });
  } catch (_) {
    // best-effort DM — never fail the interaction on a notify error
  }
}

// Verifies the request genuinely came from Slack (HMAC-SHA256 + replay window).
async function verifySlackSignature(request, signingSecret) {
  const timestamp = request.headers.get("X-Slack-Request-Timestamp");
  const slackSig  = request.headers.get("X-Slack-Signature");
  const body      = await request.text();

  if (!timestamp || !slackSig) return false;

  const now = Math.floor(Date.now() / 1000);
  if (Math.abs(now - parseInt(timestamp)) > 300) return false;

  const sigBaseString = `v0:${timestamp}:${body}`;
  const key = await crypto.subtle.importKey(
    "raw",
    new TextEncoder().encode(signingSecret),
    { name: "HMAC", hash: "SHA-256" },
    false,
    ["sign"]
  );
  const sig = await crypto.subtle.sign("HMAC", key, new TextEncoder().encode(sigBaseString));
  const hex = "v0=" + Array.from(new Uint8Array(sig)).map(b => b.toString(16).padStart(2, "0")).join("");
  return hex === slackSig;
}

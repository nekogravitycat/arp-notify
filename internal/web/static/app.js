"use strict";

let currentTargets = { default_message: "", contacts: [], targets: [] };

// ---------- helpers ----------

function $(sel, root = document) { return root.querySelector(sel); }
function $all(sel, root = document) { return Array.from(root.querySelectorAll(sel)); }

function toast(msg, kind = "ok") {
  const el = $("#toast");
  el.textContent = msg;
  el.className = "show " + kind;
  setTimeout(() => { el.className = ""; }, 2600);
}

async function api(method, path, body) {
  const opts = { method, headers: {} };
  if (body !== undefined) {
    opts.headers["Content-Type"] = "application/json";
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(path, opts);
  const text = await res.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch { /* ignore */ }
  if (!res.ok) {
    const err = (data && data.error) || res.statusText || "request failed";
    throw new Error(err);
  }
  return data;
}

function contactName(id) {
  const c = (currentTargets.contacts || []).find(c => c.id === id);
  return c ? c.name : "";
}

function relTime(iso) {
  if (!iso) return "—";
  const t = new Date(iso).getTime();
  if (!t || t < 0 || new Date(iso).getFullYear() < 2000) return "—";
  const s = Math.floor((Date.now() - t) / 1000);
  if (s < 60) return s + "s ago";
  if (s < 3600) return Math.floor(s / 60) + "m ago";
  if (s < 86400) return Math.floor(s / 3600) + "h ago";
  return Math.floor(s / 86400) + "d ago";
}

// ---------- tabs ----------

$all("nav.tabs button").forEach(btn => {
  btn.addEventListener("click", () => {
    $all("nav.tabs button").forEach(b => b.classList.remove("active"));
    $all(".view").forEach(v => v.classList.remove("active"));
    btn.classList.add("active");
    $("#view-" + btn.dataset.tab).classList.add("active");
    if (btn.dataset.tab === "status") loadStatus();
  });
});

// ---------- targets ----------

function makeReceiver(card, receiver) {
  const node = $("#tpl-receiver").content.firstElementChild.cloneNode(true);
  $(".r-id", node).value = receiver.id || "";
  $(".r-message", node).value = receiver.message || "";
  $(".r-name", node).value = contactName(receiver.id);
  $(".rid", node).textContent = receiver.id || "";

  $(".r-id", node).addEventListener("input", e => {
    $(".rid", node).textContent = e.target.value.trim();
  });
  $(".r-remove", node).addEventListener("click", () => node.remove());
  $(".r-test", node).addEventListener("click", async () => {
    const id = $(".r-id", node).value.trim();
    if (!id) { toast("Enter a LINE ID first", "error"); return; }
    try {
      await api("POST", "/api/test-notify", { id, message: $(".r-message", node).value });
      toast("Test notification sent", "ok");
    } catch (e) { toast("Failed to send: " + e.message, "error"); }
  });

  $(".receivers", card).appendChild(node);
}

function makeTarget(target) {
  const node = $("#tpl-target").content.firstElementChild.cloneNode(true);
  $(".t-enabled", node).checked = !!target.enabled;
  $(".t-name", node).value = target.name || "";
  $(".t-mac", node).value = target.mac || "";
  $(".t-mode", node).value = (target.detection && target.detection.mode) || "auto";
  $(".t-ip", node).value = (target.detection && target.detection.ip) || "";
  $(".t-message", node).value = target.message || "";

  const toggleIp = () => {
    const mode = $(".t-mode", node).value;
    $(".t-ip-wrap", node).classList.toggle("hidden", mode === "broadcast");
  };
  toggleIp();
  $(".t-mode", node).addEventListener("change", toggleIp);

  $(".t-remove", node).addEventListener("click", () => node.remove());
  $(".t-add-receiver", node).addEventListener("click", () => makeReceiver(node, { id: "", message: "" }));
  $(".t-pick", node).addEventListener("click", () => pickSeenUser(node));

  (target.receivers || []).forEach(r => makeReceiver(node, r));
  $("#targets-list").appendChild(node);
}

function renderTargets() {
  $("#targets-list").innerHTML = "";
  (currentTargets.targets || []).forEach(makeTarget);
  $("#default-message").value = currentTargets.default_message || "";
}

function collectTargets() {
  const contactsMap = new Map();
  (currentTargets.contacts || []).forEach(c => contactsMap.set(c.id, c.name));

  const targets = $all("#targets-list .target").map(card => {
    const receivers = [];
    $all(".receiver", card).forEach(rc => {
      const id = $(".r-id", rc).value.trim();
      if (!id) return;
      const name = $(".r-name", rc).value.trim();
      if (name) contactsMap.set(id, name);
      receivers.push({ id, message: $(".r-message", rc).value });
    });
    return {
      name: $(".t-name", card).value.trim(),
      mac: $(".t-mac", card).value.trim(),
      enabled: $(".t-enabled", card).checked,
      detection: { mode: $(".t-mode", card).value, ip: $(".t-ip", card).value.trim() },
      message: $(".t-message", card).value,
      receivers,
    };
  });

  const contacts = [];
  contactsMap.forEach((name, id) => { if (name) contacts.push({ id, name }); });

  return { default_message: $("#default-message").value, contacts, targets };
}

async function loadTargets() {
  try {
    currentTargets = await api("GET", "/api/targets");
    renderTargets();
  } catch (e) { toast("Failed to load targets: " + e.message, "error"); }
}

async function saveTargets() {
  try {
    currentTargets = await api("PUT", "/api/targets", collectTargets());
    renderTargets();
    toast("Saved — effective immediately", "ok");
  } catch (e) { toast("Save failed: " + e.message, "error"); }
}

$("#add-target").addEventListener("click", () => {
  makeTarget({ enabled: true, detection: { mode: "auto", ip: "" }, receivers: [] });
});
$("#save-targets").addEventListener("click", saveTargets);

// ---------- seen-user picker ----------

async function pickSeenUser(card) {
  let users;
  try { users = await api("GET", "/api/seen-users"); }
  catch (e) { toast("Failed to load recent users: " + e.message, "error"); return; }

  if (!users || users.length === 0) {
    toast("No recent messages. Ask the person to message the bot first.", "error");
    return;
  }

  const overlay = document.createElement("div");
  overlay.style.cssText = "position:fixed;inset:0;background:rgba(0,0,0,.6);display:flex;align-items:center;justify-content:center;z-index:100;";
  const box = document.createElement("div");
  box.className = "card";
  box.style.cssText = "max-width:420px;width:90%;max-height:70vh;overflow:auto;margin:0;";
  box.innerHTML = '<div class="card-head"><span class="title">Pick a receiver</span></div>';

  users.forEach(u => {
    const item = document.createElement("div");
    item.className = "receiver";
    item.style.cursor = "pointer";
    item.innerHTML =
      '<strong>' + (u.name ? escapeHtml(u.name) : "(unnamed)") + '</strong>' +
      '<div class="rid">' + escapeHtml(u.id) + '</div>' +
      '<div class="hint">' + relTime(u.lastSeen) + '</div>';
    item.addEventListener("click", () => {
      makeReceiver(card, { id: u.id, message: "" });
      const last = $(".receivers .receiver:last-child", card);
      if (last && u.name) $(".r-name", last).value = u.name;
      document.body.removeChild(overlay);
    });
    box.appendChild(item);
  });

  const cancel = document.createElement("button");
  cancel.className = "btn secondary small";
  cancel.textContent = "Cancel";
  cancel.style.marginTop = "12px";
  cancel.addEventListener("click", () => document.body.removeChild(overlay));
  box.appendChild(cancel);

  overlay.appendChild(box);
  overlay.addEventListener("click", e => { if (e.target === overlay) document.body.removeChild(overlay); });
  document.body.appendChild(overlay);
}

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}

// ---------- system ----------

async function loadSystem() {
  try {
    const s = await api("GET", "/api/system");
    $("#sys-bin").value = s.arp_scan.bin;
    $("#sys-iface").value = s.arp_scan.iface;
    $("#sys-interval").value = s.arp_scan.interval_sec;
    $("#sys-bcast").value = s.arp_scan.broadcast_timeout_sec;
    $("#sys-indiv").value = s.arp_scan.individual_timeout_sec;
    $("#sys-absence").value = s.monitor.absence_reset_min;
    $("#sys-port").value = s.server.port;
  } catch (e) { toast("Failed to load system settings: " + e.message, "error"); }
}

async function saveSystem() {
  const body = {
    arp_scan: {
      bin: $("#sys-bin").value.trim(),
      iface: $("#sys-iface").value.trim(),
      interval_sec: +$("#sys-interval").value,
      broadcast_timeout_sec: +$("#sys-bcast").value,
      individual_timeout_sec: +$("#sys-indiv").value,
    },
    monitor: { absence_reset_min: +$("#sys-absence").value },
    server: { port: +$("#sys-port").value },
  };
  try {
    await api("PUT", "/api/system", body);
    toast("System settings saved", "ok");
  } catch (e) { toast("Save failed: " + e.message, "error"); }
}

$("#save-system").addEventListener("click", saveSystem);

// ---------- status ----------

async function loadStatus() {
  try {
    const rows = await api("GET", "/api/status");
    const tbody = $("#status-rows");
    tbody.innerHTML = "";
    if (!rows || rows.length === 0) {
      $("#status-empty").classList.remove("hidden");
      return;
    }
    $("#status-empty").classList.add("hidden");
    rows.sort((a, b) => new Date(b.lastSeen) - new Date(a.lastSeen));
    rows.forEach(r => {
      const tr = document.createElement("tr");
      const badge = r.notified
        ? '<span class="badge on">Notified</span>'
        : '<span class="badge off">Pending</span>';
      tr.innerHTML =
        "<td>" + escapeHtml(r.name || "—") + "</td>" +
        '<td class="rid">' + escapeHtml(r.mac) + "</td>" +
        "<td>" + relTime(r.lastSeen) + "</td>" +
        "<td>" + badge + "</td>";
      tbody.appendChild(tr);
    });
  } catch (e) { toast("Failed to load status: " + e.message, "error"); }
}

$("#refresh-status").addEventListener("click", loadStatus);

// ---------- init ----------

loadTargets();
loadSystem();

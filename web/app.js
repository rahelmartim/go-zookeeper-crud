const SEED = [
  "CREATE /zoo",
  "CREATE /zoo/mammals",
  "CREATE /zoo/mammals/lion king of the jungle",
  "CREATE /zoo/mammals/elephant never forgets",
  "CREATE /zoo/birds",
  "CREATE /zoo/birds/parrot hello!",
  "CREATE /config",
  "CREATE /config/opening-hours 9:00-18:00",
];

const EXAMPLES = [
  "LIST /zoo",
  "READ /zoo/mammals/lion",
  "UPDATE /zoo/mammals/lion still the king",
  "CREATE /zoo/reptiles",
  "CREATE /zoo/reptiles/snake sss",
  "DELETE /zoo/birds",
  "DELETEC /zoo/birds",
  "HELP",
];

const TOUR = [
  "LIST /zoo",
  "CREATE /zoo/reptiles/snake sss",
  "CREATE /zoo/reptiles",
  "CREATE /zoo/reptiles/snake sss",
  "READ /zoo/reptiles/snake",
  "UPDATE /zoo/reptiles/snake SSSS",
  "UPDATE /zoo/reptiles/snake hiss",
  "DELETE /zoo/reptiles",
  "DELETEC /zoo/reptiles",
  "EXISTS /zoo/reptiles",
];

const $ = (selector) => document.querySelector(selector);
const output = $("#output");
const input = $("#command");
const terminalBody = $("#terminal-body");
const treeEl = $("#tree");
const tourButton = $("#tour");
const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

const history = [];
let historyIndex = 0;
let previousTree = new Map(); // path -> version
let busy = false;

function el(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, reducedMotion ? 0 : ms));

/* Terminal */

function print(text, kind) {
  output.append(el("li", kind, text));
  while (output.children.length > 300) output.firstElementChild.remove();
  terminalBody.scrollTop = terminalBody.scrollHeight;
}

async function run(line, { echo = true } = {}) {
  const text = line.trim();
  if (!text) return;
  if (echo) print(text, "entry");

  const word = text.split(/\s+/)[0].toUpperCase();
  if (word === "CLEAR") {
    output.replaceChildren();
    return;
  }
  if (word === "QUIT" || word === "EXIT") {
    print("This is the browser demo, so there is nothing to quit. In the CLI, QUIT ends the session.", "info");
    return;
  }

  const result = window.zoo.run(text);
  if (echo) {
    for (const outputLine of result.lines) print(outputLine, result.error ? "error" : "ok");
  }
  await renderTree();
}

/* Tree */

function flatten(node, map = new Map()) {
  map.set(node.path, node.version);
  for (const child of node.children) flatten(child, map);
  return map;
}

function renderNode(node, changes) {
  const item = el("li");
  const button = el("button", "znode");
  button.type = "button";
  button.dataset.path = node.path;
  if (node.path === "/") button.classList.add("root");
  if (changes.created.has(node.path)) button.classList.add("created");
  if (changes.updated.has(node.path)) button.classList.add("updated");

  button.append(
    el("span", "icon", node.path === "/" ? "/" : node.children.length > 0 ? "▾" : "•"),
    el("span", "name", node.path === "/" ? "root" : node.name),
    el("span", "value", node.value ? `"${node.value}"` : ""),
    el("span", "version", `v${node.version}`),
  );
  button.title = node.path;
  button.addEventListener("click", () => {
    if (!busy) run(`READ ${node.path}`);
  });
  item.append(button);

  if (node.children.length > 0) {
    const list = el("ul");
    for (const child of node.children) list.append(renderNode(child, changes));
    item.append(list);
  }
  return item;
}

function depth(node) {
  return node.children.length === 0 ? 0 : 1 + Math.max(...node.children.map(depth));
}

async function renderTree({ animate = true } = {}) {
  const tree = window.zoo.tree();
  const current = flatten(tree);
  const changes = { created: new Set(), updated: new Set() };
  const deleted = [];

  if (animate) {
    for (const [path, version] of current) {
      if (!previousTree.has(path)) changes.created.add(path);
      else if (previousTree.get(path) !== version) changes.updated.add(path);
    }
    for (const path of previousTree.keys()) if (!current.has(path)) deleted.push(path);
  }

  // Let deleted nodes fade out before the tree changes.
  if (deleted.length > 0) {
    for (const button of treeEl.querySelectorAll(".znode")) {
      if (deleted.includes(button.dataset.path)) button.classList.add("deleted");
    }
    await sleep(450);
  }

  treeEl.replaceChildren(renderNode(tree, changes));
  previousTree = current;

  const count = current.size - 1;
  $("#tree-summary").textContent =
    `${count} znode${count === 1 ? "" : "s"} under the root · depth ${depth(tree)}`;
}

/* Tour */

async function typeLine(line) {
  input.value = "";
  for (const char of line) {
    input.value += char;
    await sleep(28);
  }
  await sleep(250);
  input.value = "";
  await run(line);
}

async function playTour() {
  if (busy) return;
  busy = true;
  tourButton.disabled = true;
  input.disabled = true;
  print("Tour: create, read, update and delete nodes. Some commands fail on purpose, to show the rules.", "info");
  try {
    for (const line of TOUR) {
      await typeLine(line);
      await sleep(900);
    }
  } finally {
    busy = false;
    tourButton.disabled = false;
    input.disabled = false;
    input.focus();
  }
}

/* Setup */

function setupControls() {
  $("#prompt").addEventListener("submit", (event) => {
    event.preventDefault();
    const line = input.value;
    input.value = "";
    if (line.trim()) {
      history.push(line);
      historyIndex = history.length;
    }
    run(line);
  });

  input.addEventListener("keydown", (event) => {
    if (event.key === "ArrowUp" && historyIndex > 0) {
      historyIndex -= 1;
      input.value = history[historyIndex];
      event.preventDefault();
    } else if (event.key === "ArrowDown") {
      historyIndex = Math.min(history.length, historyIndex + 1);
      input.value = history[historyIndex] ?? "";
      event.preventDefault();
    }
  });

  terminalBody.addEventListener("click", () => {
    if (!window.getSelection()?.toString()) input.focus();
  });

  const chips = $("#chips");
  for (const example of EXAMPLES) {
    const chip = el("button", "chip", example);
    chip.type = "button";
    chip.addEventListener("click", () => {
      if (!busy) run(example);
    });
    chips.append(chip);
  }

  const table = $("#commands");
  for (const command of window.zoo.commands) {
    const row = el("tr");
    row.append(el("td", "", command.usage), el("td", "", command.help));
    table.append(row);
  }

  tourButton.addEventListener("click", playTour);
}

async function onReady() {
  for (const line of SEED) window.zoo.run(line);
  await renderTree({ animate: false });
  setupControls();

  input.disabled = false;
  input.placeholder = "Type a command, for example LIST /zoo";
  tourButton.disabled = false;
  print("ZooCRUD demo. The tree already has a few nodes.", "info");
  print("Type HELP to see the commands, click an example below, or play the tour.", "info");
}

function showStartError(error) {
  console.error(error);
  print("Your browser could not start the demo. Open the console for details.", "error");
  $("#tree-summary").textContent = "Not available";
}

async function start() {
  const go = new Go();
  const result = await WebAssembly.instantiateStreaming(fetch("zoo.wasm"), go.importObject);
  window.addEventListener("zoo-ready", () => onReady().catch(showStartError), { once: true });
  go.run(result.instance);
}

start().catch(showStartError);

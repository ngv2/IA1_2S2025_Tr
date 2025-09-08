const output = document.getElementById("output");

document.getElementById("fileInput").addEventListener("change", async function(event) {
  const file = event.target.files[0];
  if (!file) return;
  
  const text = await file.text();
  const res = await fetch("http://localhost:8000/load", {
    method: "POST",
    headers: { "Content-Type": "text/plain" },
    body: text
  });

  output.textContent = await res.text();
});

async function runQuery() {
  const query = document.getElementById("queryInput").value;
  const res = await fetch("http://localhost:8000/query", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ query })
  });
  const data = await res.json();
  output.textContent = data.results.map(result => JSON.stringify(result)).join("\n");
}

async function addFact() {
  const fact = document.getElementById("factInput").value;
  const res = await fetch("http://localhost:8000/add", {
    method: "POST",
    headers: { "Content-Type": "text/plain" },
    body: fact
  });
  output.textContent = await res.text();
}

function downloadCode() {
  window.open("http://localhost:8000/download", "_blank");
}
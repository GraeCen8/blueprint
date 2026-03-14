import express from "express";

import db from "./db.js";

const app = express();
app.use(express.json());

app.post("/items", (req, res) => {
  const { name } = req.body;
  if (!name) {
    return res.status(400).json({ error: "name is required" });
  }
  db.run("INSERT INTO items (name) VALUES (?)", [name], function onInsert(err) {
    if (err) {
      return res.status(500).json({ error: "db error" });
    }
    return res.json({ id: this.lastID, name });
  });
});

app.get("/items", (_req, res) => {
  db.all("SELECT id, name FROM items ORDER BY id DESC", (err, rows) => {
    if (err) {
      return res.status(500).json({ error: "db error" });
    }
    return res.json(rows);
  });
});

const port = process.env.PORT || 3000;
app.listen(port, () => {
  console.log(`listening on http://localhost:${port}`);
});

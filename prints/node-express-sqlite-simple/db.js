import sqlite3 from "sqlite3";

const db = new sqlite3.Database("data.db");

db.serialize(() => {
  db.run(
    "CREATE TABLE IF NOT EXISTS items (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)"
  );
});

export default db;

# fastapi-sqlmodel-sqlite-simple

Minimal FastAPI + SQLModel + SQLite project.

## Setup

```bash
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

## Run

```bash
uvicorn app.main:app --reload
```

## Example Requests

```bash
curl -X POST http://127.0.0.1:8000/items -H "Content-Type: application/json" -d '{"name":"example"}'
curl http://127.0.0.1:8000/items
```

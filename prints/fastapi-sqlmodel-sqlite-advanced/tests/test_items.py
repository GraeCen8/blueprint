from fastapi.testclient import TestClient

from app.main import app


def test_create_and_list_items():
    client = TestClient(app)
    response = client.post("/items", json={"name": "sample"})
    assert response.status_code == 200
    payload = response.json()
    assert payload["name"] == "sample"

    response = client.get("/items")
    assert response.status_code == 200
    items = response.json()
    assert len(items) >= 1

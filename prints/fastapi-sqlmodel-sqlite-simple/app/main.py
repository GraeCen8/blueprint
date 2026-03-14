from fastapi import Depends, FastAPI, HTTPException
from sqlmodel import Session, select

from .db import get_session, init_db
from .models import Item, ItemCreate

app = FastAPI()


@app.on_event("startup")
def on_startup() -> None:
    init_db()


@app.post("/items", response_model=Item)
def create_item(
    item: ItemCreate, session: Session = Depends(get_session)
) -> Item:
    db_item = Item(name=item.name)
    session.add(db_item)
    session.commit()
    session.refresh(db_item)
    return db_item


@app.get("/items", response_model=list[Item])
def list_items(session: Session = Depends(get_session)) -> list[Item]:
    items = session.exec(select(Item)).all()
    return items


@app.get("/items/{item_id}", response_model=Item)
def get_item(item_id: int, session: Session = Depends(get_session)) -> Item:
    item = session.get(Item, item_id)
    if not item:
        raise HTTPException(status_code=404, detail="Item not found")
    return item
